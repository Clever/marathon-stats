FROM gliderlabs/alpine:3.2
ENTRYPOINT ["/bin/marathon-stats"]

WORKDIR /

COPY data /data

COPY . /go/src/github.com/Clever/marathon-stats
RUN apk-install -t build-deps go git \
    && cd /go/src/github.com/Clever/marathon-stats \
    && export GOPATH=/go \
    && go get \
    && go build -ldflags "-X main.Version=$(cat VERSION)" -o /bin/marathon-stats \
    && rm -rf /go \
    && apk del --purge build-deps
