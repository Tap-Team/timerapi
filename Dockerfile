FROM golang:1.20-alpine3.17 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/timer /app/cmd/main/main.go

FROM alpine:3.17
COPY --from=builder /app/config /config
COPY --from=builder /app/timer /timer
EXPOSE 12700
ENTRYPOINT [ "/timer" ]