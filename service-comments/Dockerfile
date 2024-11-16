FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o service-comments

FROM golang:1.22
COPY --from=builder /app/service-comments /service-comments
COPY --from=builder /app/configKafka.json .
COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/service-comments"]
