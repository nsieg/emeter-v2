package main

import (
	"testing"
)

func TestQueryFromTemplate(t *testing.T) {
	type args struct {
		tmplFilePath string
		meterName    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"today-solar",
			args{
				"query_today.tmpl",
				"solar",
			},
			`from(bucket: "energy")
|> range(start: today(), stop: now())
|> filter(fn: (r) => r["meter"] == "solar")
|> map(fn: (r) => ({ r with _unix: uint(v: r._time) / uint(v:1000000000) }))`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QueryFromTemplate(tt.args.tmplFilePath, tt.args.meterName); got != tt.want {
				t.Errorf("got:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}
