FROM golang:1.24.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go test -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o hetzner-ddns .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/hetzner-ddns .

# Copy the config directory from builder
COPY --from=builder /config /config

USER nonroot:nonroot
ENTRYPOINT ["/app/hetzner-ddns"]
