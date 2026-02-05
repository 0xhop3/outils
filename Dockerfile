FROM golang:1.25.4

WORKDIR /app

COPY . .

RUN go build -o server ./cmd/api

EXPOSE 8080

CMD ["./server"]
