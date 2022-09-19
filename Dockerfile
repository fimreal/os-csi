FROM golang:1.19-alpine as gobuild

WORKDIR /build
ADD go.mod go.sum /build/
ADD cmd /build/cmd
ADD pkg /build/pkg

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./osdriver ./cmd/osdriver

# cosfs
FROM alpine:3.16 as cmdbuild
ENV COSFS_VERSION=1.0.20
ENV OSSFS_VERSION=1.80.6

RUN apk --update --no-cache add fuse alpine-sdk automake autoconf libxml2-dev fuse-dev curl-dev bash mailcap
RUN cd /tmp && wget https://github.com/tencentyun/cosfs/archive/v${COSFS_VERSION}.tar.gz &&\
    tar xf v${COSFS_VERSION}.tar.gz &&\
    cd cosfs-${COSFS_VERSION} &&\
    ./autogen.sh &&\
    ./configure --prefix=/usr &&\
    make && make install
RUN cd /tmp && wget https://github.com/aliyun/ossfs/archive/refs/tags/v${OSSFS_VERSION}.tar.gz &&\
    tar xf v${OSSFS_VERSION}.tar.gz &&\
    cd ossfs-${OSSFS_VERSION} &&\
    ./autogen.sh &&\
    ./configure --prefix=/usr &&\
    make && make install

FROM alpine:3.16

# apk add temporarily broken:
#ERROR: unable to select packages:
#  so:libcrypto.so.3 (no such package):
#    required by: s3fs-fuse-1.91-r1[so:libcrypto.so.3]
#RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing s3fs-fuse rclone

# ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /usr/bin/geesefs
# RUN chmod 755 /usr/bin/geesefs

# cosfs & ossfs runtime needed lib packages
RUN apk --no-cache add \
    ca-certificates \
    mailcap \
    fuse \
    libxml2 \
    libcurl \
    libgcc \
    libstdc++ \
    tini

COPY --from=gobuild /build/osdriver /osdriver
COPY --from=cmdbuild /usr/bin/cosfs /usr/bin/cosfs
COPY --from=cmdbuild /usr/bin/ossfs /usr/bin/ossfs
ENTRYPOINT ["/osdriver"]