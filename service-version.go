package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Deploy corresponds to a single entry of a deploy
type Deploy struct {
	Date    string `json:"date"`
	Version string `json:"version"`
	Restart bool   `json:"restart"`
}

// Rollback corresponds to rollback version
type Rollback struct {
	Version string `json:"version"`
}

// Response corresponds to the http response for a service deploy
type Response struct {
	Current  Deploy   `json:"current"`
	Rollback Rollback `json:"rollback"`
	History  []Deploy `json:"history"`
}

// ResponseError is a http response error
type ResponseError struct {
	Status string `json:"status"`
	Err    error  `json:"error"`
}

func main() {
	fmt.Println("service versions")

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/services/{service}", RetrieveServiceHandler).Methods("GET")
	r.HandleFunc("/services/{service}/current", RetrieveCurrentServiceHandler).Methods("GET")
	r.HandleFunc("/services/{service}/rollback", RetrieveRollbackServiceHandler).Methods("GET")
	r.HandleFunc("/services/{service}/version/{version}", StoreServiceHandler).Methods("POST", "PUT")
	r.HandleFunc("/services/{service}/version/{version}/{restart}", StoreServiceHandler).Methods("POST", "PUT")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HomeHandler handles all requests and returns usage
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w,
		`Usage: GET /services/{service_name}
Usage: GET /services/{service_name}/current
-------------------------------------------
Response format:
{
  "date": "06-18-2019:20:58:50",
  "version": "a.a.b",
  "restart": false
}

Usage: GET /services/{service_name}/rollback
--------------------------------------------
Response Format:
{
  "version": "a.a.a"
}

Usage: POST|PUT /services/{service_name}/deploy/{version}
----------------------------------------------------
Response format:
{
	"current": {
		"date": "12/30/2020:15:09:05",
		"version": "W.Y.Z",
		"restart": false
	},
	"rollback": {
		"version": "X.Y.Z"
	},
	"history" : [
		{
			"date": "12/30/2020:15:08:05",
			"version": "W.Y.Z",
			"restart": false
		},
		{
			"date": "12/29/2020:15:07:05",
			"version": "X.Y.Z",
			"restart": true
		},
		{
			"date": "12/28/2020:15:06:05",
			"version": "X.Y.Z",
			"restart": true
		},
		{
			"date": "12/27/2020:15:05:05",
			"version": "X.Y.Z",
			"restart": true
		},
		{
			"date": "12/26/2020:15:04:05",
			"version": "X.Y.Z",
			"restart": true
		}
	]
}
`)
}

// RetrieveServiceHandler handles service requests and returns versions or errors
func RetrieveServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseError := new(ResponseError)
	responseError.Status = http.StatusText(http.StatusInternalServerError)
	vars := mux.Vars(r)
	service, err := findServiceFromFile(vars["service"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responseError.Err = err
		resp, _ := json.Marshal(responseError)
		fmt.Fprintf(w, string(resp))
	} else {
		resp, err := json.Marshal(service)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseError.Err = err
			resp, _ := json.Marshal(responseError)
			fmt.Fprintf(w, string(resp))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(resp))
		}
	}
}

// RetrieveCurrentServiceHandler retrieves just the current version
func RetrieveCurrentServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseError := new(ResponseError)
	responseError.Status = http.StatusText(http.StatusInternalServerError)
	vars := mux.Vars(r)
	service, err := findServiceFromFile(vars["service"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responseError.Err = err
		resp, _ := json.Marshal(responseError)
		fmt.Fprintf(w, string(resp))
	} else {
		resp, err := json.Marshal(service.Current)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseError.Err = err
			resp, _ := json.Marshal(responseError)
			fmt.Fprintf(w, string(resp))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(resp))
		}
	}
}

// RetrieveRollbackServiceHandler retrieves just the rollback version
func RetrieveRollbackServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseError := new(ResponseError)
	responseError.Status = http.StatusText(http.StatusInternalServerError)
	vars := mux.Vars(r)
	service, err := findServiceFromFile(vars["service"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responseError.Err = err
		resp, _ := json.Marshal(responseError)
		fmt.Fprintf(w, string(resp))
	} else {
		resp, err := json.Marshal(service.Rollback)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseError.Err = err
			resp, _ := json.Marshal(responseError)
			fmt.Fprintf(w, string(resp))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(resp))
		}
	}
}

