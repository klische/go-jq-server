package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/andybalholm/brotli"
)

func jqHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter from query parameter
	jqFilter := r.URL.Query().Get("filter")
	if jqFilter == "" {
		http.Error(w, "Missing 'filter' query parameter", http.StatusBadRequest)
		return
	}
	log.Println("jq filter:", jqFilter)

	// Handle optional Brotli encoding for request body
	var body []byte
	var err error
	if strings.Contains(r.Header.Get("Content-Encoding"), "br") {
		br := brotli.NewReader(r.Body)
		body, err = io.ReadAll(br)
		if err != nil {
			log.Println("Error reading brotli-compressed request body:", err)
			http.Error(w, "Failed to read brotli-compressed request body", http.StatusBadRequest)
			return
		}
	} else {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
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

	// If request was Brotli, respond Brotli-compressed
	if strings.Contains(r.Header.Get("Content-Encoding"), "br") {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "br")
		w.WriteHeader(http.StatusOK)
		brWriter := brotli.NewWriterLevel(w, brotli.BestCompression)
		_, err := brWriter.Write(out.Bytes())
		if err != nil {
			log.Println("Error writing brotli-compressed response:", err)
		}
		brWriter.Close()
	} else {
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, &out)
	}
}

func main() {
	http.HandleFunc("/", jqHandler)
	port := "8080"
	log.Println("Server starting on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
