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
		"id":           id,
		"username":     postParams.Username,
		"refreshToken": uuid.NewString(),
	}
	goSqueal.CreateTableEntry("users", params)
	fmt.Println(goSqueal.GetTableEntry("users", id))
	//responseData, err := json.Marshal(response)
	//JsonResponse(writer, 201, responseData)
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}
