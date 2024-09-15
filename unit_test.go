package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/blakehulett7/goSqueal"
)

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
	type payloadStruct struct {
		Username string `json:"username"`
	}
	tests := map[string]struct {
	}
}
