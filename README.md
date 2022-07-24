# Git-hot-reload
A tiny server to reload any application on a simple http request. Best used with git(hub) hooks.

## Arguments
+ -a="/path/to/executable"  
   + Path to the programme executed by the reload server. E.g "python3", "go build", ...
+ -args="<arguments for the executable>"
   + __In one string__
+ -e="/current/working/directory"
   + Set the directory observed by the server. Should exist and inhabit a git-repo, otherwise the reload-server failes.
+ -p="4000"
   + Set the port which then needs to be exposed.
