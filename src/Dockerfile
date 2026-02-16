# ---- Build Stage ----
FROM --platform=linux/amd64 golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

# 强制交叉编译成 amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

# ---- Runtime Stage ----
FROM --platform=linux/amd64 alpine:latest

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]