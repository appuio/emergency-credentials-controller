FROM docker.io/library/alpine:3.18 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["emergency-credentials-controller"]
COPY emergency-credentials-controller /usr/bin/

USER 65536:0
