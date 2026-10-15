FROM golang:1.27.1 AS builder

SHELL ["bash", "-c"]

ARG BUILD_TAGS

ENV GOCACHE=/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod
ENV GOPRIVATE=github.com/Jordan-J-Brown/

WORKDIR /go/engines/engine
COPY --from=gocache . ${GOCACHE} 
COPY --from=gomodcache . ${GOMODCACHE}
COPY go.mod go.sum /go/engines/engine
COPY --from=pkg . /go/pkg
RUN go mod download

COPY . /go/engines/engine
RUN \
  flags=(-trimpath -ldflags='-s -w'); \
  if [[ -n "${BUILD_TAGS}" ]]; then \
    flags+=("-tags=${BUILD_TAGS}"); \
  fi; \
  flags+=('-o' './app'); \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build "${flags[@]}" ./cmd/engine/

FROM scratch
ARG CONFIG_PATH
COPY --from=builder /go/engines/engine/app ./app
COPY ./config ./config
COPY ${CONFIG_PATH} ./config.yaml
ENTRYPOINT ["./app"]
