package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/blakehulett7/goSqueal"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const defaultOpenPermissions = 0777
const timeFormat = time.RFC3339

var languages = []string{"Italian", "Spanish", "French"}
var listenUrls = []string{"https://www.youtube.com/watch?v=hxre-NJtPKU", "https://www.youtube.com/watch?v=-xGws7UkWx0", "https://www.youtube.com/watch?v=zxngbsXXg5M"}
var isBadToken = ""

type User struct {
	Id              string   `json:"id"`
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	RefreshToken    string   `json:"refresh_token"`
	ListeningStreak string   `json:"listening_streak"`
	Languages       []string `json:"languages"`
}

type Stats struct {
	BestListeningStreak    string `json:"best_listening_streak"`
	CurrentListeningStreak string `json:"current_listening_streak"`
	WordsLearned           string `json:"words_learned"`
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
		Languages:       GetMyLanguages(UserMap["id"]),
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

func AddLanguage(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	isBadToken := ""
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	params := struct {
		Name string `json:"name"`
	}{}
	err := json.NewDecoder(request.Body).Decode(&params)
	if err != nil {
		fmt.Println("Couldn't decode json, error:", err)
		JsonHeaderResponse(writer, 400)
		return
	}
	languageIds := GetLanguageIds()
	goSqueal.CreateTableEntry("users_languages", map[string]string{
		"user_id":     id,
		"language_id": languageIds[params.Name],
	})
	InitMyLanguageStats(id, languageIds[params.Name])
	userMap := goSqueal.GetTableEntry("users", id)
	user := User{
		Username:        userMap["username"],
		ListeningStreak: userMap["listening_streak"],
		Languages:       GetMyLanguages(id),
	}
	payload, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Couldn't marshal response json after adding language, error:", err)
	}
	JsonResponse(writer, 201, payload)
}

func GetLanguages(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	isBadToken := ""
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	myLanguages := GetMyLanguages(id)
	res := struct {
		Languages []string `json:"languages"`
	}{myLanguages}
	payload, err := json.Marshal(res)
	if err != nil {
		fmt.Println("Couldn't marshal requester's languages, error:", err)
	}
	JsonResponse(writer, 200, payload)
}

func RemoveLanguage(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	isBadToken := ""
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	languageName := request.PathValue("language_name")
	languageIdMap := GetLanguageIds()
	sqlQuery := fmt.Sprintf("DELETE FROM users_languages WHERE user_id = '%v' AND language_id = '%v';", id, languageIdMap[languageName])
	RunSqlQuery(sqlQuery)
	user := User{
		Languages: GetMyLanguages(id),
	}
	payload, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Couldn't marshal json response after deleting language, error:", err)
	}
	JsonResponse(writer, 200, payload)
}

func GetMyLanguageStats(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	isBadToken := ""
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	languageName := request.PathValue("language_name")
	languageIdMap := GetLanguageIds()
	languageId := languageIdMap[languageName]
	fmt.Println(languageId)
	stats := GetMyStatsStruct(id, languageId)
	payload, err := json.Marshal(stats)
	if err != nil {
		fmt.Println("Couldn't marshal my language stats json, error:", err)
	}
	JsonResponse(writer, 200, payload)
}

func ListenToLanguage(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	language := request.PathValue("language_name")
	language_idx := slices.Index(languages, language)
	res := struct {
		Url string `json:"url"`
	}{listenUrls[language_idx]}
	payload, bug := json.Marshal(res)
	if bug != nil {
		fmt.Println("Bug! Can't marshal the listen url for", language+", error:", bug)
		JsonHeaderResponse(writer, 500)
		return
	}
	JsonResponse(writer, 200, payload)
}

func IncrementMyListeningStreak(writer http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("Authorization")
	id := GetIdFromJWT(token)
	if id == isBadToken {
		JsonHeaderResponse(writer, 401)
		return
	}
	language := request.PathValue("language_name")
	languageIds := GetLanguageIds()
	languageID := languageIds[language]
	lastListenedAt := GetLastListenedAt(id, languageID)
	if !time.Now().After(lastListenedAt.Add(24 * time.Hour)) {
		JsonHeaderResponse(writer, 200)
		return
	}
	IncrementMyLanguageStat(id, languageID, "current_listening_streak")
	SetLastListenedAt(id, languageID)
	stats := GetMyStatsStruct(id, languageID)
	oldHighStreak, err := strconv.Atoi(stats.BestListeningStreak)
	if err != nil {
		fmt.Println("Couldn't parse the listening streak high score to check if it needs to be updated, error:", err)
		JsonHeaderResponse(writer, 200)
		return
	}
	currentStreak, err := strconv.Atoi(stats.CurrentListeningStreak)
	if err != nil {
		fmt.Println("Couldn't parse the listening streak high score to check if it needs to be updated, error:", err)
		JsonHeaderResponse(writer, 200)
		return
	}
	if oldHighStreak < currentStreak {
		IncrementMyLanguageStat(id, languageID, "best_listening_streak")
		return
	}
	userMap := goSqueal.GetTableEntry("users", id)
	bestStreak, err := strconv.Atoi(userMap["listening_streak"])
	if err != nil {
		fmt.Println("Couldn't parse the listening streak high score to check if it needs to be updated, error:", err)
		JsonHeaderResponse(writer, 200)
		return
	}
	newHighStreak := max(oldHighStreak, currentStreak)
	if bestStreak < newHighStreak {
		sqlQuery := fmt.Sprintf("UPDATE users SET listening_streak = %v WHERE id = '%v';", newHighStreak, id)
		RunSqlQuery(sqlQuery)
	}
	userMap = goSqueal.GetTableEntry("users", id)
	user := User{ListeningStreak: userMap["listening_streak"]}
	payload, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Couldn't parse the json for the updated user return for real time streak update, error:", err)
		JsonHeaderResponse(writer, 200)
		return
	}
	JsonResponse(writer, 200, payload)
}
