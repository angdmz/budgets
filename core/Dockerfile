# =============================================================================
# Base stage - shared dependencies
# =============================================================================
FROM golang:1.22-alpine AS base

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./

COPY . .

RUN go mod tidy

# =============================================================================
# Test stage - for running tests
# =============================================================================
FROM base AS test

CMD ["go", "test", "-v", "./..."]

# =============================================================================
# Builder stage - compiles the application
# =============================================================================
FROM base AS builder

# Install swag for Swagger generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs and build
RUN swag init -g main.go -o docs --parseDependency --parseInternal || true
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# =============================================================================
# Production stage - minimal runtime image
# =============================================================================
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
