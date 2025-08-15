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

		log.Println(">>> Incoming request: ", r.Method, r.URL.Path)

		var buf bytes.Buffer

		buf.WriteString("\n")
		buf.WriteString("==================================================\n")
		buf.WriteString(fmt.Sprintf("[%s] %s", r.Method, r.URL.Path) + "\n")
		buf.WriteString("==================================================\n")

		contentType := r.Header.Get("Content-Type")

		if contentType == "" {
			contentType = "[null]" // Default content type if not set
		}

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

		if r.Body == nil || r.Body == http.NoBody {
			buf.WriteString("[null]\n")

		} else {
			// check if the content type is multipart/form-data
			if strings.HasPrefix(contentType, "multipart/form-data") {
				// Parse the multipart form data
				err := r.ParseMultipartForm(32 << 20) // 32MB max memory
				if err != nil {
					log.Fatalf("Error parsing multipart form: %v", err)
					http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
					return
				}

				// Log the form values
				if r.MultipartForm.Value != nil {
					for key, values := range r.MultipartForm.Value {
						for _, value := range values {
							buf.WriteString(fmt.Sprintf("Form Field: %s, Value: %s", key, value) + "\n")
						}
					}
				}

				// Log the file uploads
				if r.MultipartForm != nil && r.MultipartForm.File != nil {
					for key := range r.MultipartForm.File {
						files := r.MultipartForm.File[key]

						for _, fileHeader := range files {
							buf.WriteString(fmt.Sprintf("Form Field: %s, FileName: %s, Content-Type: %s", key, fileHeader.Filename, fileHeader.Header.Get("Content-Type")) + "\n")
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
					// Read and log body
					body, err := io.ReadAll(r.Body)
					if err != nil {
						log.Panicf("Error reading body: %v", err)
						http.Error(w, "can't read body", http.StatusBadRequest)
						return
					}

					defer r.Body.Close() // Close the body

					r.Body = io.NopCloser(bytes.NewBuffer(body))

					detectedContentType, readBytes, err := detectContentType(&r.Body)
					if err != nil {
						http.Error(w, "Failed to read body", http.StatusInternalServerError)
						return
					}

					if slices.Contains(allowedPrintBody, detectedContentType) {
						buf.WriteString(fmt.Sprintf("%v", string(body)) + "\n")
					} else {
						buf.WriteString("Body not logged due to unsupported content type: " + detectedContentType + "\n")
					}

					// Restore the body with the read bytes included
					r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(readBytes), r.Body))

				} else {
					buf.WriteString("Body not logged due to unsupported content type: " + contentType + "\n")
				}

			}
		}

		buf.WriteString("--------------------------------------------------\n")

		log.Println(buf.String())

		next.ServeHTTP(w, r)
	})
}

// detectContentType reads a sample from the request body and detects its content type
func detectContentType(body *io.ReadCloser) (string, []byte, error) {
	buf2 := make([]byte, 512)
	n, err := (*body).Read(buf2)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	// Detect the content type
	dcontentType := http.DetectContentType(buf2[:n])

	dcontentType = strings.Split(dcontentType, ";")[0]
	dcontentType = strings.ToLower(dcontentType)

	return dcontentType, buf2[:n], nil
}
