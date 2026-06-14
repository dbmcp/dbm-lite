FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dbm-lite ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/dbm-lite .
COPY .env.example ./.env
RUN mkdir -p /data
EXPOSE 8080
CMD ["./dbm-lite"]
