# ChatGPT_to_WechatBot
基于ChatGPT的微信机器人

# 引用
基于[openwechat](https://github.com/eatmoreapple/openwechat)的微信机器人

ChatGPT的接口参考了[wechat-chatGPT](https://github.com/gtoxlili/wechat-chatGPT)

# 使用方法
## cookie
打开浏览器并进入 ChatGPT 页面, 在请求中复制整个 cookie 到 cookie 文件

![GetSessionToken](https://github.com/lihongbin99/ChatGPT_to_WechatBot/blob/master/static/cookie.png?raw=true)

打开浏览器并进入 ChatGPT 页面, 在请求中复制整个 User-Agent 到 User_Agent 文件

![GetSessionToken](https://github.com/lihongbin99/ChatGPT_to_WechatBot/blob/master/static/User_Agent.png?raw=true)

# 注意事项
- 每隔2小时就需要手动更新一下 cookie
- 在浏览器进入 ChatGPT 获取 cookie 时使用的 ip 必须跟项目启动时的 ip 一致

# 编译
```
go build -o WechatBot.exe main.go
```