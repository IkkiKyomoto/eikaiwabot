package main

import (
	"eikaiwabot/model"
	"fmt"
	"net/http"
	"os"

	//line-bot-sdk-goをインポート
	"github.com/line/line-bot-sdk-go/linebot"
)

// 証明書と鍵ファイルのパスを環境変数から取得
var (
	cert = os.Getenv("CERT_PATH")
	key  = os.Getenv("KEY_PATH")
)

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)

	if err != nil {
		fmt.Println("LINEボットの初期化に失敗しました:", err)
		return
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		w.WriteHeader(200)
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					replyMessage := linebot.NewTextMessage(model.Handler(message.Text, event.Source.UserID))
					_, err := bot.ReplyMessage(event.ReplyToken, replyMessage).Do()
					if err != nil {
						fmt.Println("返信メッセージの送信に失敗しました:", err)
					}
				}
			}
		}
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "50124"
	}
	http.ListenAndServeTLS(":"+port, cert, key, nil)
}
