package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"text/template"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	envconfig "github.com/kelseyhightower/envconfig"
	log "github.com/rs/zerolog/log"
)

type Specification struct {
	InfluxOrg    string `required:"true" envconfig:"I_ORG"`
	InfluxBucket string `required:"true" envconfig:"I_BUCKET"`
	InfluxUrl    string `required:"true" envconfig:"I_URL"`
	InfluxToken  string `required:"true" envconfig:"I_TOKEN"`
	DataDir      string `required:"true" envconfig:"DATA_DIR"`
}

func main() {
	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		log.Error().Err(err)
		panic("Environment variables not set!")
	}

	client := influxdb2.NewClient(s.InfluxUrl, s.InfluxToken)
	queryApi := client.QueryAPI(s.InfluxOrg)

	queries := [6]string{
		QueryFromTemplate("query_yesterday_deriv.tmpl", "main-reading"),
		QueryFromTemplate("query_today_deriv.tmpl", "main-reading"),
		QueryFromTemplate("query_yesterday.tmpl", "solar"),
		QueryFromTemplate("query_today.tmpl", "solar"),
		QueryFromTemplate("query_yesterday.tmpl", "main-acute-power"),
		QueryFromTemplate("query_today.tmpl", "main-acute-power"),
	}

	fileNames := [6]string{
		"main-" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".csv",
		"main-" + time.Now().Format("2006-01-02") + ".csv",
		"solar-" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".csv",
		"solar-" + time.Now().Format("2006-01-02") + ".csv",
		"main-acute-" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + ".csv",
		"main-acute-" + time.Now().Format("2006-01-02") + ".csv",
	}

	for i := range queries {
		res, err := queryApi.Query(context.Background(), queries[i])

		if err != nil {
			log.Error().Err(err).Msg("error querying InfluxDB")
			exit(client)
		}

		csvFile, err := os.Create(os.Getenv("DATA_DIR") + "/" + fileNames[i])

		if err != nil {
			log.Error().Err(err).Msg("failed creating file")
			exit(client)
		}

		csvwriter := csv.NewWriter(csvFile)

		for res.Next() {
			var unixTimestamp string = fmt.Sprintf("%v", res.Record().ValueByKey("_unix"))
			var wh string = fmt.Sprintf("%v", res.Record().ValueByKey("_value"))
			_ = csvwriter.Write([]string{unixTimestamp, wh})
		}

		if res.Err() != nil {
			log.Error().Err(err).Msg("error querying InfluxDB")
			exit(client)
		}

		csvwriter.Flush()
		csvFile.Close()
	}

	client.Close()
}

func QueryFromTemplate(tmplFilePath string, meterName string) string {
	tmpl, err := template.ParseFiles(tmplFilePath)
	if err != nil {
		log.Error().Err(err).Msg("error processing query template")
		os.Exit(1)
	}

	vars := make(map[string]interface{})
	vars["meter"] = meterName

	buf := new(bytes.Buffer)
	tmpl.Execute(buf, vars)
	return buf.String()
}

func exit(client influxdb2.Client) {
	client.Close()
	os.Exit(1)
}
