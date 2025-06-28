# Stage 1: Build Go binary
FROM --platform=linux/amd64 golang:1.22-alpine AS builder

WORKDIR /src
COPY main.go .
RUN go build -o app main.go

# Stage 2: Minimal runtime image
FROM --platform=linux/amd64 alpine:3.19

# Install curl for downloading jq, then clean up
RUN apk add --no-cache curl && \
    curl -L -o /usr/local/bin/jq https://github.com/jqlang/jq/releases/latest/download/jq-linux-amd64 && \
    chmod +x /usr/local/bin/jq && \
    apk del curl

WORKDIR /app
COPY --from=builder /src/app .

EXPOSE 8080
ENTRYPOINT ["./app"]
