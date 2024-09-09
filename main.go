package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Christ is King!")
	fmt.Println("\nWelcome to the language learner web server... Name is wip and will change to something cool")
	CheckForDatabases()
	muxHandler := http.NewServeMux()
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: muxHandler,
	}
	muxHandler.HandleFunc("GET /v1/helloworld", HelloWorld)
	server.ListenAndServe()
}

func JsonResponse(writer http.ResponseWriter, statusCode int, responseData []byte) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	writer.Write(responseData)
}

func HelloWorld(writer http.ResponseWriter, request *http.Request) {
	message, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: "Christ is King!"})
	JsonResponse(writer, 200, message)
}

func CheckForDatabases() {

}
