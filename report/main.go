package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"reflect"
	"time"

	resty "github.com/go-resty/resty/v2"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	api "github.com/influxdata/influxdb-client-go/v2/api"
	envconfig "github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Specification struct {
	InfluxOrg    string `required:"true" envconfig:"I_ORG"`
	InfluxBucket string `required:"true" envconfig:"I_BUCKET"`
	InfluxUrl    string `required:"true" envconfig:"I_URL"`
	InfluxToken  string `required:"true" envconfig:"I_TOKEN"`
	ChatId       string `required:"true" envconfig:"T_CHAT_ID"`
	BotId        string `required:"true" envconfig:"T_BOT_ID"`
}

func main() {
	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		log.Error().Err(err)
		panic("Environment variables not set!")
	}

	client := influxdb2.NewClient(s.InfluxUrl, s.InfluxToken)
	defer client.Close()
	queryApi := client.QueryAPI(s.InfluxOrg)

	dailyWh, err := getDailyWh(queryApi)
	fail(err, "Error querying InfluxDB: %v")

	restClient := resty.New().SetTimeout(5 * time.Second)
	msg := fmt.Sprintf("Guten Morgen! Gestern hat die Solaranlage %.2f kWh produziert!", dailyWh)

	log.Info().Msgf("Sending msg: %s", msg)
	err = send(msg, s.BotId, s.ChatId, *restClient)
	fail(err, "%v")
}

func getDailyWh(api api.QueryAPI) (float64, error) {
	result, err := api.Query(context.Background(), `
		import "date"
		import "experimental"

		from(bucket: "energy")
		|> range(start: experimental.subDuration(d: 1d, from: today()), stop: today())
		|> filter(fn: (r) => r["_field"] == "wh")
		|> filter(fn: (r) => r["meter"] == "solar")  
		|> sum(column: "_value")
		|> map(fn: (r) => ({r with _value: r._value / 1000.0 }))
	`)

	if err != nil {
		return math.NaN(), err
	}

	result.Next()

	if result.Record() == nil {
		log.Warn().Msg("influxDB returned empty result")
		// Do not send a message because this only occurs when there are no data points
		// Exit as successful to avoid rerunning
		os.Exit(0)
	}

	v := reflect.Indirect(reflect.ValueOf(result.Record().Value()))
	if !v.Type().ConvertibleTo(reflect.TypeOf(float64(0))) {
		return math.NaN(), fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	return v.Convert(reflect.TypeOf(float64(0))).Float(), nil
}

func fail(err error, msg string) {
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

func send(msg string, botId string, chatId string, client resty.Client) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botId)
	_, err := client.R().
		SetBody(map[string]interface{}{"chat_id": chatId, "text": msg}).
		Post(url)

	if err != nil {
		return fmt.Errorf("error sending to Telegram: %v", err.Error())
	}
	return nil
}
