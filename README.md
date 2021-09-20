# ports

##Usage:
To get the next available host ports on localhost docker.

Often docker compose files will be broken down into multiple files for different services.
When developing a new service I had to go through multiple files and look for the next available port to use.
This CLI application will receive a path and find all docker compose files in that path. It will find all the available ports and print them to stdout.

It can also filter ports based on `--port` flag. for example:

if you are cloning this repo:

`go run main.go grab --directory ${your_directory} --port=8xxx`

if you are using the binary:

`./ports grab --directory ${your_directory} --port 80xx`

above command will return the next set of available ports matching 8xxx pattern ( e.g. 8901)

`--port` is optional and `--directory` will default to `"."`