package chatgpt

import (
	"ChatGPT_to_WechatBot/config"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const BASEURL = "https://api.openai.com/v1/"

func GetChatGptMessage(requestText string, openId string) string {
	if DefaultGPT == nil {
		DefaultGPT = newChatGPT()
	}
	fmt.Println("向 ChatGPT 发送:", requestText)
	chatGptMessage := DefaultGPT.SendMsg(requestText, openId)
	chatGptMessage = strings.TrimSpace(chatGptMessage)
	chatGptMessage = strings.Trim(chatGptMessage, "\n")
	return chatGptMessage
}

func GetDALLImage(requestText string, downLoadPath string) string {
	fmt.Println("向 DALL-E 发送:", requestText)
	imagePath, err := CompletionsImage(requestText, downLoadPath)
	if err != nil {
		log.Printf("下载图片失败 %s", err)
		return "下载图片失败"
	}
	log.Printf("微信读取文件路径：%s", imagePath)
	return imagePath
}

func GetDavinciMessage(requestText string) string {
	requestBody := DavinciRequestBody{
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
		return "GetDavinciMessage 解析异常"
	}
	req, err := http.NewRequest("POST", BASEURL+"completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "GetDavinciMessage 请求异常"
	}

	apiKey := config.Config.ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "GetDavinciMessage http 请求异常"
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "GetDavinciMessage http 请求状态码异常"
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "GetDavinciMessage body IO读取异常"
	}

	gptResponseBody := &DavinciResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "body解析异常"
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Text
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply
}

var (
	cookiesFileName             = "cookie"
	User_AgentFileName          = "User_Agent"
	SessionTokenName            = "__Secure-next-auth.session-token"
	CfClearanceName             = "cf_clearance"
	DownLoadPath                = "images"
	Flag                        = "openai"
	DefaultGPT         *ChatGPT = nil
	userInfoMap                 = make(map[string]*userInfo)
	lock                        = sync.Mutex{}
)

type DavinciRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

// DavinciResponseBody 响应体
type DavinciResponseBody struct {
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

type ChatGPT struct {
	ok            bool
	authorization string
	sessionToken  string
	cf_clearance  string
	User_Agent    string
	timeOut       time.Time
}

type userInfo struct {
	parentID       string
	conversationId interface{}
	ttl            time.Time
}

func newChatGPT() *ChatGPT {
	cookies, err := os.ReadFile(cookiesFileName)
	if err != nil {
		log.Println("读取", cookiesFileName, "文件失败:", err)
		exit()
	}
	if len(cookies) < 100 {
		log.Println("你应该忘了配置", cookiesFileName, "文件")
		exit()
	}

	// 解析一下 sessionToken
	sessionToken := string(cookies)
	startIndex := strings.Index(sessionToken, SessionTokenName+"=")
	if startIndex != -1 {
		endIndex := strings.Index(sessionToken[startIndex:], ";")
		if endIndex != -1 {
			sessionToken = sessionToken[startIndex+len(SessionTokenName)+1 : startIndex+endIndex]
		} else {
			sessionToken = sessionToken[startIndex+len(SessionTokenName)+1:]
		}
	} else {
		log.Println("在 cookies 中没有查询到", SessionTokenName)
		exit()
	}

	// 解析一下 cf_clearance
	cf_clearance := string(cookies)
	startIndex = strings.Index(cf_clearance, CfClearanceName+"=")
	if startIndex != -1 {
		endIndex := strings.Index(cf_clearance[startIndex:], ";")
		if endIndex != -1 {
			cf_clearance = cf_clearance[startIndex+len(CfClearanceName)+1 : startIndex+endIndex]
		} else {
			cf_clearance = cf_clearance[startIndex+len(CfClearanceName)+1:]
			log.Println("cf_clearance = ", cf_clearance)
		}
	} else {
		log.Println("在 cookies 中没有查询到", CfClearanceName)
		exit()
	}

	// 获取一下 User-Agent
	User_AgentBytes, err := os.ReadFile(User_AgentFileName)
	if err != nil {
		log.Println("读取", User_AgentFileName, "文件失败:", err)
		exit()
	}

	User_Agent := string(User_AgentBytes)
	User_Agent = strings.TrimSpace(User_Agent)
	if strings.HasPrefix(User_Agent, "user-agent: ") {
		User_Agent = User_Agent[12:]
	}
	if len(User_Agent) == 0 {
		log.Println("你应该忘了配置", User_AgentFileName, "文件")
		exit()
	}
	log.Println("User_Agent:", User_Agent)

	gpt := &ChatGPT{
		sessionToken: sessionToken,
		cf_clearance: cf_clearance,
		User_Agent:   User_Agent,
		timeOut:      time.Now().Add(2 * time.Hour),
	}
	if !gpt.updateSessionToken() {
		exit()
	}

	// 每 10 分钟更新一次 sessionToken
	go func() {
		for range time.Tick(10 * time.Minute) {
			if !gpt.updateSessionToken() {
				gpt.ok = false
			}
		}
	}()
	return gpt
}

func (c *ChatGPT) updateSessionToken() bool {
	if c.timeOut.Before(time.Now()) {
		log.Println(CfClearanceName, "已失效, 请重新配置")
		return false
	}

	session, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		log.Println("更新 Token 调用 NewRequest 失败:", err)
		return false
	}
	session.Header.Set("User-Agent", c.User_Agent)
	session.AddCookie(&http.Cookie{
		Name:  SessionTokenName,
		Value: c.sessionToken,
	})
	session.AddCookie(&http.Cookie{
		Name:  CfClearanceName,
		Value: c.cf_clearance,
	})
	session.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.callback-url",
		Value: "https://chat.openai.com",
	})

	resp, err := http.DefaultClient.Do(session)
	if err != nil {
		log.Println("更新 Token 调用接口失败:", err)
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == SessionTokenName {
			c.sessionToken = cookie.Value
			newCookie := SessionTokenName + "=" + cookie.Value + ";" + CfClearanceName + "=" + c.cf_clearance
			_ = os.WriteFile(cookiesFileName, []byte(newCookie), 0600)
			break
		}
	}
	var accessToken map[string]interface{}
	bodyByes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("更新 Token 获取响应数据失败:", err)
		return false
	}
	err = json.Unmarshal(bodyByes, &accessToken)
	if err != nil {
		log.Println("更新 Token 解析响应数据失败(可能是需要更新", CfClearanceName, "):", err)
		//log.Println("解析响应数据:", string(bodyByes))
		return false
	}
	c.authorization = accessToken["accessToken"].(string)
	log.Println("sessionToken 更新成功")
	c.ok = true
	return true
}

