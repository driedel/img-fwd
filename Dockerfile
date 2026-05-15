FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY app/go.mod ./
RUN go mod download
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o img-fwd .

FROM darthsim/imgproxy:latest
COPY --from=builder /app/img-fwd /usr/local/bin/img-fwd
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8888
ENTRYPOINT ["/entrypoint.sh"]
