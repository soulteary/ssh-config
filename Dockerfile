ARG ALPINE_VERSION=3.24.1
ARG ALPINE_DIGEST=sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

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
