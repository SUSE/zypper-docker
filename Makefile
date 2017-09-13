test :: checks unit_test test_integration

unit_test :: build
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker /opt/test.sh

test_integration :: build_zypper_docker build_integration_tests
#	docker run \
#		--rm \
#		--volume="/var/run/docker.sock:/var/run/docker.sock" \
#		--volume="$(CURDIR):/code" \
#		zypper-docker-integration-tests \
#		rake test

# Run only the RSpec tests flagged as 'quick', does NOT build the zypper-docker
# binary or the testing images
# Note well: "docker -ti" is required to use byebug inside of the ruby tests
test_integration_quick ::
#	docker run \
#		--rm \
#		-ti \
#		--volume="/var/run/docker.sock:/var/run/docker.sock" \
#		--volume="$(CURDIR):/code" \
#		zypper-docker-integration-tests \
#		rspec -t quick

checks :: vet fmt lint gotest climate

vet ::
	@echo "+ $@"
		@go vet .

fmt ::
	@echo "+ $@"
		@test -z "$$(gofmt -l . | grep -v vendor | tee /dev/stderr)" || \
					echo "+ please format Go code with 'gofmt'"

lint ::
	@echo "+ $@"
		@test -z "$$(golint . | grep -v vendor | tee /dev/stderr)"

climate:
	@echo "+ $@"
		@(./scripts/climate -o -p -a .)

# TODO: use '-race' once it works in openSUSE's golang
gotest:
	@echo "+ $@"
		@go test

clean ::
	docker rmi zypper-docker
	docker rmi zypper-docker-integration-tests
	rm -f zypper-docker
	rm -f man/man1

man ::
	@ cd man && go run generate.go

build ::
	@echo Building zypper-docker
	docker build -f docker/Dockerfile -t zypper-docker docker

build_zypper_docker ::
	go build

build_integration_tests ::
	@echo Building zypper-docker-integration-tests
	docker build -f docker/Dockerfile-integration-tests -t zypper-docker-integration-tests $(CURDIR)

help ::
	@echo usage: make [target]
	@echo
	@echo build: Creates the dockerimage.
	@echo clean: Remove the dockerimage.
	@echo test: Testing zypper-docker.
	@echo test_integration: Integration Tests
	@echo man: Generate man pages of zypper-docker
