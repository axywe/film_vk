FROM golang:1.17 as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o filmotheka ./cmd/filmotheka/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/filmotheka .
COPY .env .

CMD ["./filmotheka"]
