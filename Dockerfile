FROM golang:latest
ENV GOPATH=/go/

RUN apt update && apt install -y zip
RUN mkdir /go/src/ghen

WORKDIR /go/src/ghen

ADD . /go/src/ghen
