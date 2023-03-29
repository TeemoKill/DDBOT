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
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
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
	chatPrompt := strings.Join(chatContent, " ")

	log.WithField("chatPrompt", chatPrompt).Infof("chat command prompt")

	// call chatgpt api and receive entire reply
	apiAddr := config.GlobalConfig.GetString("chatGPT.apiAddr")
	apiKey := config.GlobalConfig.GetString("chatGPT.apiKey")

	gptReply, err := lgc.callChatGPT(apiAddr, apiKey, chatPrompt)
	if err != nil {
		log.WithError(err).
			Errorf("call chatgpt error")
		return
	}

	// reply to group
	lgc.textReply(gptReply)
}

func (lgc *LspGroupCommand) callChatGPT(apiAddr string, apiKey string, chatPrompt string) (reply string, err error) {
	log := lgc.DefaultLoggerWithCommand("callChatGPT")
	log.Infof("run %v command", "callChatGPT")

	opts := []requests.Option{
		requests.HeaderOption("Content-Type", "application/json"),
		requests.HeaderOption("Authorization", fmt.Sprintf("Bearer %s", apiKey)),
		requests.TimeoutOption(time.Second * 20),
		requests.RetryOption(5),
	}
	params := map[string]interface{}{
		"model": config.GlobalConfig.GetString("chatGPT.model"),
		"messages": []map[string]string{
			{
				"role": "system",
				"content": fmt.Sprintf("current time: %s ; %s",
					time.Now().Format("2006-01-02 Mon (UTC+8)15:04:05"),
					config.GlobalConfig.GetString("chatGPT.reminder"),
				),
			},
			{"role": "user", "content": chatPrompt},
		},
		"temperature": 1,
		"max_tokens":  800,
	}

	var body = new(bytes.Buffer)
	err = requests.PostJson(apiAddr, params, body, opts...)
	if err != nil {
		lgc.textSend("陷入了混乱")
		log.WithField("error", err).Errorf("call chatgpt api error")
		return
	}

	resp := &ChatGPTResp{}
	err = json.Unmarshal(body.Bytes(), resp)
	if err != nil {
		lgc.textSend("陷入了迷茫")
		log.WithField("error", err).Errorf("chat unmarshal gpt api response error")
		return
	}

	reply = resp.Choices[0].Message.Content
	for strings.HasPrefix(reply, "\n") {
		reply = strings.TrimPrefix(reply, "\n")
	}
	log.WithField("reply", reply).Infof("gpt reply")

	return reply, nil
}
