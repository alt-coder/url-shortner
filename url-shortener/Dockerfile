FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=1

RUN go build -o /opt/bin/url-shortener url-shortener/app/app.go

FROM busybox:latest

COPY --from=builder /opt/bin/url-shortener /opt/bin/url-shortener

EXPOSE 8080

CMD ["/opt/bin/url-shortener"]