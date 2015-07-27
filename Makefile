test_stable ::
	@echo Running unit tests inside of stable
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-golang-stable godep go test -v ./...
	@echo Running climate inside of stable
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-golang-stable climate -open=false -threshold=80.0 -errcheck -vet -fmt .

test_tip ::
	@echo Running unit tests inside of tip
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-golang-tip godep go test -v ./...
	@echo Running climate inside of tip
	docker run --rm -v `pwd`:/go/src/github.com/SUSE/zypper-docker zypper-docker-testing-golang-tip climate -open=false -threshold=80.0 -errcheck -vet -fmt .

test :: test_stable test_tip

clean ::
	docker rmi zypper-docker-testing-golang-stable
	docker rmi zypper-docker-testing-golang-tip

build_stable ::
	@echo Building zypper-docker-testing-golang-stable
	docker build -f Dockerfile.golang.stable -t zypper-docker-testing-golang-stable .

build_tip ::
	@echo Building zypper-docker-testing-golang-tip
	docker build -f Dockerfile.golang.tip -t zypper-docker-testing-golang-tip .

build :: build_stable build_tip
