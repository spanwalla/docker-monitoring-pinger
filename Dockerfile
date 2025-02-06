# Step 1: Modules caching
FROM golang:1.23-alpine3.21 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.23-alpine3.21 AS builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/docker-pinger github.com/spanwalla/docker-monitoring-pinger

# Step 3: Final
FROM scratch
COPY --from=builder /app/config /config
COPY --from=builder /bin/docker-pinger /docker-pinger
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/docker-pinger", "--config=config/config.yaml"]