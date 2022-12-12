package main

import (
	"ChatGPT_to_WechatBot/chatgpt"
	"ChatGPT_to_WechatBot/config"
	"github.com/eatmoreapple/openwechat"
	"log"
	"os"
	"strings"
)

func main() {
	startBot()
}

const (
	ChatGPT = "chatGPT"
	OpenAi  = "openai"
	Image   = "image"
)

var (
	Model = ChatGPT
)

// startBot 登录微信
func startBot() {
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// 注册消息处理函数
	bot.MessageHandler = HandlerMessage
	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 创建热存储容器对象
	reloadStorage := openwechat.NewJsonFileHotReloadStorage("storage.json")
	// 执行热登录
	err := bot.HotLogin(reloadStorage)
	if err != nil {
		if err = bot.Login(); err != nil {
			log.Printf("login error: %v \n", err)
			return
		}
	}
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	_ = bot.Block()
	exit()
}

// HandlerMessage 处理消息
func HandlerMessage(msg *openwechat.Message) {
	if msg.IsText() { // 暂时只处理文本消息
		if msg.IsSendByGroup() {
			// 群消息
			groupMessage(msg)
		} else {
			// 私聊
			privateMessage(msg)
		}
	}
}

// groupMessage 处理群组消息
func groupMessage(msg *openwechat.Message) {
	if !msg.IsAt() {
		return
	}

	sender, err := msg.Sender()
	if err != nil {
		log.Println("群组消息中获取 Sender 失败:", err)
		return
	}

	groupSender, err := msg.SenderInGroup()
	if err != nil {
		log.Println("群组消息中获取 SenderInGroup 失败:", err)
		return
	}
	log.Println("群:", msg.FromUserName, "用户:", groupSender.NickName, "发送了:", msg.Content)

	// 删除@
	atText := "@" + sender.Self.NickName
	replaceMessage := strings.TrimSpace(strings.ReplaceAll(msg.Content, atText, ""))
	if replaceMessage == "" {
		return
	}

	// 判断模型切换
	if groupSender.NickName == config.Config.Master {
		lowerMessage := strings.ToLower(replaceMessage)
		if strings.Contains(lowerMessage, "chatgpt") {
			Model = ChatGPT
			if err = replayUserText(msg, "\"切换成功，当前模型为 ChatGPT，我们可以使用对话的方式进行交互。\\n\""); err != nil {
				log.Printf("回复用户失败，%s", err)
			}
			return
		} else if strings.Contains(lowerMessage, "openai") {
			Model = OpenAi
			if err = replayUserText(msg, "切换成功，当前模型为 OpenAi，我可以获取当下的讯息。\n"); err != nil {
				log.Printf("回复用户失败，%s", err)
			}
			return
		} else if strings.Contains(replaceMessage, "生成图像") {
			Model = Image
			if err = replayUserText(msg, "切换成功，当前模型为 Image，我是一个可以通过文本描述中生成图像的人工智能程序。\n"); err != nil {
				log.Printf("回复用户失败，%s", err)
			}
			return
		}
	}

	if replaceMessage == "当前模型" {
		if err = replayUserText(msg, "当前模型为"+Model); err != nil {
			log.Printf("回复用户失败，%s", err)
		}
		return
	}

	if replaceMessage == "查看模型" {
		reply := "模型1: ChatGPT, 我们可以使用对话的方式进行交互.\n" +
			"模型2: OpenAi, 我可以获取当下的讯息,\n" +
			"模型3: Image,我是一个可以通过文本描述中生成图像的人工智能程序.\n"
		if err = replayUserText(msg, reply); err != nil {
			log.Printf("回复用户失败，%s", err)
		}
		return
	}

	// 发送逻辑

	if Model == ChatGPT {
		reply := chatgpt.GetChatGptMessage(replaceMessage, msg.FromUserName+":"+groupSender.NickName)
		if err = replayUserText(msg, reply); err != nil {
			log.Printf("ChatGPT 回复用户失败: %s \n", err)
		}
	} else if Model == OpenAi {
		reply := chatgpt.GetDavinciMessage(replaceMessage)
		if err = replayUserText(msg, reply); err != nil {
			log.Printf("OpenAi 回复用户失败: %v \n", err)
		}
	} else if Model == Image {
		reply := chatgpt.GetDALLImage(replaceMessage, chatgpt.DownLoadPath)
		log.Printf("微信读取文件路径：%s", reply)
		if err = replayUserImage(msg, reply); err != nil {
			log.Printf("Image 回复用户失败: %s \n", err)
		}
	}

}

// privateMessage 处理私聊消息
func privateMessage(msg *openwechat.Message) {

	sender, err := msg.Sender()
	if err != nil {
		log.Println("私聊消息中获取 Sender 失败:", err)
		return
	}
	log.Println("用户:", sender.NickName, "发送了:", msg.Content)

	// 获取 ChatGPT 消息
	chatGptMessage := chatgpt.GetChatGptMessage(msg.Content, sender.ID())

	// 回复
	chatGptMessage = strings.TrimSpace(chatGptMessage)
	chatGptMessage = strings.Trim(chatGptMessage, "\n")
	chatGptMessage = "ChatGPT 回复: \n" + chatGptMessage
	_, err = msg.ReplyText(chatGptMessage)
	if err != nil {
		log.Println("发送私聊消息失败:", err)
	}
	return
}

func exit() {
	log.Println("请输入任意字符退出程序")
	_, _ = os.Stdin.Read([]byte{0})
	os.Exit(0)
}

// replayUserText 回复用户文字
func replayUserText(msg *openwechat.Message, reply string) error {
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	groupSender, _ := msg.SenderInGroup()
	atText := "@" + groupSender.NickName
	replyText := atText + "chatGPT回复：\n" + reply
	_, err := msg.ReplyText(replyText)
	return err
}

// replayUserImage 回复用户图片
func replayUserImage(msg *openwechat.Message, imagePath string) error {
	file, _ := os.Open(imagePath)
	log.Println("回复图片读取的路径为", imagePath)
	_, err := msg.ReplyImage(file)
	return err
}
