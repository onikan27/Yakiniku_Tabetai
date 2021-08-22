package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type responseType struct {
	Results results `json:"results"`
}

type results struct {
	Shop []shop `json:"shop"`
}

type shop struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Photo   photo  `json:"photo"`
	URLS    urls   `json:"urls"`
}

type photo struct {
	Mobile mobile `json:"mobile"`
}

type mobile struct {
	L string `json:"l"`
}

type urls struct {
	PC string `json:"pc"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}
	http.HandleFunc("/call", callHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func callHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := linebot.New(
		os.Getenv("LINE_SECRET_KEY"),
		os.Getenv("LINE_ACCESS_KEY"),
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
			switch event.Message.(type) {
			case *linebot.LocationMessage:
				sendYakinikuRestaurantInfo(bot, event)
			default:
				replyMessage := "位置情報を送ってください。"
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func sendYakinikuRestaurantInfo(bot *linebot.Client, event *linebot.Event) {
	message := event.Message.(*linebot.LocationMessage)

	latitude := strconv.FormatFloat(message.Latitude, 'f', 2, 64)
	longitude := strconv.FormatFloat(message.Longitude, 'f', 2, 64)

	replyMsg := getYakinikuRestaurantInfo(latitude, longitude)

	res := linebot.NewTemplateMessage(
		"焼肉一覧",
		linebot.NewCarouselTemplate(replyMsg...).WithImageOptions("rectangle", "cover"),
	)

	if _, err := bot.ReplyMessage(event.ReplyToken, res).Do(); err != nil {
		log.Print(err)
	}
}

func getYakinikuRestaurantInfo(latitude string, longitude string) []*linebot.CarouselColumn {
	apikey := os.Getenv("TABELOG_API_KEY")
	searchWord := "焼肉"
	endpoint := "https://webservice.recruit.co.jp/hotpepper/gourmet/v1/"

	url := fmt.Sprintf(
		"%s?format=json&key=%s&lat=%s&lng=%s&keyword=%s",
		endpoint, apikey, latitude, longitude, searchWord)

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var data responseType

	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	var carouselList []*linebot.CarouselColumn
	for _, shop := range data.Results.Shop {
		address := shop.Address
		if 60 < utf8.RuneCountInString(address) {
			address = string([]rune(address)[:60])
		}

		carouselItem := linebot.NewCarouselColumn(
			shop.Photo.Mobile.L,
			shop.Name,
			address,
			linebot.NewURIAction("詳しく見てみる", shop.URLS.PC),
		).WithImageOptions("#ffffff")
		carouselList = append(carouselList, carouselItem)
	}
	return carouselList
}
