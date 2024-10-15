FROM alpine:3.20.0 as builder
RUN apk --update add ca-certificates

FROM alpine:3.20.0
RUN apk --update add bash
LABEL maintainer "soulteary <soulteary@gmail.com>"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ssh-config /usr/bin/ssh-config
SHELL ["/bin/bash", "-c"]
CMD ["/usr/bin/ssh-config"]