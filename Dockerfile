FROM golang:1.11.13-alpine3.10

LABEL maintainer="Manala <contact@manala.io>"

##########
# System #
##########

# The 'container' environment variable tells systemd that it's running inside a
# Docker container environment.
# It's also internally used for checking we're running inside a container too.
ENV container docker

RUN \
    apk --no-cache add \
        make \
        git \
        gcc \
        musl-dev
