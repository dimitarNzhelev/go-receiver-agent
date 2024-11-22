# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o receiver-agent cmd/main.go

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
    DORIS_HOST="" \
    DORIS_PORT="5439" \
    DORIS_USER="" \
    DORIS_PASSWORD="" \
    DORIS_DATABASE=""

EXPOSE 5000

CMD ["./receiver-agent"]