func findServiceFromFile(serviceName string) (Response, error) {
	resp := new(Response)
	_ = resp
	f, err := ioutil.ReadDir(serviceName)
	if err != nil {
		return *resp, err
	}
	for _, file := range f {
		f, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", serviceName, file.Name()))
		var current Deploy
		var rollback Rollback
		var history []Deploy
		switch file.Name() {
		case "current":
			json.Unmarshal(f, &current)
			resp.Current = current
		case "rollback":
			json.Unmarshal(f, &rollback)
			resp.Rollback = rollback
		case "history":
			json.Unmarshal(f, &history)
			resp.History = history
		default:
			log.Printf("What file is that: %s", file.Name())
		}
	}
	if err != nil {
		log.Panic(err)
		return *resp, err
	}
	return *resp, nil
}

// StoreServiceHandler is the service for data persistence
func StoreServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responseError := new(ResponseError)
	responseError.Status = http.StatusText(http.StatusInternalServerError)

	vars := mux.Vars(r)
	_, err := ioutil.ReadDir(vars["service"])
	if err != nil {
		createFirstTime(vars["service"], vars["version"])
	} else {
		addDeployVersion(vars["service"], vars["version"], vars["restart"])
	}
	RetrieveServiceHandler(w, r)
}

func createFirstTime(serviceName string, version string) {
	err := os.Mkdir(serviceName, os.FileMode(0764))
	if err != nil {
		log.Panic(err) // Because if we fail to create a dir, something is very wrong
	}
	currentTime := time.Now().Format("01-02-2006:15:04:05")

	dep := Deploy{Version: version, Date: currentTime}
	depEntry, _ := json.Marshal(dep)
	rb, _ := json.Marshal(Rollback{Version: version})
	hist := make([]Deploy, 1)
	hist[0] = dep
	histEntry, _ := json.Marshal(hist)

	ioutil.WriteFile(fmt.Sprintf("%s/current", serviceName), depEntry, os.FileMode(0764))
	ioutil.WriteFile(fmt.Sprintf("%s/rollback", serviceName), rb, os.FileMode(0764))
	ioutil.WriteFile(fmt.Sprintf("%s/history", serviceName), histEntry, os.FileMode(0764))
}

func addDeployVersion(serviceName string, version string, restart string) {
	currentTime := time.Now().Format("01-02-2006:15:04:05")
	// unless it's restart, move current version to rollback
	new := Deploy{Version: version, Date: currentTime, Restart: len(restart) > 0}
	depEntry, _ := json.Marshal(new)

	if len(restart) == 0 {
		// This means it's not a restart
		// so we need to push the current version to the rollback version
		var current Deploy
		f, _ := ioutil.ReadFile(fmt.Sprintf("%s/current", serviceName))
		json.Unmarshal(f, &current)

		rb, _ := json.Marshal(Rollback{Version: current.Version})
		ioutil.WriteFile(fmt.Sprintf("%s/rollback", serviceName), rb, os.FileMode(0764))
	}

	// replace current with new current
	ioutil.WriteFile(fmt.Sprintf("%s/current", serviceName), depEntry, os.FileMode(0764))

	// add that to the history
	var history []Deploy
	his, _ := ioutil.ReadFile(fmt.Sprintf("%s/history", serviceName))
	json.Unmarshal(his, &history)
	reverse(history)
	history = append(history, new)
	reverse(history)
	histEntry, _ := json.Marshal(history)
	ioutil.WriteFile(fmt.Sprintf("%s/history", serviceName), histEntry, os.FileMode(0764))
}

func reverse(deploys []Deploy) {
	for i, j := 0, len(deploys)-1; i < j; i, j = i+1, j-1 {
		deploys[i], deploys[j] = deploys[j], deploys[i]
	}
}
