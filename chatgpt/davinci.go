package chatgpt

import (
	"ChatGPT_to_WechatBot/config"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// ChatGPTRequestImageBody 请求体
type ChatGPTRequestImageBody struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// ChatGPTResponseImageBody 响应体
type ChatGPTResponseImageBody struct {
	Data    []ImageURLItem `json:"data"`
	Created int            `json:"created"`
}

// ImageURLItem
type ImageURLItem struct {
	Url string `json:"url"`
}

func CompletionsImage(msg string, downPath string) (string, error) {
	CompletionsPngURL := "https://api.openai.com/v1/images/"
	requestBody := ChatGPTRequestImageBody{
		Prompt: msg,
		N:      1,
		Size:   "1024x1024",
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", CompletionsPngURL+"generations", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}
	apiKey := config.Config.ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	fmt.Println(response)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("gtp api status code not equals 200,code is %d", response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	gptResponseBody := &ChatGPTResponseImageBody{}
	log.Println("收到image图片响应----" + string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}
	imageName := md5V2(msg) + ".jpg"
	imagePath := downPath + "/" + imageName
	if len(gptResponseBody.Data) > 0 {
		for _, imageUrl := range gptResponseBody.Data {
			err := downLoadImage(imagePath, imageUrl.Url)
			if err != nil {
				log.Println("Download pic file failed!", err)
			} else {
				log.Println("Download file success.")
			}
		}
	}
	log.Printf("gpt response image path: %s \n", downPath)

	return imagePath, nil
}

func md5V2(str string) string {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	data := []byte(str + timeStr)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

// 下载图片信息
func downLoadImage(base string, url string) error {

	log.Printf("开始下载图片 %s 到本地 %s：", url, base)
	v, err := http.Get(url)
	if err != nil {
		log.Printf("Http get [%v] failed! %v \n", url, err)
		return err
	}
	defer v.Body.Close()
	content, err := ioutil.ReadAll(v.Body)
	if err != nil {
		log.Printf("Read http response failed! %v \n", err)
		return err
	}
	err = ioutil.WriteFile(base, content, 0666)
	if err != nil {
		log.Printf("Save to file failed! %v \n", err)
		return err
	}
	return nil
}
