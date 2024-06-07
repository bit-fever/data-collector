//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package ds

import (
	"context"
	"fmt"
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"log"
	"time"
)

//=============================================================================

//--- Fields

const Open   = "open"
const Close  = "close"
const High   = "high"
const Low    = "low"
const Volume = "volume"

//--- Tags

const ConnCode   = "connCode"
const SystemCode = "sysCode"
const User       = "user"

//--- Variables

var client   influxdb2.Client
var queryAPI api.QueryAPI
var writeAPI api.WriteAPIBlocking

//=============================================================================

func InitDatastore(data *app.Data) {

	log.Println("Starting datastore...")

	client   = influxdb2.NewClient(data.Url, data.Token)
	queryAPI = client.QueryAPI(data.Org)
	writeAPI = client.WriteAPIBlocking(data.Org, data.Bucket)
}

//=============================================================================

func LoadData() {


	query := `from(bucket: "symbol-data")
            |> range(start: -10m)
            |> filter(fn: (r) => r._measurement == "measurement1")`
	results, err := queryAPI.Query(context.Background(), query)

	if err != nil {
		log.Fatal(err)
	}

	for results.Next() {
		fmt.Println(results.Record())
	}

	if err := results.Err(); err != nil {
		log.Fatal(err)
	}
}

//=============================================================================

func WriteData(symbol string, timestamp time.Time, open, close, high, low float64, volume int,
				connCode, systemCode, user string) error {

	tags := map[string]string{
		ConnCode  : connCode,
		SystemCode: systemCode,
		User      : user,
	}

	fields := map[string]interface{}{
		Open  : open,
		Close : close,
		High  : high,
		Low   : low,
		Volume: volume,
	}

	point := write.NewPoint(symbol, tags, fields, timestamp)
	err   := writeAPI.WritePoint(context.Background(), point)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

//=============================================================================
