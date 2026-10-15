# This Dockerfile is purely for local development purposes, allowing replacement of the "pkg" module with a local module.

FROM golang:1.27.1 AS builder

SHELL ["bash", "-c"]

ARG BUILD_TAGS

ENV GOCACHE=/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod
ENV GOPRIVATE=github.com/Jordan-J-Brown/

# Sets up build modules (--build-context flags need to be used to specify the caches and pkg module).
# Note to future self: copying only the go.mod and go.sum files caches these layers, so we do not have to download the modules again if only the casino code changes.
WORKDIR /go/casino
COPY --from=gocache . ${GOCACHE} 
COPY --from=gomodcache . ${GOMODCACHE}
COPY go.mod go.sum /go/casino
COPY --from=pkg . /go/pkg
RUN go mod download

# Build application
COPY . /go/casino
RUN \
  flags=(-trimpath -ldflags='-s -w'); \
  if [[ -n "${BUILD_TAGS}" ]]; then \
    flags+=("-tags=${BUILD_TAGS}"); \
  fi; \
  flags+=('-o' './app'); \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build "${flags[@]}" ./cmd/casino/

# Application entrypoint
FROM scratch
ARG CONFIG_PATH
COPY --from=builder /go/casino/app ./app
COPY ${CONFIG_PATH} ./config.yaml
ENTRYPOINT ["./app"]
