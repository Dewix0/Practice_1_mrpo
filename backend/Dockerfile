FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server
RUN CGO_ENABLED=0 go build -o seed ./cmd/seed

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh && mkdir -p /app/data /app/uploads

ENV DB_PATH=/app/data/data.db
ENV UPLOADS_DIR=/app/uploads

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
