FROM gliderlabs/alpine:3.2
RUN apk-install ca-certificates
ENTRYPOINT ["/bin/marathon-stats"]
COPY ./marathon-stats /bin/marathon-stats
