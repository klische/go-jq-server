package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jqFilter := req.QueryStringParameters["filter"]
	if jqFilter == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Missing 'filter' query parameter",
		}, nil
	}
	log.Println("jq filter:", jqFilter)

	// Require Gzip encoding for request body
	if !strings.Contains(req.Headers["content-encoding"], "gzip") {
		return events.APIGatewayProxyResponse{
			StatusCode: 415,
			Body:       "Only gzip-compressed request bodies are accepted. Please set Content-Encoding: gzip.",
		}, nil
	}

	var bodyBytes []byte
	var err error
	if req.IsBase64Encoded {
		bodyBytes, err = base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			log.Println("Error decoding base64 body:", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       "Failed to decode base64 body",
			}, nil
		}
	} else {
		bodyBytes = []byte(req.Body)
	}

	gz, err := gzip.NewReader(bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println("Error creating gzip reader:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Failed to create gzip reader",
		}, nil
	}
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		log.Println("Error reading gzip-compressed request body:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Failed to read gzip-compressed request body",
		}, nil
	}

	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		log.Println("Invalid JSON:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid JSON input",
		}, nil
	}

	inputJSON, err := json.Marshal(jsonData)
	if err != nil {
		log.Println("Error marshaling input:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to prepare JSON",
		}, nil
	}

	cmd := exec.Command("jq", jqFilter)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		log.Println("jq error output:", out.String())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "jq error: " + out.String(),
		}, nil
	}

	var gzOut bytes.Buffer
	gzWriter := gzip.NewWriter(&gzOut)
	_, err = gzWriter.Write(out.Bytes())
	if err != nil {
		log.Println("Error writing gzip-compressed response:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to gzip response",
		}, nil
	}
	gzWriter.Close()

	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"},
		IsBase64Encoded: true,
		Body:            string(gzOut.Bytes()),
	}, nil
}

func main() {
	lambda.Start(handler)
}
