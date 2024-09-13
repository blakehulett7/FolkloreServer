package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/blakehulett7/goSqueal"
	"github.com/golang-jwt/jwt/v5"
)

func TestCreateUser(t *testing.T) {
	goSqueal.CheckForTable("users")
	jwtSecret := os.Getenv("JWT_SECRET")
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
			token, err := jwt.ParseWithClaims(response.Token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil })
			id, err := token.Claims.GetSubject()
			dbEntry := goSqueal.GetTableEntry("users", response.Id)
			if !reflect.DeepEqual(responseMap, dbEntry) {
				t.Fatalf("response does not match database entry, got %v, want %v...", responseMap, dbEntry)
			}
			goSqueal.DeleteTableEntry("users", response.Id)
		})
	}
}
