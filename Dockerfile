FROM ubuntu:noble-20250127

COPY ./bin/ /tmp/appsvr/

RUN mkdir -p /tmp/appsvr && \
    ls -lah /tmp/appsvr && \
    ARCH="$(dpkg --print-architecture)" && \
    case "$ARCH" in \
        x86_64|amd64) ARCH="amd64" ;; \
        aarch64|arm64) ARCH="arm64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    cp -v /tmp/crusnik/appsvrd-linux-${ARCH} /usr/local/bin/appsvrd && \
    chmod +x /usr/local/bin/appsvrd && \
    rm -rf /tmp/appsvr

CMD ["appsvrd"]