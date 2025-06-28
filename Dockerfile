# -------- Stage 1: Build Go server for ARM64 --------
FROM --platform=linux/arm64 golang:1.21 AS build

WORKDIR /app
COPY main.go .

# Build a statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server main.go

# -------- Stage 2: Runtime image with jq --------
FROM --platform=linux/arm64 ubuntu:22.04

# Install jq and required dependencies
RUN apt-get update && \
    apt-get install -y jq ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy the Go server binary
COPY --from=build /app/server /server

EXPOSE 8080

CMD ["/server"]
