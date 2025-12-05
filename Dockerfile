# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build Lambda binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o bootstrap \
    cmd/lambda/main.go

# Runtime stage
FROM public.ecr.aws/lambda/provided:al2

# Copy binary from builder
COPY --from=builder /app/bootstrap ${LAMBDA_TASK_ROOT}/

# Set the CMD to the Lambda handler
CMD ["bootstrap"]
