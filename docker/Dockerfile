FROM golang

RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/golang/lint/golint
RUN curl -sL -o /usr/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
RUN chmod +x /usr/bin/gimme
COPY ./test.sh /opt/test.sh
RUN chmod +x /opt/test.sh

WORKDIR /go/src/github.com/SUSE/zypper-docker
ENV GOPATH=/go
ENV GOROOT_BOOTSTRAP=/opt/go1.4
ENV GIMME_OS=linux
ENV GIMME_ARCH=amd64
ENV GO15VENDOREXPERIMENT=1

RUN useradd -m travis
USER travis
