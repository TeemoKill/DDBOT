package lsp

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Sora233/DDBOT/requests"

	"github.com/Sora233/MiraiGo-Template/config"
)

type ChatGPTResp struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (lgc *LspGroupCommand) ChatCommand() {
	log := lgc.DefaultLoggerWithCommand("ChatCommand")
	log.Infof("run %v command", "ChatCommand")
	defer func() { log.Infof("%v command end", "ChatCommand") }()

	var err error

	// retrieve the entire input, including the command text
	firstWord := lgc.CommandName()
	if strings.HasPrefix(firstWord, "，") {
		firstWord = strings.TrimPrefix(firstWord, "，")
	} else if strings.HasPrefix(firstWord, ",") {
		firstWord = strings.TrimPrefix(firstWord, ",")
	}
	chatContent := []string{firstWord}
	chatContent = append(chatContent, lgc.Args...)
	chatPrompt := fmt.Sprintf(
		"[current time: %s]%s%s",
		time.Now().Format("2006-01-02 Mon (UTC+8)15:04:05"),
		strings.Join(chatContent, " "),
		"<|endoftext|>",
	)
	log.WithField("chatPrompt", chatPrompt).Infof("chat command prompt")

	// call chatgpt api and receive entire reply
	apiAddr := config.GlobalConfig.GetString("chatGPT.apiAddr")
	apiKey := config.GlobalConfig.GetString("chatGPT.apiKey")

	opts := []requests.Option{
		requests.HeaderOption("Content-Type", "application/json"),
		requests.HeaderOption("Authorization", fmt.Sprintf("Bearer %s", apiKey)),
		requests.TimeoutOption(time.Second * 15),
		requests.RetryOption(3),
	}
	params := map[string]interface{}{
		"model":       "text-davinci-003",
		"prompt":      chatPrompt,
		"temperature": 1,
		"max_tokens":  800,
	}

	var body = new(bytes.Buffer)

	err = requests.PostJson(apiAddr, params, body, opts...)
	if err != nil {
		lgc.textSend("陷入了混乱")
		log.WithField("error", err).Errorf("chat call gpt api error")
		return
	}

	resp := &ChatGPTResp{}

	err = json.Unmarshal(body.Bytes(), resp)
	if err != nil {
		lgc.textSend("陷入了迷茫")
		log.WithField("error", err).Errorf("chat unmarshal gpt api response error")
		return
	}

	gptReply := resp.Choices[0].Text
	for strings.HasPrefix(gptReply, "\n") {
		gptReply = strings.TrimPrefix(gptReply, "\n")
	}
	log.WithField("gptReply", gptReply).Infof("gpt reply")

	// reply to group
	lgc.textReply(gptReply)

}
