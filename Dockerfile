FROM golang:1.25.10-alpine AS builder

WORKDIR /farm

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o farm ./cmd/farm

FROM alpine:latest

WORKDIR /farm

COPY --from=builder /farm/farm .
COPY --from=builder /farm/static ./static
COPY --from=builder /farm/htmlPages ./htmlPages

EXPOSE 443

CMD ["./farm"]