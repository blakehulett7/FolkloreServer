package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HelloWorld(writer http.ResponseWriter, request *http.Request) {
	message, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: "Christ is King!"})
	JsonResponse(writer, 200, message)
}

func CreateUser(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	postParams := struct {
		Username string `json:"username"`
	}{}
	err := decoder.Decode(&postParams)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(postParams.Username)
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}
