package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	ENV  string = "."
	APP  string
	ARGS string
	PORT string = "4000"
)

type Reloader struct {
	application *exec.Cmd
	reloadqueue chan bool
	startcmd    string
	args        string
}

type Reloadwrapper struct {
	reloader *Reloader
}

func (R Reloadwrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	R.reloader.reloadqueue <- true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := make(map[string]string)
	response["reload"] = "attempted"
	body, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(body)
}

func (R *Reloader) reload() error {
	if R.application == nil {
		return fmt.Errorf("tried to reload non-existing application")
	}
	if R.application.Process == nil {
		return fmt.Errorf("tried to reload not-yet started process")
	}
	if err := R.application.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	gitPull()

	log.Println("Reloading Application ...")
	newInstance, err := R.startApplication()
	if err != nil {
		return nil
	}
	R.application = newInstance

	return nil
}

func (R *Reloader) startApplication() (*exec.Cmd, error) {
	cmd := exec.Command(R.startcmd, R.args)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func (R *Reloader) start() error {
	gitPull()
	newApplication, err := R.startApplication()
	R.application = newApplication
	log.Println("Starting Application ...")
	return err
}

func (R *Reloader) stopApplication() error {
	if R.application == nil {
		return fmt.Errorf("tried to stop non-existing application")
	}
	if R.application.Process == nil {
		return fmt.Errorf("tried to stop not-yet started process")
	}
	if err := R.application.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return nil

}

func newReloader(name string, args string, reloadq chan bool) Reloader {
	return Reloader{
		application: nil,
		reloadqueue: reloadq,
		startcmd:    name,
		args:        args,
	}
}

func gitPull() {
	log.Println("Pulling ...")
	gitCMD := exec.Command("git", "pull")
	gitCMD.Stderr = os.Stderr
	gitCMD.Stdout = os.Stdout
	gitCMD.Run()
}
func Flags() {
	flag.StringVar(&ENV, "e", ".", "Set the cwd for the reload-server.")
	flag.StringVar(&APP, "a", "", "The programme the reload-server watches. Arguments should be listed afterwards.")
	flag.StringVar(&PORT, "p", "4000", "The port the server listenes to.")
	flag.StringVar(&ARGS, "args", "", "Arguments for the programme in one string.")
	flag.Parse()
	PORT = strings.Join([]string{":", PORT}, "")
}
func main() {
	Flags()

	if ENV != "." {
		log.Fatal(os.Chdir(ENV))
	}
	reloadqueue := make(chan bool)
	R := newReloader(APP, ARGS, reloadqueue)
	rw := Reloadwrapper{&R}
	log.Println("Starting Reloader ...")
	err := R.start()
	if err != nil {
		log.Fatal(err)
	}
	defer R.stopApplication()

	go func() {
		for r := range reloadqueue {
			if r {
				R.reload()
			}

		}
	}()

	http.Handle("/reload", rw)
	log.Fatal(http.ListenAndServe(PORT, nil))
}
