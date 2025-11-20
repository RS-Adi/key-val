# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
# Build the binary (static linking)
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy-bin cmd/proxy/main.go

# Stage 2: Run (Small Image)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/proxy-bin .
EXPOSE 8080
CMD ["./server"]
