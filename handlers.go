package main

import (
	"encoding/json"
	"net/http"
)

func HelloWorld(writer http.ResponseWriter, request *http.Request) {
	message, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: "Christ is King!"})
	JsonResponse(writer, 200, message)
}

func CreateUser(writer http.ResponseWriter, request *http.Request) {
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}
