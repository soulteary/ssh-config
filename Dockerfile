ARG ALPINE_VERSION=3.24.4
ARG ALPINE_DIGEST=sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

FROM alpine:${ALPINE_VERSION}@${ALPINE_DIGEST} AS certificates
RUN apk add --no-cache ca-certificates

FROM alpine:${ALPINE_VERSION}@${ALPINE_DIGEST}
LABEL maintainer "soulteary <soulteary@gmail.com>"
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN addgroup -S -g 65532 ssh-config \
    && adduser -S -D -u 65532 -G ssh-config -h /home/ssh-config ssh-config \
    && mkdir -p /home/ssh-config/.ssh \
    && chown -R 65532:65532 /home/ssh-config
ENV HOME=/home/ssh-config
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/ssh-config /usr/bin/ssh-config
USER 65532:65532
ENTRYPOINT ["/usr/bin/ssh-config"]
