package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunSqlQuery(sqlQueryString string) error {
	os.WriteFile("query.sql", []byte(sqlQueryString), defaultOpenPermissions)
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	return exec.Command("bash", "-c", command).Run()
}

func OutputSqlQuery(sqlQueryString string) ([]byte, error) {
	os.WriteFile("query.sql", []byte(sqlQueryString), defaultOpenPermissions)
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	return exec.Command("bash", "-c", command).Output()
}

func InitListeningStreak(id string) {
	sqlQueryString := fmt.Sprintf("UPDATE users SET listening_streak = 0 WHERE id = '%v'", id)
	err := RunSqlQuery(sqlQueryString)
	if err != nil {
		fmt.Println("Couldn't initialize listening streak, error:", err)
	}
}

func IncrementListeningStreak(id string) {
	sqlQueryString := fmt.Sprintf("UPDATE users SET listening_streak = listening_streak + 1 WHERE id = '%v'", id)
	os.WriteFile("query.sql", []byte(sqlQueryString), defaultOpenPermissions)
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	err := exec.Command("bash", "-c", command).Run()
	if err != nil {
		fmt.Println("Couldn't increment listening streak, error:", err)
	}
}

func GetLanguageIds() map[string]string {
	languageIds := map[string]string{}
	for _, language := range languages {
		sqlQueryString := fmt.Sprintf("SELECT id FROM languages WHERE name = '%v'", language)
		os.WriteFile("query.sql", []byte(sqlQueryString), defaultOpenPermissions)
		command := "cat query.sql | sqlite3 database.db"
		output, err := exec.Command("bash", "-c", command).Output()
		if err != nil {
			fmt.Println("Couldn't execute query, error:", err)
		}
		id := string(output)
		if id == "" {
			exec.Command("rm", "query.sql").Run()
			return nil
		}
		id = strings.ReplaceAll(id, "\n", "")
		languageIds[language] = id
		exec.Command("rm", "query.sql").Run()
	}
	return languageIds
}

func GetMyLanguages(userID string) []string {
	sqlQueryString := fmt.Sprintf("SELECT name FROM languages WHERE id IN (SELECT language_id FROM users_languages WHERE user_id = '%v');", userID)
	data, err := OutputSqlQuery(sqlQueryString)
	if err != nil {
		fmt.Println("Couldn't retrieve my languages, error:", err)
	}
	languageString := string(data)
	languageSlice := strings.Split(languageString, "\n")
	return languageSlice[:len(languageSlice)-1]
}

func InitMyLanguageStats(userID, languageID string) {
	sqlQuery := fmt.Sprintf("UPDATE users_languages SET best_listening_streak = 0, current_listening_streak = 0, words_learned = 0 WHERE user_id = '%v' AND language_id = '%v';", userID, languageID)
	err := RunSqlQuery(sqlQuery)
	if err != nil {
		fmt.Println("Couldn't initialize language stats...")
	}
}
