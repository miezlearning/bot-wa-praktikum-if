FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code & build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot_ascii ./cmd/bot

# Runtime Stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bot_ascii .

# Folder untuk menyimpan database sesi WhatsApp permanen
RUN mkdir -p /data
ENV DB_PATH=/data/whatsapp.db
ENV SERVER_PORT=8080

EXPOSE 8080
CMD ["./bot_ascii"]
