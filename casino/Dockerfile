FROM golang:1.27.1 AS builder

ENV GOPRIVATE=github.com/Jordan-J-Brown/

WORKDIR /go/build
COPY go.mod go.sum ./
RUN --mount=type=secret,id=GITHUB_TOKEN,env=GITHUB_TOKEN \
  git config --global url.https://Jordan-J-Brown:$GITHUB_TOKEN@.insteadOf https://
RUN go mod download
RUN rm ~/.gitconfig

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./casino ./cmd/casino/

FROM scratch

ARG CONFIG_PATH

COPY --from=builder /go/build/casino ./casino
COPY ${CONFIG_PATH} ./config.yaml

ENTRYPOINT ["./casino"]
