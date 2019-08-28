FROM golang:1.12
LABEL maintainer="Nirmoy Das <nirmoy.aiemd@gmail.com>"
WORKDIR $GOPATH/src/k8s.io/sample-controller
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
ENTRYPOINT ["sample-controller"]
