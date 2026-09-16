# =============================================================================
# Base stage - shared dependencies
# =============================================================================
FROM golang:1.22-alpine AS base

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

# =============================================================================
# Test stage - for running tests
# =============================================================================
FROM base AS test

RUN apk add --no-cache gcc musl-dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install gotest.tools/gotestsum@v1.12.0

ENV CGO_ENABLED=1
COPY . .
CMD ["gotestsum", "--junitfile", "/tmp/test-results/junit.xml", "--", "-v", "-race", "-coverprofile=/tmp/test-results/coverage.out", "-covermode=atomic", "./..."]

# =============================================================================
# Builder stage - compiles the application
# =============================================================================
FROM base AS builder

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/main .

# =============================================================================
# Production stage - minimal runtime image
# =============================================================================
FROM alpine:3.20 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch AS production

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/main /main
EXPOSE 8080
ENTRYPOINT ["/main"]
