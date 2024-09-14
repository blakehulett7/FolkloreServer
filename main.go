package main

import (
	"fmt"
	"net/http"

	"github.com/blakehulett7/goSqueal"
)

func main() {
	fmt.Println("Christ is King!")
	fmt.Println("\nWelcome to Folklore!")
	goSqueal.CheckForTable("users")
	muxHandler := http.NewServeMux()
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: muxHandler,
	}
	muxHandler.HandleFunc("GET /v1/helloworld", HelloWorld)
	muxHandler.HandleFunc("POST /v1/users", CreateUser)
	muxHandler.HandleFunc("GET /v1/users", GetUser)
	server.ListenAndServe()
}

func JsonResponse(writer http.ResponseWriter, statusCode int, responseData []byte) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	writer.Write(responseData)
}
