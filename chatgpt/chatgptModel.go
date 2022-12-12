package chatgpt

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
)

type ChatRequestBody struct {
	Action          string               `json:"action"`
	Messages        []ChatRequestMessage `json:"messages"`
	ConversationId  interface{}          `json:"conversation_id"`
	ParentMessageId string               `json:"parent_message_id"`
	Model           string               `json:"model"`
}

type ChatRequestMessage struct {
	Id      string             `json:"id"`
	Role    string             `json:"role"`
	Content ChatRequestContent `json:"content"`
}

type ChatRequestContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

func (msg *ChatRequestBody) ToJson() []byte {
	body, _ := json.Marshal(msg)
	return body
}

func CreateChatGPTRequestBody(message, parentID string, conversationId interface{}) *bytes.Buffer {
	req := &ChatRequestBody{
		Action: "next",
		Messages: []ChatRequestMessage{
			{
				Id:   uuid.New().String(),
				Role: "user",
				Content: ChatRequestContent{
					ContentType: "text",
					Parts:       []string{message},
				},
			},
		},
		ConversationId:  conversationId,
		ParentMessageId: parentID,
		Model:           "text-davinci-002-render",
	}
	return bytes.NewBuffer(req.ToJson())
}

/**********************************************************************************************************************/

type ChatGPTResponseBody struct {
	Message        ChatGPTResponseMessage `json:"message"`
	ConversationId string                 `json:"conversation_id"`
}

type ChatGPTResponseMessage struct {
	Id      string                 `json:"id"`
	Content ChatGPTResponseContent `json:"content"`
}

type ChatGPTResponseContent struct {
	Parts []string `json:"parts"`
}

func ToChatRes(body []byte) *ChatGPTResponseBody {
	var msg ChatGPTResponseBody
	_ = json.Unmarshal(body, &msg)
	return &msg
}
