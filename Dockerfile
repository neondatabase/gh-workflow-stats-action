ARG	GO_VER=1.23.1
ARG DEBIAN_VER=bookworm

FROM golang:${GO_VER}-${DEBIAN_VER} AS go-build

ENV USER=gh-action
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ pkg/
COPY cmd/gh-action/ ./

ENV CGO_ENABLED=0
RUN go build -v -o ./gh-action-workflow-stats


FROM scratch

# Setup SSL certs
COPY --from=go-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Setup user
COPY --from=go-build /etc/passwd /etc/group /etc/
USER gh-action:gh-action

# Copy the static executable
COPY --from=go-build /build/gh-action-workflow-stats /gh-action-workflow-stats

# Run the binary
ENTRYPOINT ["/gh-action-workflow-stats"]


