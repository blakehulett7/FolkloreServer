package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

func PurgeOldStreaks() {
	sqlQuery := fmt.Sprintf("SELECT user_id, language_id FROM users_languages;")
	data, err := OutputSqlQuery(sqlQuery)
	if err != nil {
		fmt.Println("Couldn't check on the listening streaks, error:", err)
		return
	}
	stringified := string(data)
	entries := strings.Split(stringified, "\n")
	entriesSlice := [][]string{}
	for _, entry := range entries {
		entrySlice := strings.Split(entry, "|")
		entriesSlice = append(entriesSlice, entrySlice)
	}
	entriesSlice = entriesSlice[:len(entriesSlice)-1]
	for _, entry := range entriesSlice {
		userID := entry[0]
		languageID := entry[1]
		lastListenedAt := GetLastListenedAt(userID, languageID)
		streakExpiresAt := lastListenedAt.Add(48 * time.Hour)
		if !time.Now().After(streakExpiresAt) {
			continue
		}
		sqlQuery = fmt.Sprintf("UPDATE users_languages SET current_listening_streak = 0 WHERE user_id = '%v' AND language_id = '%v';", userID, languageID)
		RunSqlQuery(sqlQuery)
	}
}

func PurgeStreaksWorker() {
	for {
		PurgeOldStreaks()
		time.Sleep(24 * time.Hour)
	}
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

func IncrementMyLanguageStat(userID, languageID, statToIncrement string) {
	sqlQuery := fmt.Sprintf("UPDATE users_languages SET %v = %v + 1 WHERE user_id = '%v' AND language_id = '%v';", statToIncrement, statToIncrement, userID, languageID)
	RunSqlQuery(sqlQuery)
}

func SetLastListenedAt(userID, languageID string) {
	timeString := time.Now().Format(timeFormat)
	sqlQuery := fmt.Sprintf("UPDATE users_languages SET last_listened_at = '%v' WHERE user_id = '%v' AND language_id = '%v';", timeString, userID, languageID)
	RunSqlQuery(sqlQuery)
}

func GetMyStatsStruct(userID, languageID string) Stats {
	sqlQuery := fmt.Sprintf("SELECT best_listening_streak, current_listening_streak, words_learned FROM users_languages WHERE user_id = '%v' AND language_id = '%v';", userID, languageID)
	data, err := OutputSqlQuery(sqlQuery)
	if err != nil {
		fmt.Println("Couldn't get my stats from the db...")
	}
	statsString := string(data)
	statsString = strings.ReplaceAll(statsString, "\n", "")
	statsSlice := strings.Split(statsString, "|")
	return Stats{
		BestListeningStreak:    statsSlice[0],
		CurrentListeningStreak: statsSlice[1],
		WordsLearned:           statsSlice[2],
	}
}

func GetLastListenedAt(userID, languageID string) time.Time {
	sqlQuery := fmt.Sprintf("SELECT last_listened_at FROM users_languages WHERE user_id = '%v' AND language_id = '%v';", userID, languageID)
	data, err := OutputSqlQuery(sqlQuery)
	if err != nil {
		fmt.Println("Couldn't get last listened at from the db...")
	}
	stringified := string(data)
	string := strings.ReplaceAll(stringified, "\n", "")
	if string == "" {
		fmt.Println("empty field")
		return time.Now().Add(-7 * 24 * time.Hour)
	}
	time, err := time.Parse(timeFormat, string)
	if err != nil {
		fmt.Println("Couldn't parse the stored last listend at date in the db, error:", err)
	}
	return time
}
