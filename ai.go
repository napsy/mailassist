package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jmorganca/ollama/api"
)

type LLM interface {
	bio(string)
	summary(string) (string, error)
	actionItems(string) ([]string, error)
}

type ollamaLLM struct {
	model     string
	biography string
	c         *api.Client
}

func newOllama(model string) (*ollamaLLM, error) {
	c, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &ollamaLLM{
		model: model,
		c:     c,
	}, nil
}

var (
	promptMessage = "Create a short summary in bullet points and any possible action items for me of the following email. Based on my given bio, propritize the action items accordingly by using either 'low', 'med' or 'high' qualifiers and identify the urgency and importance of the message. Always separate action items from the summary. The email conversation is : "
	promptSystem  = "You are an assitant that summarizes email conversations. When interpreting the contex of the email content, use my bio to identify action item priorities. My bio is: "
)

func (ollama *ollamaLLM) bio(summary string) {
	ollama.biography = summary
}

func (ollama *ollamaLLM) summary(msg string) (string, error) {
	a := ""
	//prompt := "Create a really short summary in bullet points of the following email: " + msg
	rq := api.GenerateRequest{
		Model:     "zephyr",
		Prompt:    promptMessage + msg,
		Template:  "",
		System:    promptSystem + ollama.biography,
		Context:   []int{},
		Raw:       false,
		Format:    "",
		KeepAlive: &api.Duration{},
		Images:    []api.ImageData{},
		Options:   map[string]interface{}{},
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(5*time.Minute))
	if err := ollama.c.Generate(ctx, &rq, func(resp api.GenerateResponse) error {
		a += resp.Response
		return nil
	}); err != nil {
		return "", err
	}
	return a, nil
}

func (ollama *ollamaLLM) actionItems(msg string) ([]string, error) {
	return []string{}, nil
}

// Define structures to match the JSON response format
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int              `json:"index"`
	Message      Message          `json:"message"`
	Logprobs     *json.RawMessage `json:"logprobs"` // Use a pointer to RawMessage for potential null value
	FinishReason string           `json:"finish_reason"`
}

type Response struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Choices           []Choice       `json:"choices"`
	Usage             map[string]int `json:"usage"`
	SystemFingerprint string         `json:"system_fingerprint"`
}

type openAI struct {
	biography string
	token     string
}

func newOpenAI(token string) (*openAI, error) {
	return &openAI{token: token}, nil
}

func (openai *openAI) bio(summary string) {
	openai.biography = summary
}

func (openai *openAI) summary(msg string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": promptSystem + openai.biography},
			{"role": "user", "content": promptMessage + msg},
		},
		"temperature": 0.7,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshalling payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openai.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	responseContent := ""
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
		responseContent = response.Choices[0].Message.Content
	}

	return responseContent, nil
}

func (openai *openAI) actionItems(msg string) ([]string, error) {
	return []string{}, nil
}
