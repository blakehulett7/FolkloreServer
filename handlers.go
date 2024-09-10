package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blakehulett7/goSqueal"
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
	id := uuid.NewString()
	params := map[string]string{
		"id":            id,
		"username":      postParams.Username,
		"refresh_token": uuid.NewString(),
	}
	goSqueal.CreateTableEntry("users", params)
	entryMap := goSqueal.GetTableEntry("users", id)
	responseStruct := struct {
		Id           string `json:"id"`
		Username     string `json:"username"`
		RefreshToken string `json:"refresh_token"`
	}{
		entryMap["id"], entryMap["username"], entryMap["refresh_token"],
	}
	responseData, err := json.Marshal(responseStruct)
	if err != nil {
		fmt.Println("can't marshal response to create user, error:", err)
	}
	JsonResponse(writer, 201, responseData)
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}
