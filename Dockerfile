# Stage 1: Build Go binary
FROM --platform=linux/amd64 golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod go.sum main.go .

# 3. Download dependencies (includes brotli)
RUN go mod download

# Build a statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

# Stage 2: Minimal runtime image
FROM --platform=linux/amd64 alpine:3.19

# Install curl for downloading jq, then clean up
RUN apk add --no-cache curl && \
    curl -L -o /usr/local/bin/jq https://github.com/jqlang/jq/releases/latest/download/jq-linux-amd64 && \
    chmod +x /usr/local/bin/jq && \
    apk del curl

# 7. Copy the built binary from the build stage
COPY --from=build /app/server /server

# 9. Expose the port your app runs on (optional)
EXPOSE 8080

# 10. Run your Go binary
CMD ["/server"]
