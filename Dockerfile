FROM golang:1.18 as builder
ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY . .

RUN go mod download github.com/stretchr/testify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-notification-payment

FROM scratch
WORKDIR /app
COPY --from=builder /app/api-notification-payment /app/
CMD ["/app/api-notification-payment"]
