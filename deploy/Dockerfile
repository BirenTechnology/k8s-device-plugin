FROM golang:1.20 AS builder

ARG build_arch

COPY . /src
WORKDIR /src

RUN if [ "$build_arch" = "arm64" ]; then  \
      make build-arm ; \
    else  \
      make build ; \
    fi


FROM ubuntu:22.04

COPY --from=builder /src/k8s-device-plugin /root/k8s-device-plugin

WORKDIR /root/
