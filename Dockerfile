FROM golang:1.23.1-alpine3.20 AS builder
WORKDIR /app
COPY src/ ./
RUN go mod download
COPY . .
RUN go build -o main .
FROM scratch
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8086
CMD ["./main"]
