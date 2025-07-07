# ----------- Build stage -----------
FROM golang:1.24.4 AS builder

# Set working directory
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy the application code
COPY main.go ./

# Build the binary for Lambda's custom runtime
# name it "bootstrap" as required by Lambda
RUN CGO_ENABLED=0 GOOS=linux go build -o bootstrap main.go

# ----------- Final stage (Lambda) -----------
FROM public.ecr.aws/lambda/provided:al2

# Copy the built Go binary
COPY --from=builder /app/bootstrap /var/task/bootstrap

# Install jq (latest) in the final Lambda container
RUN curl -L -o /usr/local/bin/jq https://github.com/stedolan/jq/releases/download/jq-1.8.1/jq-linux64 \
    && chmod +x /usr/local/bin/jq

# (Optional) set an environment variable to help with debugging
ENV JQ_PATH=/usr/local/bin/jq

# Lambda will run /var/task/bootstrap automatically
ENTRYPOINT ["/var/task/bootstrap"]
