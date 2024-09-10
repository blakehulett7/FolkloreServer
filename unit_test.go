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
	tests := map[string]struct {
		payloadStruct struct {
			Username string `json:"username"`
		}
	}{
		"simple": {
			payloadStruct: struct {
				Username string `json:"username"`
			}{"bhulett"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(test.payloadStruct)
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
				Id           string `json:"id"`
				Username     string `json:"username"`
				RefreshToken string `json:"refresh_token"`
			}{}
			err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatal("error: response in the wrong shape...")
			}
			responseMap := map[string]string{"id": response.Id, "username": response.Username, "refresh_token": response.RefreshToken}
			if !reflect.DeepEqual(responseMap, goSqueal.GetTableEntry("users", response.Id)) {
				t.Fatal("response does not match database entry...")
			}
			goSqueal.DeleteTableEntry("users", response.Id)
		})
	}
}
