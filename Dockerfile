FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /conductor ./cmd/conductor \
    && mkdir -p /data

FROM gcr.io/distroless/static:nonroot
COPY --from=builder --chown=nonroot:nonroot /conductor /conductor
COPY --from=builder --chown=65532:65532 /data /data
VOLUME /data
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/conductor", "healthz"]
USER nonroot:nonroot
ENTRYPOINT ["/conductor"]
