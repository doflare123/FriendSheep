FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o notify_service .

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates \
    && addgroup -S notify \
    && adduser -S notify -G notify

COPY --from=builder /app/notify_service .
USER notify

EXPOSE 8080

CMD ["./notify_service"]
