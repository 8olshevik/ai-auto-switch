# Build Go backend
FROM golang:1.24-alpine AS backend-builder

# Install Go dependencies
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy all source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY services/ ./services/
COPY main.go .
COPY version_service.go .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o codeswitch ./cmd/server/main.go

# Stage 3: Alpine runtime image
FROM alpine:3.20 AS runtime

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from backend builder
COPY --from=backend-builder /app/codeswitch .

# Copy pre-built frontend static files
COPY frontend/dist ./frontend/dist

# Expose ports
# 8080: Web UI + API
# 18100: Proxy service
EXPOSE 8080 18100

# Run the application
CMD ["./codeswitch"]