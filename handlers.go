package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/blakehulett7/goSqueal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const defaultOpenPermissions = 0777

type User struct {
	Id           string
	Username     string
	Password     string
	RefreshToken string
}

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
	username := request.PathValue("username")
	if goSqueal.ParamExistsInTable("users", "username", username) {
		JsonHeaderResponse(writer, 208)
		return
	}
	JsonHeaderResponse(writer, 200)
}

func Login(writer http.ResponseWriter, request *http.Request) {
	loginParams := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	err := json.NewDecoder(request.Body).Decode(&loginParams)
	if err != nil {
		fmt.Println("error decoding login params:", err)
		return
	}
	sqlQuery := fmt.Sprintf("SELECT password FROM users WHERE username = '%v'", loginParams.Username)
	os.WriteFile("query.sql", []byte(sqlQuery), fs.FileMode(defaultOpenPermissions))
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	entryData, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println("error finding username:", err)
	}
	entry := string(entryData)
	fmt.Println(entry)
	passwordMatches := bcrypt.CompareHashAndPassword(entryData, []byte(loginParams.Password))
	if passwordMatches != nil {
		fmt.Println("password is incorrect")
		return
	}
	fmt.Println("password is correct")
}
