FROM alpine

LABEL org.opencontainers.image.source = "https://github.com/zebbra/vmanage-exporter"
LABEL org.opencontainers.image.license = "MIT"

COPY vmanage-exporter /vmanage-exporter
ENTRYPOINT ["/vmanage-exporter"]
