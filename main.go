package main

import (
	"eikaiwabot/model"
	"fmt"
	"net/http"
	"os"

	//line-bot-sdk-goをインポート
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

// 環境変数を取得
var (
	KEY_PATH       = os.Getenv("CERT_PATH")
	CERT_PATH      = os.Getenv("KEY_PATH")
	CHANNEL_SECRET = os.Getenv("CHANNEL_SECRET")
	PORT           = os.Getenv("PORT")
	CHANNEL_TOKEN  = os.Getenv("CHANNEL_TOKEN")
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("環境変数の読み込みに失敗しました:", err)
	}
	bot, err := linebot.New(
		CHANNEL_SECRET,
		CHANNEL_TOKEN,
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
	if PORT == "" {
		PORT = "443"
	}
	http.ListenAndServeTLS(":"+PORT, CERT_PATH, KEY_PATH, nil)
}
