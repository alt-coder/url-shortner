FROM golang:latest AS builder

WORKDIR /app

COPY . .
ENV CGO_ENABLED=0

RUN  go build -o /opt/bin/url-shortener url-shortener/app/app.go

FROM alpine:latest

COPY --from=builder /opt/bin/url-shortener /opt/bin/url-shortener

EXPOSE 8080

CMD ["/opt/bin/url-shortener"]