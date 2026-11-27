FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

FROM alpine:latest
# Run as an unprivileged user instead of root
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /app/main .
RUN mkdir -p uploads && chown app:app uploads
USER app
EXPOSE 8080
CMD ["./main"]
