package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/google/uuid"

	"github.com/hashicorp/go-memdb"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appsLog *log.Logger
)

func main() {
	initializeLog()

	var err error

	db, err := initializeDB()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	log.Println("Starting api-simulator..")

	insertJsonDataToDb(db)

	// Set up apps
	mux := http.DefaultServeMux

	mux.HandleFunc("/", apiHandler(db)) // Register handler for all routes

	var handler http.Handler = mux
	handler = MiddlewareLog(handler)

	server := new(http.Server)

	port := "8800"
	server.Addr = ":" + port
	server.Handler = handler

	log.Printf("api-simulator started on port %s", port)
	appsLog.Printf("api-simulator started on port %s", port)

	err = server.ListenAndServe()

	if err != nil {
		appsLog.Fatalf("Could not start api-simulator: %s\n", err)
	}

}

func initializeLog() {
	// Ensure log directory exists
	if err := os.MkdirAll("log", 0755); err != nil {
		log.Fatal("Error creating log directory:", err)
	}

	// appsLogFile, err2 := os.OpenFile("log/apps.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// traficsLogFile, err1 := os.OpenFile("log/trafics.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	appsLogFile := &lumberjack.Logger{
		Filename:   "log/apps.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     1,     //days
		Compress:   false, // disabled by default
		LocalTime:  true,  // use local time for timestamps
	}

	traficsLogFile := &lumberjack.Logger{
		Filename:   "log/trafics.log",
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
}

func initializeDB() (*memdb.MemDB, error) {
	// Get schema from APIData.go
	schema := GetAPIDataSchema()

	// Create a new database
	db, err := memdb.NewMemDB(schema)

	if err != nil {
		return nil, fmt.Errorf("error creating memdb: %v", err)
	}

	return db, nil
}

func apiHandler(db *memdb.MemDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		url := r.RequestURI
		contentType := r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Panicf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Close the body
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		txn := db.Txn(false)
		defer txn.Abort()

		it, err := txn.Get("apidata", "method_url", method, url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if it != nil {
			var targetApi APIData

			for obj := it.Next(); obj != nil; obj = it.Next() {
				p := obj.(APIData)
				if checkBody(contentType, string(body), p.Request.Body) {
					targetApi = p
				}
			}

			if !reflect.DeepEqual(targetApi, APIData{}) {
				response := targetApi.Response

				for key, val := range response.Headers {
					w.Header().Set(key, val)
				}
				w.WriteHeader(response.StatusCode)

				if response.Body != "" {
					w.Write([]byte(response.Body))
				}

				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		errMsg := map[string]string{"message": "request not found"}

		result, err := json.Marshal(errMsg)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(result)
	}
}

func checkBody(bodyType string, rBody string, dBody string) bool {
	if bodyType == "json" {
		var obj1, obj2 map[string]interface{}

		err1 := json.Unmarshal([]byte(rBody), &obj1)
		err2 := json.Unmarshal([]byte(dBody), &obj2)

		if err1 != nil || err2 != nil {
			fmt.Println("Error unmarshaling JSON")
			return false
		}

		if reflect.DeepEqual(obj1, obj2) {
			fmt.Println("JSON strings are logically equal")
		} else {
			fmt.Println("JSON strings are different")
		}
	}

	return rBody == dBody
}

func insertJsonDataToDb(db *memdb.MemDB) {
	apiDataList, err := readAPIDataJson()
	if err != nil {
		log.Printf("Error reading API data JSON: %v", err)
		return
	}

	txn := db.Txn(true)
	// defer txn.Abort()

	for _, apiData := range apiDataList {
		apiData.ID = uuid.NewString() // Generate a new UUID for each APIData entry

		log.Printf("Inserting API data into DB: %v", apiData)

		if err := txn.Insert("apidata", apiData); err != nil {
			log.Printf("Error inserting API data into DB: %v", err)
		}
	}

	txn.Commit()
}

// readAPIDataJson reads data from data/api-data.json
func readAPIDataJson() ([]APIData, error) {
	// Open the JSON file
	file, err := os.Open("data/api-data.json")
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
