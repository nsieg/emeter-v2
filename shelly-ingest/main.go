package main

import (
	"context"
	"os"
	"time"

	resty "github.com/go-resty/resty/v2"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	envconfig "github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Specification struct {
	InfluxOrg    string `required:"true" envconfig:"I_ORG"`
	InfluxBucket string `required:"true" envconfig:"I_BUCKET"`
	InfluxUrl    string `required:"true" envconfig:"I_URL"`
	InfluxToken  string `required:"true" envconfig:"I_TOKEN"`
	ShellyUrl    string `required:"true" envconfig:"S_URL"`
}

type response struct {
	Counters  []float64 `json:"counters"`
	Is_valid  bool      `json:"is_valid"`
	Overpower float64   `json:"overpower"`
	Power     float64   `json:"power"`
	Timestamp int64     `json:"timestamp"`
	Total     float64   `json:"total"`
}

type Sender interface {
	Send(t int64, wh float64)
}

type influxSender struct {
	client influxdb2.Client
	api    api.WriteAPIBlocking
}

type Checker interface {
	Check(url string)
}

type restChecker struct{}

type dataPoint struct {
	Time int64
	Wh   float64
}

func (i influxSender) Send(t int64, wh float64) error {
	log.Debug().Msgf("sending %0.2f wh at %d to influx", wh, t)
	err := i.api.WritePoint(context.Background(), influxdb2.NewPointWithMeasurement("energy").
		AddTag("meter", "solar").AddField("wh", wh).SetTime(time.Unix(t, 0)))
	if err != nil {
		log.Error().Err(err).Msg("error writing to influx")
		return err
	}
	return nil
}

func (r restChecker) Check(url string) (*response, error) {
	log.Debug().Msgf("Querying shelly at %s", url)
	restClient := resty.New().SetTimeout(5 * time.Second)
	resp, err := restClient.R().SetResult(&response{}).
		Get(url)
	if err != nil {
		log.Error().Err(err).Msg("error querying shelly")
		return nil, err
	}
	return resp.Result().(*response), nil
}

func calcSends(res response) []dataPoint {
	firstFullMinute := res.Timestamp - (res.Timestamp % 60)
	return []dataPoint{
		{firstFullMinute, res.Counters[0] / 60},
		{firstFullMinute - 60, res.Counters[1] / 60},
		{firstFullMinute - 120, res.Counters[2] / 60},
	}
}

func main() {
	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		log.Error().Err(err)
		panic("Environment variables not set!")
	}

	shellyResponse, err := restChecker{}.Check(s.ShellyUrl)

	if err != nil {
		os.Exit(1)
	}

	influxClient := influxdb2.NewClient(s.InfluxUrl, s.InfluxToken)
	writeApi := influxClient.WriteAPIBlocking(s.InfluxOrg, s.InfluxBucket)
	sender := influxSender{influxClient, writeApi}

	for _, d := range calcSends(*shellyResponse) {
		err := sender.Send(d.Time, d.Wh)
		if err != nil {
			os.Exit(1)
		}
	}
}
