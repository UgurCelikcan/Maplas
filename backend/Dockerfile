FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./ # go.sum might not exist yet if only standard lib used, but good practice
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
