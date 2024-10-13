FROM golang:1.23.2-bullseye

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update && \
    apt-get install -y postgresql ca-certificates && \
    mkdir /opt/src

WORKDIR /opt/src
ADD . /opt/src

RUN go build -o /usr/bin/server ./cmd/main.go

CMD ["server"]