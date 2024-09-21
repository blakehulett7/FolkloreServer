package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/blakehulett7/goSqueal"
)

func TestInit(t *testing.T) {
	goSqueal.CheckForTable("users")
	goSqueal.CheckForTable("languages")
	goSqueal.CheckForTable("users_languages")
	languageIds := GetLanguageIds()
	if languageIds == nil {
		languageIds = InitializeLanguagesTable()
	}
}

func TestCreateUser(t *testing.T) {
	goSqueal.CheckForTable("users")
	type payloadStruct struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	tests := map[string]payloadStruct{
		"simple": {"bhulett", "1234"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(test)
			if err != nil {
				t.Fatal("error: could not marshal payload")
			}
			req, err := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal("error: problem with the http request")
			}

			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(CreateUser)
			handler.ServeHTTP(responseRecorder, req)

			if responseRecorder.Code != 201 {
				t.Fatal("error: wrong status code...")
			}
			response := struct {
				Token        string `json:"token"`
				RefreshToken string `json:"refresh_token"`
			}{}
			err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatal("error: response in the wrong shape...")
			}
			responseMap := map[string]string{"token": response.Token, "refresh_token": response.RefreshToken}
			id := GetIdFromJWT(response.Token)
			dbEntry := goSqueal.GetTableEntry("users", id)
			if !reflect.DeepEqual(responseMap["refresh_token"], dbEntry["refresh_token"]) {
				t.Fatalf("response does not match database entry, got %v, want %v...", responseMap, dbEntry)
			}
			goSqueal.DeleteTableEntry("users", id)
		})
	}
}

func TestCheckUsername(t *testing.T) {
	goSqueal.CreateTableEntry("users", map[string]string{"id": "1", "username": "FireMage", "refresh_token": "asdf"})
	defer goSqueal.DeleteTableEntry("users", "1")
	tests := map[string]struct {
		Username string
		Want     int
	}{
		"simple":              {"bhulett", 200},
		"user already exists": {"FireMage", 208},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://test.com", bytes.NewBuffer([]byte("")))
			req.SetPathValue("username", test.Username)
			if err != nil {
				t.Fatalf("error making request: %v", err)
			}

			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(CheckUsername)
			handler.ServeHTTP(responseRecorder, req)

			if responseRecorder.Code != test.Want {
				fmt.Println(responseRecorder.Code, test.Want)
				t.Fatal("Wrong response code sent")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	type payloadStruct struct {
		Username string
		Password string
	}
	createUserParams, err := json.Marshal(payloadStruct{"bhulett", "1234"})
	request, err2 := http.NewRequest("POST", "/v1/users", bytes.NewBuffer(createUserParams))
	createFunc := http.HandlerFunc(CreateUser)
	createFunc.ServeHTTP(httptest.NewRecorder(), request)
	sqlQuery := "DELETE FROM users WHERE username = 'bhulett';"
	os.WriteFile("query2.sql", []byte(sqlQuery), defaultOpenPermissions)
	command := "cat query2.sql | sqlite3 database.db"
	defer exec.Command("rm", "query2.sql").Run()
	defer exec.Command("bash", "-c", command).Run()
	if err != nil || err2 != nil {
		t.Fatal("Couldn't create test user...")
	}
	tests := map[string]struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Want     string `json:"want"`
	}{
		"correct password":   {"bhulett", "1234", "refresh_token"},
		"incorrect password": {"bhulett", "5678", "refresh_token"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(test)
			if err != nil {
				t.Fatal("couldn't marshal json")
			}
			req, err := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal("error making the request...")
			}
			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(Login)
			handler.ServeHTTP(responseRecorder, req)

			resStruct := struct {
				JWT          string `json:"jwt"`
				RefreshToken string `json:"refresh_token"`
			}{}
			json.NewDecoder(responseRecorder.Body).Decode(&resStruct)
			fmt.Println(resStruct)
		})
	}
}

func TestInitListeningStreak(t *testing.T) {
	want := 0
	goSqueal.CreateTableEntry("users", map[string]string{
		"id":            "1",
		"username":      "firemage",
		"refresh_token": "asdf",
	})
	defer goSqueal.DeleteTableEntry("users", "1")
	InitListeningStreak("1")
	entryMap := goSqueal.GetTableEntry("users", "1")
	got, err := strconv.Atoi(entryMap["listening_streak"])
	if err != nil {
		t.Fatal("Couldn't parse listening streak value:", err)
	}
	if want != got {
		t.Fatalf("failed to initialize listening streak: expected %v, got %v", want, got)
	}
}

func TestIncrementListeningStreak(t *testing.T) {
	want := 1
	goSqueal.CreateTableEntry("users", map[string]string{
		"id":            "1",
		"username":      "firemage",
		"refresh_token": "asdf",
	})
	defer goSqueal.DeleteTableEntry("users", "1")
	InitListeningStreak("1")
	IncrementListeningStreak("1")
	entryMap := goSqueal.GetTableEntry("users", "1")
	got, err := strconv.Atoi(entryMap["listening_streak"])
	if err != nil {
		t.Fatal("Couldn't parse listening streak value:", err)
	}
	if want != got {
		t.Fatalf("failed to initialize listening streak: expected %v, got %v", want, got)
	}
}

func TestGetMyLanguages(t *testing.T) {
	languageIds := GetLanguageIds()
	for _, language := range languages {
		goSqueal.CreateTableEntry("users_languages", map[string]string{
			"user_id":     "1",
			"language_id": languageIds[language],
		})
	}
	got := GetMyLanguages("1")
	sqlQuery := "DELETE FROM users_languages WHERE user_id = '1'"
	RunSqlQuery(sqlQuery)
	if !reflect.DeepEqual(got, languages) {
		t.Fatalf("failed to get my languages: expected %v, got %v", languages, got)
	}
}

func TestGetMyStatsStruct(t *testing.T) {
	languageIds := GetLanguageIds()
	for _, language := range languages {
		goSqueal.CreateTableEntry("users_languages", map[string]string{
			"user_id":     "1",
			"language_id": languageIds[language],
		})
		InitMyLanguageStats("1", languageIds[language])
	}
	got := GetMyStatsStruct("1", languageIds["Italian"])
	want := Stats{BestListeningStreak: "0", CurrentListeningStreak: "0", WordsLearned: "0"}
	sqlQuery := "DELETE FROM users_languages WHERE user_id = '1'"
	RunSqlQuery(sqlQuery)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("failed to get my stats struct: expected %v, got %v", want, got)
	}
}

func TestPurgeOldStreaks(*testing.T) {
	PurgeOldStreaks()
}
