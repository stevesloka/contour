FROM golang:1.12.1 AS build
WORKDIR /contour

ENV GOPROXY=https://gocenter.io
COPY go.mod ./
RUN go mod download

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.14.1/bin/linux/amd64/kubectl && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin

COPY integration integration
COPY cmd cmd
COPY internal internal
COPY apis apis

COPY ./integration/run.sh /contour/run.sh

RUN mkdir -p /tmp/results
