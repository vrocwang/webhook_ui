FROM golang:1.23-bullseye AS Builder
ENV GOPROXY=https://goproxy.cn
ENV CGO_ENABLED=0
WORKDIR /app
COPY . .
RUN go mod download -x
RUN go build -ldflags "-w -s" -o webhook .

FROM debian:stable-slim

WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates  \
    netbase \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y
COPY --from=builder /app/webhook .
CMD ["./webhook"]
