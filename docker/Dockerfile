FROM golang

RUN go get github.com/mssola/climate
RUN go get golang.org/x/tools/cmd/cover
RUN go get golang.org/x/tools/cmd/vet
RUN go get github.com/golang/lint/golint
RUN go get github.com/kisielk/errcheck
RUN go get github.com/tools/godep
RUN curl -sL -o /usr/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
RUN chmod +x /usr/bin/gimme
ADD ./test.sh /opt/test.sh
RUN chmod +x /opt/test.sh

WORKDIR /go/src/github.com/SUSE/zypper-docker
ENV GOPATH=/go:/go/src/github.com/SUSE/zypper-docker/Godeps/_workspace/
ENV GOROOT_BOOTSTRAP=/opt/go1.4
ENV GIMME_OS=linux
ENV GIMME_ARCH=amd64

RUN useradd -m travis
USER travis

