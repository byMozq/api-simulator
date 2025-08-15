package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appsLog     *log.Logger
	apiDataList []APIData
)

func main() {
	// Set up log
	// Ensure log directory exists
	if err := os.MkdirAll("log", 0755); err != nil {
		log.Fatal("Error creating log directory:", err)
	}

	// appsLogFile, err2 := os.OpenFile("log/apps.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// traficsLogFile, err1 := os.OpenFile("log/trafics.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	appsLogFile := &lumberjack.Logger{
		Filename:   "log/apps2.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     1,     //days
		Compress:   false, // disabled by default
		LocalTime:  true,  // use local time for timestamps
	}

	traficsLogFile := &lumberjack.Logger{
		Filename:   "log/trafics2.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     1,     //days
		Compress:   false, // disabled by default
		LocalTime:  true,  // use local time for timestamps
	}

	// application logging (appsLog)
	appsLog = log.New(appsLogFile, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	// trafic logging (log)
	multiWriter := io.MultiWriter(os.Stdout, traficsLogFile)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(multiWriter)

	defer traficsLogFile.Close()
	defer appsLogFile.Close()

	log.Println("Starting api-simulator..")

	var err error

	// Read and print API data from JSON file
	apiDataList, err = readAPIDataJson()

	if err != nil {
		log.Printf("Error reading API data: %v", err)
		appsLog.Printf("Error reading API data: %v", err)
	}

	// Set up apps
	mux := http.DefaultServeMux

	mux.HandleFunc("/", handler) // Register handler for all routes

	port := "8080"

	var handler http.Handler = mux
	handler = MiddlewareLog(handler)

	server := new(http.Server)
	server.Addr = "localhost:" + port
	server.Handler = handler

	log.Printf("api-simulator started on port %s", port)
	appsLog.Printf("api-simulator started on port %s", port)

	err = server.ListenAndServe()

	if err != nil {
		appsLog.Fatalf("Could not start api-simulator: %s\n", err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	method := r.Method

	index := slices.IndexFunc(apiDataList, func(ad APIData) bool {
		return ad.Method == method && ad.URL == url
	})

	var targetApi APIData

	if index != -1 {
		log.Printf("API Found! index: %d", index)

		targetApi = apiDataList[index]

		response := targetApi.Response

		for key, val := range response.Headers {
			w.Header().Set(key, val)
		}
		w.WriteHeader(response.StatusCode)

		if response.Body.Body != "" {
			w.Write([]byte(response.Body.Body))
		}

		return

	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		var errMsg = map[string]string{"message": "request not found"}
		var result, err = json.Marshal(errMsg)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(result)
		return
	}

}

// readAPIDataJson reads data from tmp/api-data.json
func readAPIDataJson() ([]APIData, error) {
	// Open the JSON file
	file, err := os.Open("tmp/api-data.json")
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var apiDataList []APIData

	// Parse JSON data
	err = json.Unmarshal(data, &apiDataList)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return apiDataList, nil
}
