package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
)

func jqHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter from query parameter
	jqFilter := r.URL.Query().Get("filter")
	if jqFilter == "" {
		http.Error(w, "Missing 'filter' query parameter", http.StatusBadRequest)
		return
	}
	log.Println("jq filter:", jqFilter)

	// Read JSON body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, &out)
}

func main() {
	http.HandleFunc("/", jqHandler)
	port := "8080"
	log.Println("Server starting on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
