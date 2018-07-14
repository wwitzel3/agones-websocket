REPOSITORY = gcr.io/YOUR_PROJECT_ID

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
project_path := $(dir $(mkfile_path))
server_tag = $(REPOSITORY)/simple-ws
launcher_tag = $(REPOSITORY)/simple-launcher
package = github.com/wwitzel3/agones-simple

# build both launcher and server
build: build-launcher build-server
build-image: build-server-image build-launcher-image
push-image: push-server push-launcher

# Build the launcher 
build-launcher:
	CGO_ENABLED=0 go build -o $(project_path)/launcher/bin/launcher -a -installsuffix cgo $(package)/launcher

# Build the server
build-server:
	CGO_ENABLED=0 go build -o $(project_path)/server/bin/server -a -installsuffix cgo $(package)/server

# Build a docker image for the server, and tag it
build-server-image:
	docker build $(project_path)/server/ --tag=$(server_tag)

# Build a docker image for the launcher, and tag it
build-launcher-image:
	docker build $(project_path)/launcher/ --tag=$(launcher_tag)

push-server:
	docker push $(server_tag)

push-launcher:
	docker push $(launcher_tag)

deploy-server: build-server build-server-image push-server
deploy-launcher: build-launcher build-launcher-image push-launcher
deploy: deploy-launcher deploy-server 
