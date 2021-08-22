package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	http.HandleFunc("/", confirmHandler)
	http.HandleFunc("/call", callHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func confirmHandler(w http.ResponseWriter, r *http.Request) {
	msg := "動作確認"
	fmt.Fprintf(w, msg)
}

func callHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := linebot.New(
		"シークレットキー",
		"アクセスキー",
	)
	if err != nil {
		log.Fatal(err)
	}

	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				replyMessage := message.Text
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do()
				if err != nil {
					log.Print(err)
				}
			case *linebot.LocationMessage:
				sendRestaurantInfo(bot, event)
			}
		}
	}
}

func sendRestaurantInfo(bot *linebot.Client, event *linebot.Event) {
	msg := event.Message.(*linebot.LocationMessage)

	latitude := strconv.FormatFloat(msg.Latitude, 'f', 2, 64)
	longitude := strconv.FormatFloat(msg.Longitude, 'f', 2, 64)

	replyMsg := fmt.Sprintf("緯度：%s\n経度：%s", latitude, longitude)
	_, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do()
	if err != nil {
		log.Print(err)
	}
}
