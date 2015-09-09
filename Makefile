test_14 ::
	@echo Running unit tests inside of Go 1.4
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-1.4 godep go test -race -v ./...
	@echo Running climate inside of Go 1.4
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-1.4 climate -open=false -threshold=80.0 -errcheck -vet -fmt .

test_15 ::
	@echo Running unit tests inside of Go 1.5
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-1.5 godep go test -race -v ./...
	@echo Running climate inside of Go 1.5
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-1.5 climate -open=false -threshold=80.0 -errcheck -vet -fmt .

test_tip ::
	@echo Running unit tests inside of tip
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-tip godep go test -race -v ./...
	@echo Running climate inside of tip
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-tip climate -open=false -threshold=80.0 -errcheck -vet -fmt .

test :: test_14 test_15 test_tip

test_integration :: build_zypper_docker build_integration_tests
	docker run \
		--rm \
		--volume="/var/run/docker.sock:/var/run/docker.sock" \
		--volume="$(CURDIR):/code" \
		zypper-docker-integration-tests \
		rake test

# Run only the RSpec tests flagged as 'quick', does NOT build the zypper-docker
# binary or the testing images
# Note well: "docker -ti" is required to use byebug inside of the ruby tests
test_integration_quick ::
	docker run \
		--rm \
		-ti \
		--volume="/var/run/docker.sock:/var/run/docker.sock" \
		--volume="$(CURDIR):/code" \
		zypper-docker-integration-tests \
		rspec -t quick

clean ::
	docker rmi zypper-docker-testing-1.4
	docker rmi zypper-docker-testing-1.5
	docker rmi zypper-docker-testing-tip
	docker rmi zypper-docker-integration-tests
	rm -f zypper-docker

build_zypper_docker :: build_tip
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-tip godep go build

build_14 ::
	@echo Building zypper-docker-testing-1.4
	docker build -f docker/Dockerfile-1.4 -t zypper-docker-testing-1.4 docker

build_15 ::
	@echo Building zypper-docker-testing-1.5
	docker build -f docker/Dockerfile-1.5 -t zypper-docker-testing-1.5 docker

build_tip ::
	@echo Building zypper-docker-testing-tip
	docker build -f docker/Dockerfile-tip -t zypper-docker-testing-tip docker

build_integration_tests ::
	@echo Building zypper-docker-integration-tests
	docker build -f docker/Dockerfile-integration-tests -t zypper-docker-integration-tests $(CURDIR)

build :: build_14 build_15 build_tip
