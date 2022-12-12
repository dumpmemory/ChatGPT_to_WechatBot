package config

import (
	"encoding/json"
	"log"
	"os"
)

// Configuration 项目配置
type Configuration struct {
	// gtp apikey
	ApiKey string `json:"api_key"`
	// 微信名称
	Master string `json:"master"`
	// 启动时是否判断ChatGPT是否可用
	JudgeChatGPT bool `json:"judge_chatgpt"`
}

var Config *Configuration

func init() {
	// 从文件中读取
	Config = &Configuration{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("open config err: %v", err)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	encoder := json.NewDecoder(f)
	err = encoder.Decode(Config)
	if err != nil {
		log.Fatalf("decode config err: %v", err)
		return
	}

	log.Println("api_key:", Config.ApiKey)
	log.Println("master:", Config.Master)
}
