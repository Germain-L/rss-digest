FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rss-digest .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/rss-digest .
COPY --from=builder /app/frontend ./frontend

RUN mkdir -p /data

EXPOSE 8080

ENV PORT=8080
ENV DATA_DIR=/data

ENTRYPOINT ["./rss-digest"]
CMD ["--serve"]
