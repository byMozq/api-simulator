package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appsLog *log.Logger
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

	err := server.ListenAndServe()

	if err != nil {
		appsLog.Fatalf("Could not start api-simulator: %s\n", err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	method := r.Method

	log.Println(">>> Incoming request:", method, url)

	var result []byte
	var err error
	var reqs []request

	var r1 request
	r1.url = "/v1/rekeningsb"
	r1.method = "GET"

	var r2 response
	r2.statusCode = 200

	var rheaders = make(map[string]string)
	rheaders["Content-Type"] = "application/json"
	rheaders["X-Header"] = "X-Value"

	r2.headers = rheaders
	r2.body = `{"status": "ok"}`

	r1.response = r2

	reqs = append(reqs, r1)

	var r11 request
	r11.url = "/delete"
	r11.method = "DELETE"

	var r12 response
	r12.statusCode = 200

	r12.headers = rheaders
	r12.body = `{"status": "ok"}`

	r11.response = r12

	reqs = append(reqs, r11)

	index := slices.IndexFunc(reqs, func(r request) bool {
		return r.method == method && r.url == url
	})

	var targetReq request

	if index != -1 {
		targetReq = reqs[index]

		response := targetReq.response

		for key, val := range response.headers {
			w.Header().Set(key, val)
		}
		w.WriteHeader(response.statusCode)

		if response.body != "" {
			w.Write([]byte(response.body))
		}

		return

	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		var errMsg = map[string]string{"message": "request not found"}
		result, err = json.Marshal(errMsg)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(result)
		return
	}

}
