FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY app/go.mod ./
RUN go mod download
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o img-fwd .

FROM darthsim/imgproxy:latest
COPY --from=builder /app/img-fwd /usr/local/bin/img-fwd
COPY --chmod=755 docker/entrypoint.sh /entrypoint.sh

EXPOSE 8888
ENTRYPOINT ["/entrypoint.sh"]
