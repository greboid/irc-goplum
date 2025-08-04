FROM golang:1.24 as builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -trimpath -ldflags=-buildid= -o main ./cmd/goplum

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0

COPY --from=builder /app/main /irc-goplum
CMD ["/irc-goplum"]
