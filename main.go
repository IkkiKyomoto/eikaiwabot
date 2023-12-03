package main

import (
	"fmt"
	"net/http"
	"os"
	"eikaiwabot/model"
	"github.com/line/line-bot-sdk-go/linebot"
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
					replyMessage := linebot.NewTextMessage(model.CreateMessage(message.Text))
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
	cert := "home/ec2-user/eikaiwabot/fullchain.pem"
	key := "home/ec2-user/eikaiwabot/privkey.pem"
	http.ListenAndServeTLS(":"+port, cert, key, nil)
}
