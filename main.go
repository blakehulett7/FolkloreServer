package main

import (
	"fmt"
	"net/http"

	"github.com/blakehulett7/goSqueal"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	fmt.Println("Christ is King!")
	fmt.Println("\nWelcome to Folklore!")
	goSqueal.CheckForTable("users")
	goSqueal.CheckForTable("languages")
	goSqueal.CheckForTable("users_languages")
	languageIds := GetLanguageIds()
	if languageIds == nil {
		fmt.Println("initializing languages")
		languageIds = InitializeLanguagesTable()
	}
	muxHandler := http.NewServeMux()
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: muxHandler,
	}
	muxHandler.HandleFunc("GET /v1/helloworld", HelloWorld)
	muxHandler.HandleFunc("POST /v1/users", CreateUser)
	muxHandler.HandleFunc("GET /v1/users", GetUser)
	muxHandler.HandleFunc("GET /v1/users/{username}", CheckUsername)
	muxHandler.HandleFunc("POST /v1/login", Login)
	muxHandler.HandleFunc("POST /v1/users_languages", AddLanguage)
	muxHandler.HandleFunc("GET /v1/users_languages", GetLanguages)
	muxHandler.HandleFunc("GET /v1/users_languages/{language_name}", GetMyLanguageStats)
	muxHandler.HandleFunc("DELETE /v1/users_languages/{language_name}", RemoveLanguage)
	muxHandler.HandleFunc("GET /v1/listen/{language_name}", ListenToLanguage)
	server.ListenAndServe()
}

func JsonResponse(writer http.ResponseWriter, statusCode int, responseData []byte) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	writer.Write(responseData)
}

func JsonHeaderResponse(writer http.ResponseWriter, statusCode int) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
}

func InitializeLanguagesTable() map[string]string {
	languageIds := map[string]string{}
	for idx, language := range languages {
		fmt.Println(language)
		id := uuid.NewString()
		goSqueal.CreateTableEntry("languages", map[string]string{
			"id":         id,
			"name":       language,
			"listen_url": listenUrls[idx],
		})
		languageIds[language] = id
	}
	return languageIds
}
