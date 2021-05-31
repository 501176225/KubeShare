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
    make kubeshare-device-manager

FROM alpine:3.9

COPY --from=build /go/src/kubeshare/bin/kubeshare-device-manager /usr/bin/kubeshare-device-manager

CMD ["kubeshare-device-manager", "-alsologtostderr", "-v=4"]