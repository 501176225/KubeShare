# call from repo root

FROM ubuntu:18.04 AS build

ENV GOLANG_VERSION 1.13.5
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/kubeshare

COPY . .

RUN  sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN  apt-get clean

RUN apt update && \
    apt install -y g++ wget make

RUN wget -nv -O - https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz | tar -C /usr/local -xz

RUN export GO111MODULE=on && \
    export GOPROXY=https://goproxy.cn && \
    make kubeshare-config-client

FROM debian:stretch-slim

ENV NVIDIA_VISIBLE_DEVICES      all
ENV NVIDIA_DRIVER_CAPABILITIES  utility

COPY --from=build /go/src/kubeshare/bin/kubeshare-config-client /usr/bin/kubeshare-config-client

CMD ["kubeshare-config-client", "-alsologtostderr", "-server-ip=kubeshare-device-manager.kube-system.svc:9797", "-v=4"]
