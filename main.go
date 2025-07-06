package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func jqHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter from query parameter
	jqFilter := r.URL.Query().Get("filter")
	if jqFilter == "" {
		http.Error(w, "Missing 'filter' query parameter", http.StatusBadRequest)
		return
	}
	log.Println("jq filter:", jqFilter)

	// Require Gzip encoding for request body
	if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		http.Error(w, "Only gzip-compressed request bodies are accepted. Please set Content-Encoding: gzip.", http.StatusUnsupportedMediaType)
		return
	}

	gz, gzErr := gzip.NewReader(r.Body)
	if gzErr != nil {
		log.Println("Error creating gzip reader:", gzErr)
		http.Error(w, "Failed to create gzip reader", http.StatusBadRequest)
		return
	}
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		log.Println("Error reading gzip-compressed request body:", err)
		http.Error(w, "Failed to read gzip-compressed request body", http.StatusBadRequest)
		return
	}

	// Validate and reformat input JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		log.Println("Invalid JSON:", err)
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	inputJSON, err := json.Marshal(jsonData)
	if err != nil {
		log.Println("Error marshaling input:", err)
		http.Error(w, "Failed to prepare JSON", http.StatusInternalServerError)
		return
	}

	// Run jq
	cmd := exec.Command("jq", jqFilter)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		log.Println("jq error output:", out.String())
		http.Error(w, "jq error: "+out.String(), http.StatusInternalServerError)
		return
	}

	// Always respond Gzip-compressed
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(http.StatusOK)
	gzWriter := gzip.NewWriter(w)
	_, err = gzWriter.Write(out.Bytes())
	if err != nil {
		log.Println("Error writing gzip-compressed response:", err)
	}
	gzWriter.Close()
}

func main() {
	http.HandleFunc("/", jqHandler)
	port := "8080"
	log.Println("Server starting on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
