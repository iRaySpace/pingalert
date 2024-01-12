package main

import (
	"os"
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	"strings"
)

type Config struct {
	WebhookUrl string
	Servers []Server
}

type Server struct {
	Url string
	ExpectedStatusCode int
}

type ComparisonResult struct {
	Url string
	HasExpectedStatus bool
	ExpectedStatus int
	Status int
	ResponseTime int64
}

type EmbedField struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type EmbedData struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Fields []EmbedField `json:"fields"`
}

type WebhookData struct {
	Content string `json:"content"`
	Embeds []EmbedData `json:"embeds"`
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	fmt.Println("Reading config from: ", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
	var results[] ComparisonResult
	for _, server := range config.Servers {
		startTime := time.Now()
		r, err := http.Get(server.Url)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		if r.StatusCode != server.ExpectedStatusCode {
			fmt.Println("Error: ", server.Url, " is down")
		}
		responseTime := time.Since(startTime).Milliseconds()
		result := ComparisonResult{
			Url: server.Url,
			HasExpectedStatus: server.ExpectedStatusCode == r.StatusCode,
			ExpectedStatus: server.ExpectedStatusCode,
			Status: r.StatusCode,
			ResponseTime: responseTime,
		}
		results = append(results, result)
	}
	embedFields := []EmbedField{}
	for _, result := range results {
		isOk := result.HasExpectedStatus
		okText := "✅"
		if !isOk {
			okText = "❌"
		}
		embedFields = append(embedFields, EmbedField{
			Name: result.Url,
			Value: fmt.Sprintf("**Response Time**: %dms\n**Status**: %d %s", result.ResponseTime, result.Status, okText),
		})
	}
	description := fmt.Sprintf("%d server(s) as of %s", len(results), time.Now().Format("2006-01-02 15:04:05"))
	webhookData := &WebhookData{
		Embeds: []EmbedData{
			{
				Title: "Server Status",
				Description: description,
				Fields: embedFields,
			},
		},
	}
	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		panic(err)
	}
	_, err2 := http.Post(config.WebhookUrl, "application/json", strings.NewReader(string(jsonData)))
	if err2 != nil {
		panic(err)
	}
}