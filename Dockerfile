FROM alpine:3.24.1 AS builder
RUN apk add --no-cache ca-certificates

FROM alpine:3.24.1
RUN apk add --no-cache bash
LABEL maintainer "soulteary <soulteary@gmail.com>"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/ssh-config /usr/bin/ssh-config
SHELL ["/bin/bash", "-c"]
CMD ["/usr/bin/ssh-config"]
