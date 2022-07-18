package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

type WeatherResult struct {
	Main     MainStruct      `json:"main"`
	Weather  []WeatherStruct `json:"weather"`
	ErrorMsg string          `json:"message"`
	Name     string          `json:"name"`
	Sys      SysStruct       `json:"sys"`
}

type MainStruct struct {
	Temp       float32 `json:"temp"`
	Feels_like float32 `json:"feels_like"`
}

type WeatherStruct struct {
	Id    int    `json:"id"`
	Emoji string `json:"main"`
}

type SysStruct struct {
	Country string `json:"country"`
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TG_WEATHER_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnText, func(c tele.Context) error {

		userMsg := html.EscapeString(c.Message().Text)

		switch userMsg {
		case "/start":
			return c.Send("Введите город", tele.ForceReply)
		case "/help":
			return c.Send("Бот показывает текущую погоду. Введите название города для показа. Для уточнения можно добавить двухбуквенный код страны через запятую, например, Moscow, US")
		}

		query := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", userMsg, os.Getenv("TG_WEATHER_APIKEY"))

		resp, err := http.Get(query)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		defer resp.Body.Close()

		var data WeatherResult

		if err := json.Unmarshal(bs, &data); err != nil {
			return fmt.Errorf("%w", err)
		}

		var botReplyMsg string

		if data.ErrorMsg != "" {
			log.Printf("UserMsg: %s, ErrorMsg: %s", userMsg, data.ErrorMsg)
			botReplyMsg = "Город не найден"
			return c.Send(botReplyMsg)
		}

		var emoji string

		switch data.Weather[0].Emoji {
		case "Clouds":
			emoji = "\u2601\ufe0f"
		case "Clear":
			emoji = "\u2600\ufe0f"
		case "Rain":
			emoji = "\U0001f327\ufe0f"
		}

		botReplyMsg = fmt.Sprintf("%s (%s) %s \nПогода: %.1f°C \nПо ощущениям: %.1f°C \n", data.Name, data.Sys.Country, emoji, data.Main.Temp, data.Main.Feels_like)

		return c.Send(botReplyMsg)
	})

	b.Start()
}
