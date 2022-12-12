package chatgpt

import (
	"ChatGPT_to_WechatBot/config"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type OpenAiRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

type OpenAiResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceItem struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	Logprobs     int    `json:"logprobs"`
	FinishReason string `json:"finish_reason"`
}

func GetOpenAiMessage(requestText string) string {
	if len(config.Config.ApiKey) < 20 {
		log.Println("请配置api_key")
		return "服务器异常, 请稍后再试"
	}

	requestBody := OpenAiRequestBody{
		Model:            "text-davinci-003",
		Prompt:           requestText,
		MaxTokens:        2048,
		Temperature:      0.7,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "问题消息格式异常"
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(requestData))
	if err != nil {
		log.Println("GetOpenAiMessage 的 NewRequest 异常:", err)
		return "服务器异常, 请稍后再试"
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Config.ApiKey)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println("GetOpenAiMessage 调用接口异常:", err)
		return "服务器异常, 请稍后再试"
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("GetOpenAiMessage 读取响应数据异常:", err)
		return "服务器异常, 请稍后再试"
	}

	if response.StatusCode != 200 {
		log.Println("GetOpenAiMessage 响应状态码异常:", response.StatusCode)
		log.Println(string(body))
		return "服务器异常, 请稍后再试"
	}

	gptResponseBody := &OpenAiResponseBody{}
	if err = json.Unmarshal(body, gptResponseBody); err != nil {
		log.Println("GetOpenAiMessage 解析响应体异常:", err)
		log.Println(string(body))
		return "服务器异常, 请稍后再试"
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Text
	}
	return reply
}
