package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func InitListeningStreak(id string) {
	sqlQueryString := fmt.Sprintf("UPDATE users SET listening_streak = 0 WHERE id = '%v'", id)
	os.WriteFile("query.sql", []byte(sqlQueryString), defaultOpenPermissions)
	defer exec.Command("rm", "query.sql").Run()
	command := "cat query.sql | sqlite3 database.db"
	err := exec.Command("bash", "-c", command).Run()
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
