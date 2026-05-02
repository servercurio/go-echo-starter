FROM ubuntu:noble-20250127

COPY ./bin/ /tmp/crusnik/

RUN mkdir -p /tmp/crusnik && \
    ls -lah /tmp/crusnik && \
    ARCH="$(dpkg --print-architecture)" && \
    case "$ARCH" in \
        x86_64|amd64) ARCH="amd64" ;; \
        aarch64|arm64) ARCH="arm64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    cp -v /tmp/crusnik/crusnikd-linux-${ARCH} /usr/local/bin/crusnikd && \
    chmod +x /usr/local/bin/crusnikd && \
    rm -rf /tmp/crusnik

CMD ["crusnikd"]