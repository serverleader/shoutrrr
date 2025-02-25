FROM alpine:3.21.3 AS alpine

RUN apk add --no-cache ca-certificates

FROM scratch

COPY --from=alpine \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt
COPY shoutrrr/shoutrrr /

ENTRYPOINT ["/shoutrrr"]
