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

var (
	cookiesFileName             = "cookie"
	User_AgentFileName          = "User_Agent"
	SessionTokenName            = "__Secure-next-auth.session-token"
	CfClearanceName             = "cf_clearance"
	DefaultGPT         *ChatGPT = nil
	DefaultGPTLock              = sync.Mutex{}
	DefaultGPTSuccess1          = false // 用来判断是否成功过
	userInfoMap                 = make(map[string]*userInfo)
)

type ChatGPT struct {
	ok            bool
	authorization string
	sessionToken  string
	cf_clearance  string
	User_Agent    string
	timeOut       time.Time
	lock          sync.Mutex
}

type userInfo struct {
	parentID       string
	conversationId interface{}
	ttl            time.Time
}

func init() {
	if config.Config.JudgeChatGPT {
		DefaultGPT = newChatGPT()
		if DefaultGPT == nil {
			exit()
		}
	}
}

func GetChatGptMessage(requestText string, openId string) string {
	// 单例
	if DefaultGPT == nil {
		DefaultGPTLock.Lock()
		defer DefaultGPTLock.Unlock()
		if DefaultGPT == nil {
			DefaultGPT = newChatGPT()
		}
	}
	if DefaultGPT == nil {
		return "服务器异常, 请稍后再试"
	}

	fmt.Println("向 ChatGPT 发送:", requestText)
	chatGptMessage := DefaultGPT.SendMsg(requestText, openId)
	chatGptMessage = strings.TrimSpace(chatGptMessage)
	chatGptMessage = strings.Trim(chatGptMessage, "\n")
	return chatGptMessage
}

func newChatGPT() (gpt *ChatGPT) {
	gpt = &ChatGPT{
		sessionToken: "",
		cf_clearance: "",
		User_Agent:   "",
		timeOut:      time.Now().Add(2 * time.Hour),
		lock:         sync.Mutex{},
	}

	cookies, err := os.ReadFile(cookiesFileName)
	if err != nil {
		log.Println("读取", cookiesFileName, "文件失败:", err)
		return
	}
	if len(cookies) < 100 {
		log.Println("你应该忘了配置", cookiesFileName, "文件")
		return
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
		return
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
		return
	}

	// 获取一下 User-Agent
	User_AgentBytes, err := os.ReadFile(User_AgentFileName)
	if err != nil {
		log.Println("读取", User_AgentFileName, "文件失败:", err)
		return
	}

	User_Agent := string(User_AgentBytes)
	User_Agent = strings.TrimSpace(User_Agent)
	if strings.HasPrefix(User_Agent, "user-agent: ") {
		User_Agent = User_Agent[12:]
	}
	if len(User_Agent) == 0 {
		log.Println("你应该忘了配置", User_AgentFileName, "文件")
		return
	}
	log.Println("User_Agent:", User_Agent)

	gpt.sessionToken = sessionToken
	gpt.cf_clearance = cf_clearance
	gpt.User_Agent = User_Agent

	if !gpt.updateSessionToken() {
		return
	}

	// 每 10 分钟更新一次 sessionToken
	go func() {
		for range time.Tick(10 * time.Minute) {
			if !gpt.updateSessionToken() {
				gpt.ok = false
			}
		}
	}()
	return
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
	if err = json.Unmarshal(bodyByes, &accessToken); err != nil {
		log.Println("更新 Token 解析响应数据失败", err)
		if !DefaultGPTSuccess1 {
			DefaultGPTSuccess1 = true
			println("大概率是 ip 问题, 请重点检查ip")
			println("或者是 User_Agent 的问题")
		} else {
			println("可能是" + CfClearanceName + "获取了")
		}
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

	c.lock.Lock()
	defer c.lock.Unlock()

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
	info.ttl = time.Now().Add(10 * time.Minute)

	// 发送请求
	req, err := http.NewRequest("POST", "https://chat.openai.com/backend-api/conversation", CreateChatGPTRequestBody(msg, info.parentID, info.conversationId))
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
