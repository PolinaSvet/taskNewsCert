FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o service-censor

FROM golang:1.22
COPY --from=builder /app/service-censor /service-censor
COPY --from=builder /app/configKafka.json .
COPY --from=builder /app/configOffensive.json .
COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/service-censor"]
