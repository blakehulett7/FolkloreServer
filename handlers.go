package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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
	//id := uuid.New()
	username := postParams.Username
	refreshToken := uuid.New()
	response := struct {
		Username     string    `json:"username"`
		RefreshToken uuid.UUID `json:"refresh_token"`
	}{username, refreshToken}
	responseData, err := json.Marshal(response)
	JsonResponse(writer, 201, responseData)
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}
