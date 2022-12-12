package chatgpt

import (
	"fmt"
	"log"
)

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
