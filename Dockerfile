FROM alpine:3.10
RUN apk add ca-certificates && update-ca-certificates
ENTRYPOINT ["/bin/marathon-stats"]
COPY ./marathon-stats /bin/marathon-stats
