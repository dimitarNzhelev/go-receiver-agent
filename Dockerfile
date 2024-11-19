# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o receiver-agent

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/receiver-agent .

# Create a non-root user
RUN adduser -D appuser
USER appuser

ENV PORT=5000 \
    AUTH_TOKEN="" \
    REDSHIFT_HOST="" \
    REDSHIFT_PORT="5439" \
    REDSHIFT_USER="" \
    REDSHIFT_PASSWORD="" \
    REDSHIFT_DATABASE="" \
    REDSHIFT_SSLMODE="disable"

EXPOSE 5000

CMD ["./receiver-agent"]
