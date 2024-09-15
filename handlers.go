package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/blakehulett7/goSqueal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HelloWorld(writer http.ResponseWriter, request *http.Request) {
	message, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: "Christ is King!"})
	JsonResponse(writer, 200, message)
}

func GenerateJWT(id string) string {
	jwtSecret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		Subject:   id,
	})
	jwt, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Println("error signing the jwt:", err)
	}
	return jwt
}

func GetIdFromJWT(tokenString string) string {
	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil })
	if err != nil {
		fmt.Println("error validating token:", err)
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		fmt.Println("error extracting id from token:", err)
	}
	return id
}

func CreateUser(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	postParams := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	err := decoder.Decode(&postParams)
	if err != nil {
		fmt.Println(err)
	}
	id := uuid.NewString()
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(postParams.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error generating password hash:", err)
	}
	password := string(passwordBytes)
	params := map[string]string{
		"id":            id,
		"password":      password,
		"username":      postParams.Username,
		"refresh_token": uuid.NewString(),
	}
	goSqueal.CreateTableEntry("users", params)
	entryMap := goSqueal.GetTableEntry("users", id)
	token := GenerateJWT(id)
	responseStruct := struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		token, entryMap["refresh_token"],
	}
	responseData, err := json.Marshal(responseStruct)
	if err != nil {
		fmt.Println("can't marshal response to create user, error:", err)
	}
	JsonResponse(writer, 201, responseData)
}

func GetUser(writer http.ResponseWriter, request *http.Request) {
}

func CheckUsername(writer http.ResponseWriter, request *http.Request) {
	reqParams := struct {
		Username string `json:"username"`
	}{}
	err := json.NewDecoder(request.Body).Decode(&reqParams)
	if err != nil {
		fmt.Println("bad request:", err)
		JsonHeaderResponse(writer, 400)
		return
	}
	if goSqueal.ParamExistsInTable("users", "username", reqParams.Username) {
		fmt.Println("username already exists")
		JsonHeaderResponse(writer, 208)
		return
	}
	JsonHeaderResponse(writer, 201)
}
