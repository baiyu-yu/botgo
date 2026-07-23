package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sealdice/botgo"
	"github.com/sealdice/botgo/dto"
	"github.com/sealdice/botgo/dto/message"
	"github.com/sealdice/botgo/event"
	"github.com/sealdice/botgo/interaction/webhook"
	"github.com/sealdice/botgo/openapi"
	"github.com/sealdice/botgo/token"
	"gopkg.in/yaml.v3"
)

const (
	host_ = "0.0.0.0"
	port_ = 9000
	path_ = "/qqbot"
)

func main() {
	// 初始化新的文件logger，并使用相对路径来作为日志存放位置，设置鏈€灏忔棩蹇楃晫鍒负 DebugLevel
	logger, err := New("./", DebugLevel)
	if err != nil {
		log.Fatalln("error log new", err)
	}
	// 把新的logger 设置鍒?sdk 上，替换鎺夎€佺殑鎺у埗可logger
	botgo.SetLogger(logger)
	content, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalln("load config file failed, err:", err)
	}
	credentials := &token.QQBotCredentials{}
	if err = yaml.Unmarshal(content, &credentials); err != nil {
		log.Fatalln("parse config failed, err:", err)
	}

	tokenSource := token.NewQQBotTokenSource(credentials)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() //释放刷新协程
	if err = token.StartRefreshAccessToken(ctx, tokenSource); err != nil {
		log.Fatalln(err)
	}
	// 初始鍖?openapi，正式环澧?	api := botgo.NewOpenAPI(credentials.AppID, tokenSource).WithTimeout(5 * time.Second).SetDebug(true)
	// 根据不同的回调，生成 intents
	_ = event.RegisterHandlers(GuildATMessageEventHandler(api))
	http.HandleFunc(path_, func(writer http.ResponseWriter, request *http.Request) {
		webhook.HTTPHandler(writer, request, credentials)
	})
	if err = http.ListenAndServe(fmt.Sprintf("%s:%d", host_, port_), nil); err != nil {
		log.Fatal("setup server fatal:", err)
	}
}

// GuildATMessageEventHandler 实现处理 at 消息的回璋?func GuildATMessageEventHandler(api openapi.OpenAPI) event.ATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSATMessageData) error {
		log.Printf("[%s] %s", event.Type, data.Content)
		input := strings.ToLower(message.ETLInput(data.Content))
		log.Printf("clear input content is: %s", input)
		return nil
	}
}
