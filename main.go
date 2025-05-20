package main

import (
	"encoding/json"
	"io"
	"log" // Import the mime/multipart package
	"net/http"
	"os"
)

var appsLog *log.Logger

func main() {
	log.Printf("Starting api-simulator..")

	// Set up log
	// Ensure log directory exists
	if err := os.MkdirAll("log", 0755); err != nil {
		log.Fatal("Error creating log directory:", err)
	}

	traficsLogFile, err1 := os.OpenFile("log/trafics.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	appsLogFile, err2 := os.OpenFile("log/apps.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

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

	// Set up apps
	mux := http.DefaultServeMux

	mux.HandleFunc("/", handler) // Register handler for all routes

	port := "8080"
	log.Printf("1 api-simulator started on port %s", port)
	appsLog.Printf("2 api-simulator started on port %s", port)

	var handler http.Handler = mux
	handler = MiddlewareLog(handler)

	server := new(http.Server)
	server.Addr = "localhost:" + port
	server.Handler = handler

	err3 := server.ListenAndServe()
	if err3 != nil {
		appsLog.Fatalf("Could not start server: %s\n", err3)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	method := r.Method

	log.Println("Request receiveds:", method, url)

	var result []byte
	var err error

	if method == "GET" && url == "/status" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var unauthMsg = map[string]string{"message": "oks"}
		result, err = json.Marshal(unauthMsg)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(result)
		return
	}

}
