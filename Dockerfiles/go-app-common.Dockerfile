ARG BASE_IMAGE=debian:bookworm-slim

FROM --platform=$TARGETPLATFORM ${BASE_IMAGE}
ARG BINARY_TO_ADD

ADD ${BINARY_TO_ADD} /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/${BINARY_TO_ADD}"]