func (c *ChatGPT) SendMsg(msg, openId string) string {
	if !c.ok {
		log.Println("请处理 ChatGPT 的 sessionToken 更新失败")
		return "服务器异常, 请联系管理员"
	}

	lock.Lock()
	defer lock.Unlock()

	info, ok := userInfoMap[openId]
	if !ok || info.ttl.Before(time.Now()) {
		log.Printf("用户 %s 启动新的对话\n", openId)
		info = &userInfo{
			parentID:       uuid.New().String(),
			conversationId: nil,
		}
		userInfoMap[openId] = info
	} else {
		log.Printf("用户 %s 继续对话\n", openId)
	}
	info.ttl = time.Now().Add(5 * time.Minute)

	// 发送请求
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", CreateChatReqBody(msg, info.parentID, info.conversationId))
	if err != nil {
		log.Println("调用 ChatGPT 的 NewRequestWithContext 异常:", err)
		return "服务器异常, 请稍后再试"
	}
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Authorization", "Bearer "+c.authorization)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Openai-Assistant-App-Id", "")
	req.Header.Set("Connection", "close")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("origin", "https://chat.openai.com")
	req.Header.Set("Referer", "https://chat.openai.com/chat")

	req.Header.Set("User-Agent", c.User_Agent)
	req.AddCookie(&http.Cookie{
		Name:  SessionTokenName,
		Value: c.sessionToken,
	})
	req.AddCookie(&http.Cookie{
		Name:  CfClearanceName,
		Value: c.cf_clearance,
	})
	req.AddCookie(&http.Cookie{
		Name:  "__Secure-next-auth.callback-url",
		Value: "https://chat.openai.com",
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("调用 ChatGPT 接口异常:", err)
		return "服务器异常, 请稍后再试"
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取 ChatGPT 数据异常", err)
		return "服务器异常, 请稍后再试"
	}
	line := bytes.Split(bodyBytes, []byte("\n\n"))
	if len(line) < 2 {
		log.Println("回复数据格式不对:", string(bodyBytes))
		return "服务器异常, 请稍后再试"
	}
	endBlock := line[len(line)-3][6:]
	res := ToChatRes(endBlock)
	info.conversationId = res.ConversationId
	info.parentID = res.Message.Id
	if len(res.Message.Content.Parts) > 0 {
		return res.Message.Content.Parts[0]
	} else {
		return "没有获取到..."
	}
}

func exit() {
	log.Println("请输入任意字符退出程序")
	_, _ = os.Stdin.Read([]byte{0})
	os.Exit(0)
}
