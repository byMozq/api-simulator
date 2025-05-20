package main

import (
	"fmt"
	"io"
	"log" // Import the mime/multipart package
	"net/http"
	"os"
	"strings"
)

var appsLog *log.Logger

func main() {
	traficsLogFile, err1 := os.OpenFile("trafics.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	appsLogFile, err2 := os.OpenFile("apps.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	appsLog = log.New(appsLogFile, "", log.LstdFlags|log.Lshortfile)

	if err1 != nil {
		appsLog.Fatal("Error opening log file:", err1)
	}
	if err2 != nil {
		appsLog.Fatal("Error opening log file:", err2)
	}

	defer traficsLogFile.Close()
	defer appsLogFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, traficsLogFile)

	log.SetOutput(multiWriter)

	http.HandleFunc("/", handler) // Register handler for all routes

	port := "8080"
	log.Printf("Starting api-simulator on port %s", port)
	appsLog.Printf("Starting api-simulator on port %s", port)

	err3 := http.ListenAndServe("localhost:"+port, nil)
	if err3 != nil {
		appsLog.Fatalf("Could not start server: %s\n", err3)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("==================================================")
	log.Printf("[%s] %s", r.Method, r.URL.Path)
	log.Printf("==================================================")

	contentType := r.Header.Get("Content-Type")

	// Correctly use strings.HasPrefix
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse the multipart form data
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			http.Error(w, "Error parsing multipart form", http.StatusInternalServerError)
			return
		}

		// Iterate over the form parts (files)
		if r.MultipartForm != nil && r.MultipartForm.File != nil {
			for key := range r.MultipartForm.File {
				files := r.MultipartForm.File[key]
				for _, fileHeader := range files {
					log.Printf("Part Name: %s, FileName: %s, Content-Type: %s",
						key, fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
					// You can also access the file content itself if needed:
					// file, err := fileHeader.Open()
					// if err != nil { ... handle error ... }
					// defer file.Close()
					// ... read file content ...
				}
			}
		}

		if r.MultipartForm.Value != nil {
			for key, values := range r.MultipartForm.Value {
				for _, value := range values {
					log.Printf("Form Field: %s, Value: %s", key, value)
				}
			}
		}

	} else {
		// Read and log body for non-multipart requests
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Close the body
		printLog(r.Header, body)
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "success")
}

func printLog(header http.Header, body []byte) {
	// Log headers
	log.Println("Headers")
	log.Printf("--------------------------------------------------")

	for name, values := range header {
		for _, val := range values {
			log.Printf("%v: %v", name, val)
		}
	}

	log.Printf("--------------------------------------------------")
	log.Printf("")

	log.Printf("Body")
	log.Printf("--------------------------------------------------")

	log.Printf("%v", string(body))
	log.Printf("--------------------------------------------------")
}
