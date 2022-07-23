package main

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	url := "https://us-west-2-1.aws.cloud2.influxdata.com"
	token := "my-token"
	org := "my-org"
	bucket := "my-bucket"

	client := influxdb2.NewClient(url, token)
	queryAPI := client.QueryAPI(org)
	query := fmt.Sprintf(`from(bucket: "%v") |> range(start: -1d)`, bucket)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}
	for result.Next() {
		record := result.Record()
		fmt.Printf("%v %v: %v=%v\n", record.Time(), record.Measurement(), record.Field(), record.Value())
	}
	client.Close()
}
