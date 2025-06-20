package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)

func MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer

		buf.WriteString("\n")
		buf.WriteString("==================================================\n")
		buf.WriteString(fmt.Sprintf("[%s] %s", r.Method, r.URL.Path) + "\n")
		buf.WriteString("==================================================\n")

		contentType := r.Header.Get("Content-Type")

		// Log headers
		buf.WriteString("Headers\n")
		buf.WriteString("--------------------------------------------------\n")

		for name, values := range r.Header {
			for _, val := range values {
				buf.WriteString(fmt.Sprintf("%v: %v", name, val) + "\n")
			}
		}

		buf.WriteString("--------------------------------------------------\n")
		buf.WriteString("\n")

		buf.WriteString("Body\n")
		buf.WriteString("--------------------------------------------------\n")

		// Correctly use strings.HasPrefix
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Parse the multipart form data
			err := r.ParseMultipartForm(32 << 20) // 32MB max memory
			if err != nil {
				log.Fatalf("Error parsing multipart form: %v", err)
				http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
				return
			}

			if r.MultipartForm.Value != nil {
				for key, values := range r.MultipartForm.Value {
					for _, value := range values {
						buf.WriteString(fmt.Sprintf("Form Field: %s, Value: %s", key, value) + "\n")
					}
				}
			}

			// Iterate over the form parts (files)
			if r.MultipartForm != nil && r.MultipartForm.File != nil {
				for key := range r.MultipartForm.File {
					files := r.MultipartForm.File[key]
					for _, fileHeader := range files {
						buf.WriteString(fmt.Sprintf("Form Field: %s, FileName: %s, Content-Type: %s", key, fileHeader.Filename, fileHeader.Header.Get("Content-Type")) + "\n")
						// You can also access the file content itself if needed:
						// file, err := fileHeader.Open()
						// if err != nil { ... handle error ... }
						// defer file.Close()
						// ... read file content ...
					}
				}
			}

		} else {
			var allowedPrintBody = []string{"application/x-www-form-urlencoded",
				"application/javascript",
				"application/json",
				"application/xml",
				"text/plain",
				"text/html",
				"text/csv",
				"text/xml"}

			if slices.Contains(allowedPrintBody, contentType) {
				// Read and log body for non-multipart requests
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Panicf("Error reading body: %v", err)
					http.Error(w, "can't read body", http.StatusBadRequest)
					return
				}

				defer r.Body.Close() // Close the body

				buf.WriteString(fmt.Sprintf("%v", string(body)) + "\n")

				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

		}

		buf.WriteString("--------------------------------------------------\n")

		log.Println(buf.String())

		next.ServeHTTP(w, r)
	})
}
