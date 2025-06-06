ARG DEBIAN_VERSION=bookworm
ARG DEBIAN_FLAVOR=${DEBIAN_VERSION}-slim

ARG BOOKWORM_SLIM_SHA=sha256:40b107342c492725bc7aacbe93a49945445191ae364184a6d24fedb28172f6f7
ARG BULLSEYE_SLIM_SHA=sha256:e831d9a884d63734fe3dd9c491ed9a5a3d4c6a6d32c5b14f2067357c49b0b7e1

ARG BASE_IMAGE_SHA=debian@$BOOKWORM_SLIM_SHA
# ARG BASE_IMAGE_SHA=${BASE_IMAGE_SHA/debian:bookworm-slim/debian@$BOOKWORM_SLIM_SHA}
# ARG BASE_IMAGE_SHA=${BASE_IMAGE_SHA/debian:bullseye-slim/debian@$BULLSEYE_SLIM_SHA}

FROM --platform=$TARGETPLATFORM ${BASE_IMAGE_SHA}
ARG BINARY_TO_ADD

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

ADD ${BINARY_TO_ADD} /usr/local/bin/
RUN ln -s /usr/local/bin/${BINARY_TO_ADD} /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
