package chatgpt

import (
	"ChatGPT_to_WechatBot/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
)

type ImageRequestBody struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type ImageResponseBody struct {
	Data    []ImageURLItem `json:"data"`
	Created int            `json:"created"`
}

type ImageURLItem struct {
	Url string `json:"url"`
}

func GetDALLImage(requestText string, downLoadPath string) string {
	if len(config.Config.ApiKey) < 20 {
		log.Println("请配置api_key")
		return "服务器异常, 请稍后再试"
	}

	log.Println("向 Image 发送:", requestText)

	imagePath, err := CompletionsImage(requestText, downLoadPath)
	if err != nil {
		log.Printf("下载图片失败 %s\n", err)
		return "下载图片失败"
	}

	log.Printf("返回图片的地址是: %s", imagePath)
	return imagePath
}

func CompletionsImage(msg string, downPath string) (string, error) {
	requestBody := ImageRequestBody{
		Prompt: msg,
		N:      1,
		Size:   "1024x1024",
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", errors.New("问题格式异常")
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Config.ApiKey)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		log.Println(body)
		return "", errors.New(fmt.Sprintf("Image 响应状态码异常: %d", response.StatusCode))
	}

	gptResponseBody := &ImageResponseBody{}
	if err = json.Unmarshal(body, gptResponseBody); err != nil {
		log.Println(string(body))
		return "", errors.New(fmt.Sprintf("ImageResponseBody 解析响应体异常:%v", err))
	}

	imageName := uuid.New().String() + ".jpg"
	imagePath := downPath + "/" + imageName
	if len(gptResponseBody.Data) > 0 {
		for _, imageUrl := range gptResponseBody.Data {
			if err = downLoadImage(imagePath, imageUrl.Url); err != nil {
				return "", err
			}
			break
		}
	}
	return imagePath, nil
}

// 下载图片信息
func downLoadImage(base string, url string) error {
	log.Printf("开始下载图片 %s 到本地 %s\n", url, base)
	res, err := http.Get(url)
	if err != nil {
		return errors.New(fmt.Sprintf("下载图片时 Get 函数错误: %v", err))
	}

	defer func() {
		_ = res.Body.Close()
	}()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("下载图片时 Get 函数错误: %v", err))
	}

	err = ioutil.WriteFile(base, content, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("下载图片时 Get 函数错误: %v", err))
	}
	return nil
}
