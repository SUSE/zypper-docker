VERSIONS = 1.4 1.5 tip

test ::
		for version in $(VERSIONS);do \
			echo Running unit tests inside of Go $${version} ...; \
			docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker /bin/bash -c "/opt/test.sh -v=$${version}"; \
		done

test_integration :: build_zypper_docker build_integration_tests
		docker run \
						--rm \
						--volume="/var/run/docker.sock:/var/run/docker.sock" \
						--volume="$(CURDIR):/code" \
						zypper-docker-integration-tests \
						rake test

test_integration :: build_zypper_docker build_integration_tests
	docker run \
		--rm \
		--volume="/var/run/docker.sock:/var/run/docker.sock" \
		--volume="$(CURDIR):/code" \
		zypper-docker-integration-tests \
		rake test

clean ::
		docker rmi zypper-docker
		docker rmi zypper-docker-integration-tests
		rm -f zypper-docker

build ::
		@echo Building zypper-docker
		docker build -f docker/Dockerfile -t zypper-docker docker

build_zypper_docker :: build
		docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker godep go build

build_integration_tests ::
		@echo Building zypper-docker-integration-tests
		docker build -f docker/Dockerfile-integration-tests -t zypper-docker-integration-tests $(CURDIR)

help ::
		@echo usage: make [target]
		@echo
		@echo build: Creates the dockerimage.
		@echo clean: Remove the dockerimage.
		@echo test: Testing zypper-docker.
