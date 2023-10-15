ARG REPOSITORY

FROM --platform=$BUILDPLATFORM docker.io/debian:stable-slim as builder

ARG REPOSITORY

RUN apt-get update && apt-get install -y busybox-static ca-certificates
RUN useradd -d / -M -r -s /bin/false ${REPOSITORY}
RUN mkdir -p /e/bin && ln -sf /${REPOSITORY} /e/entrypoint
RUN busybox --install /e/bin

FROM --platform=$TARGETPLATFORM scratch as stage

ARG REPOSITORY
ARG TARGETARCH

COPY --from=builder /e /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /usr/bin/busybox /usr/bin/busybox
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY ${REPOSITORY}_linux_${TARGETARCH} /${REPOSITORY}

FROM --platform=$TARGETPLATFORM scratch

ARG REPOSITORY

LABEL org.opencontainers.image.source=https://github.com/candiddev/${REPOSITORY}
LABEL org.opencontainers.image.base.name=scratch

COPY --from=stage / /

ENTRYPOINT ["/entrypoint"]

USER ${REPOSITORY}:${REPOSITORY}
