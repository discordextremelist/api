FROM golang:1.23-alpine AS builder
RUN apk add git
WORKDIR /app
COPY . .
RUN go get
RUN go build

FROM alpine:3.12
WORKDIR /app
COPY --from=builder /app/api /app/
CMD ["./api"]