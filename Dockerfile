# NOTE: pin to a digest for supply-chain integrity. Replace the tag below
# with the form `ubuntu:noble-20250127@sha256:<digest>` after running:
#   docker buildx imagetools inspect ubuntu:noble-20250127
# Dependabot's docker ecosystem (.github/dependabot.yml) keeps this updated.
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
    cp -v /tmp/appsvr/appsvrd-linux-${ARCH} /usr/local/bin/appsvrd && \
    chmod 0755 /usr/local/bin/appsvrd && \
    chown root:root /usr/local/bin/appsvrd && \
    rm -rf /tmp/appsvr && \
    groupadd --system --gid 10001 appsvr && \
    useradd --system --uid 10001 --gid 10001 --shell /usr/sbin/nologin \
            --no-create-home --comment "appsvrd service account" appsvr

USER 10001:10001

CMD ["appsvrd"]
