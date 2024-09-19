package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/blakehulett7/goSqueal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const defaultOpenPermissions = 0777

var languages = []string{"Italian", "Spanish", "French"}

type User struct {
	Id              string `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	RefreshToken    string `json:"refresh_token"`
	ListeningStreak string `json:"listening_streak"`
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
		return ""
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		fmt.Println("error extracting id from token:", err)
		return ""
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
	InitListeningStreak(id)
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
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	isBadToken := ""
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	UserMap := goSqueal.GetTableEntry("users", id)
	user := User{
		Id:              UserMap["id"],
		Username:        UserMap["username"],
		ListeningStreak: UserMap["listening_streak"],
	}
	payload, err := json.Marshal(user)
	if err != nil {
		fmt.Println("error marshalling response:", err)
	}
	JsonResponse(writer, 200, payload)
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
	sqlQuery := fmt.Sprintf("SELECT id, password, refresh_token FROM users WHERE username = '%v'", loginParams.Username)
	os.WriteFile("query.sql", []byte(sqlQuery), fs.FileMode(defaultOpenPermissions))
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	entryData, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Println("error finding username:", err)
	}
	entry := string(entryData)
	entry = strings.ReplaceAll(entry, "\n", "")
	values := strings.Split(entry, "|")
	passwordMatches := bcrypt.CompareHashAndPassword([]byte(values[1]), []byte(loginParams.Password))
	if passwordMatches != nil {
		JsonHeaderResponse(writer, 401)
		return
	}
	jwt := GenerateJWT(values[0])
	res, err := json.Marshal(struct {
		JWT          string `json:"jwt"`
		RefreshToken string `json:"refresh_token"`
	}{jwt, values[2]})
	if err != nil {
		fmt.Println("Error couldn't marshal response:", err)
	}
	JsonResponse(writer, 200, res)
}
