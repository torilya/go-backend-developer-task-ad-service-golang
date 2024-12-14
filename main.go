package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const pathLogs = "logs"

//goland:noinspection GoSnakeCaseUsage
type bidRequest struct {
	RANDOM_UUID               string
	CLIENT_ID                 int `schema:"client"`
	SLOT_ID                   int `schema:"slot"`
	USER_ID                   int `schema:"user"`
	IP_FROM_INCOMMING_REQUEST string
}

func adHandler(writer http.ResponseWriter, request *http.Request) {
	schemaDecoder := schema.NewDecoder()
	bidRequestData := &bidRequest{}
	err := schemaDecoder.Decode(bidRequestData, request.URL.Query())

	// НЕ КРАСИВО: НЕ ИНФОРМАТИВНО
	if err != nil {
		log.Printf("[ERR] query values: %s", err)
		http.Error(writer, "Invalid query values", http.StatusBadRequest)
		return
	}

	bidRequestData.RANDOM_UUID = uuid.NewString()
	bidRequestData.IP_FROM_INCOMMING_REQUEST = request.RemoteAddr

	tmpl := template.Must(template.ParseFiles("tmpl/bid-request.json"))

	err = os.MkdirAll("./bidRequests", os.ModePerm)

	if err != nil {
		log.Printf("[ERR] dir: %s", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	bidRequestFile, err := os.Create("./bidRequests/" + bidRequestData.RANDOM_UUID + ".json")

	if err != nil {
		log.Printf("[ERR] file: %s", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer bidRequestFile.Close()

	err = tmpl.Execute(bidRequestFile, bidRequestData)

	if err != nil {
		log.Printf("[ERR] template: %s", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	bidResponseFile, err := os.Open("./bidResponses/bid-response.json")

	if err != nil {
		log.Printf("[ERR] file: %s", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonDecoder := json.NewDecoder(bidResponseFile)
	var tokenPrev json.Token

	for {
		tokenCurrent, err := jsonDecoder.Token()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("[ERR] JSON token: %s", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			break
		}

		if tokenPrev == "adm" {
			fmt.Fprint(writer, tokenCurrent)
		}

		tokenPrev = tokenCurrent
	}
}

func middlewareAccessLog(handlerNext http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		timeStart := time.Now()
		handlerNext.ServeHTTP(writer, request)
		log.Printf("[%s] %s, %s %s", request.Method, request.RemoteAddr, request.URL.String(),
			time.Since(timeStart))
	})
}

func main() {
	muxSite := http.NewServeMux()
	muxSite.HandleFunc("/ad", adHandler)

	handlerSite := middlewareAccessLog(muxSite)

	err := os.MkdirAll(pathLogs, os.ModePerm)

	if err != nil {
		log.Fatalf("[ERR] dir: %s", err)
	}

	logFile, err := os.Create(filepath.Join(pathLogs, time.Now().Format("20060102_150405")+".log"))

	if err != nil {
		log.Fatalf("[ERR] file: %s", err)
	}

	log.SetOutput(logFile)

	http.ListenAndServe(":8080", handlerSite)
}
