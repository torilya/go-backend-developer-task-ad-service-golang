package main

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
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

//goland:noinspection GoSnakeCaseUsage,LongLine
type bidRequest struct {
	RANDOM_UUID               string `schema:"-" valid:"-"`
	CLIENT_ID                 int    `schema:"client,required" valid:"required~The query parameter \"client\" cannot be 0"`
	SLOT_ID                   int    `schema:"slot,required" valid:"required~The query parameter \"slot\" cannot be 0"`
	USER_ID                   int    `schema:"user,required" valid:"required~The query parameter \"user\" cannot be 0"`
	IP_FROM_INCOMMING_REQUEST string `schema:"-" valid:"-"`
}

func adHandler(writer http.ResponseWriter, request *http.Request) {
	schemaDecoder := schema.NewDecoder()
	bidRequestData := &bidRequest{}

	if err := schemaDecoder.Decode(bidRequestData, request.URL.Query()); err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		if errs, ok := err.(schema.MultiError); ok {
			for _, err := range errs {
				switch err := err.(type) {
				case schema.UnknownKeyError:
					log.Printf("[ERR] query param: %s", err)
					fmt.Fprintf(writer, "The query parameter \"%s\" is unknown\n", err.Key)
				case schema.ConversionError:
					log.Printf("[ERR] query param: %s", err)
					fmt.Fprintf(writer, "The query parameter \"%s\" must be %s\n", err.Key,
						err.Type)
				case schema.EmptyFieldError:
					log.Printf("[ERR] query param: %s", err)
					fmt.Fprintf(writer, "The query parameter \"%s\" is required\n", err.Key)
				default:
					log.Printf("[ERR] query param: %#v", err)
					fmt.Fprintf(writer, "The query parameter is invalid\n")
				}
			}
		}

		return
	}

	if _, err := govalidator.ValidateStruct(bidRequestData); err != nil {
		writer.WriteHeader(http.StatusBadRequest)

		if errs, ok := err.(govalidator.Errors); ok {
			for _, err := range errs {
				log.Printf("[ERR] query param: %s", err)
				fmt.Fprintf(writer, "%s\n", err)
			}
		}

		return
	}

	bidRequestData.RANDOM_UUID = uuid.NewString()
	bidRequestData.IP_FROM_INCOMMING_REQUEST = request.RemoteAddr

	tmpl := template.Must(template.ParseFiles("tmpl/bid-request.json"))

	if err := os.MkdirAll("./bidRequests", os.ModePerm); err != nil {
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

	if err := tmpl.Execute(bidRequestFile, bidRequestData); err != nil {
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

	defer bidResponseFile.Close()

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

func middlewarePanic(handlerNext http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			err := recover()

			if err != nil {
				log.Printf("[ERR] panic: %s", err)
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			}
		}()

		handlerNext.ServeHTTP(writer, request)
	})
}

func main() {
	muxSite := http.NewServeMux()
	muxSite.HandleFunc("/ad", adHandler)

	handlerSite := middlewareAccessLog(muxSite)
	handlerSite = middlewarePanic(handlerSite)

	const pathLogs = "logs"

	if err := os.MkdirAll(pathLogs, os.ModePerm); err != nil {
		log.Fatalf("[ERR] dir: %s", err)
	}

	logFile, err := os.Create(filepath.Join(pathLogs, time.Now().Format("20060102_150405")+".log"))

	if err != nil {
		log.Fatalf("[ERR] file: %s", err)
	}

	defer logFile.Close()

	log.SetOutput(logFile)

	const serverAddr = ":8080"
	log.Printf("[INFO] The server is starting on %s", serverAddr)
	http.ListenAndServe(serverAddr, handlerSite)
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}
