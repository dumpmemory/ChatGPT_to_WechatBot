# ChatGPT_to_WechatBot
基于ChatGPT的微信机器人

# 引用
基于[openwechat](https://github.com/eatmoreapple/openwechat)的微信机器人

ChatGPT的接口参考了[wechat-chatGPT](https://github.com/gtoxlili/wechat-chatGPT)

# 注意事项
- 每隔2小时就需要手动更新一下 cookie
- 在浏览器进入 ChatGPT 获取 cookie 时使用的 ip 必须跟项目启动时的 ip 一致



# 配置项目
## 重要:cookie
打开浏览器并进入 ChatGPT 页面, 在请求中复制整个 cookie 到 cookie 文件

![GetSessionToken](https://github.com/lihongbin99/ChatGPT_to_WechatBot/blob/master/static/cookie.png?raw=true)

## 重要: User-Agent

打开浏览器并进入 ChatGPT 页面, 在请求中复制整个 User-Agent 到 User_Agent 文件

![GetSessionToken](https://github.com/lihongbin99/ChatGPT_to_WechatBot/blob/master/static/User_Agent.png?raw=true)

## 不重要: 配置文件
配置文件config.json
```
{
  // OpenAi的app_key 用作OpenAi模式和图片模式, 不需要可以不配置
  "api_key": "<openai_api_key>",
  
  // 你的微信用户名, 用作管理员修改
  "master": "微信用户名",
  
  // 用于启动检测ChatGPT是否初始化成功, true表示ChatGPT启动失败则退出程序, false表示需要用的时候再初始化, 初始化失败也不关闭
  "judge_chatgpt": true
}

```

# 使用方法
## 编译
```
go build -o WechatBot.exe main.go
```
- 管理员发送带有 chatgpt 的消息则会把 Bot 切换成 ChatGPT 模式
- 管理员发送带有 openai  的消息则会把 Bot 切换成 OpenAi  模式
- 管理员发送   生成图像    消息则会把 Bot 切换成 图片  模式