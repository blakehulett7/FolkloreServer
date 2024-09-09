package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Christ is King!")
	fmt.Println("\nWelcome to the language learner web server... Name is wip and will change to something cool")
	CheckForDatabase()
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

func CheckForDatabase() {
	_, err := os.Stat("./database/users.db")
	if !errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Db exists")
		return
	}
	fmt.Println("Db does not exist, creating db")
	command := "cat init/users.sql | sqlite3 database/users.db"
	err = exec.Command("bash", "-c", command).Run()
	if err != nil {
		fmt.Println(err)
	}
}
