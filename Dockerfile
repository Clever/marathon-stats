FROM gliderlabs/alpine:3.2
ENTRYPOINT ["/bin/marathon-stats"]

WORKDIR /

COPY data /data
COPY ./marathon-stats /bin/marathon-stats
