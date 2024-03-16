package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"strings"
)

type T struct {
}
type GithubRepo struct {
	Name  string `json:"name"`
	Owner struct {
	} `json:"owner"`
	Description string `json:"description"`
	URL         string `json:"html_url"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("токен тг")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {

		if update.InlineQuery != nil {
			go handleInlineQuery(bot, update.InlineQuery)
		}

		if update.Message == nil {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "/githubrepos") {
			handleGitHubReposCommand(bot, update)
			continue
		}

		if update.InlineQuery != nil {
			go handleInlineQuery(bot, update.InlineQuery)
		}
	}
}

func handleGitHubReposCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	username := strings.TrimSpace(update.Message.Text[len("/githubrepos"):])
	repos, err := getGithubRepos(username, "токен")
	if err != nil {
		log.Println("Failed to get GitHub repositories:", err)
		return
	}

	var response string
	if len(repos) > 0 {
		for _, repo := range repos {
			response += fmt.Sprintf("Repository: %s\nDescription: %s\nURL: %s\n\n", repo.Name, repo.Description, repo.URL)
		}
	} else {
		response = "No repositories found for the provided username."
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Failed to send message:", err)
	}
}

func handleInlineQuery(bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery) {
	repos, err := getGithubRepos("olma22a", "токен")
	if err != nil {
		log.Println("Failed to get GitHub repositories:", err)
		return
	}

	var results []interface{}
	for _, repo := range repos {
		article := tgbotapi.NewInlineQueryResultArticle(repo.Name, repo.Description, repo.URL)
		article.URL = repo.URL
		article.HideURL = false
		article.Description = repo.Description

		results = append(results, article)
	}

	answer := tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.ID,
		IsPersonal:    true,
		Results:       results,
	}

	if _, err := bot.AnswerInlineQuery(answer); err != nil {
		log.Println("Failed to answer inline query:", err)
	}
}

func getGithubRepos(username string, token string) ([]GithubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+"токен")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repos []GithubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}
