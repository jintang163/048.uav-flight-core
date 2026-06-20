FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o groundstation-server ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /root/

COPY --from=builder /app/groundstation-server .
COPY --from=builder /app/config/config.yaml ./config/

EXPOSE 8080 5760 14550/udp

CMD ["./groundstation-server"]
