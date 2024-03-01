FROM golang:1.22-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY . .
RUN go build -o myapp

FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/myapp .
ENV PORT=8080
CMD ["./myapp"]
