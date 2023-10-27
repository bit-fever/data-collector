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
	"github.com/bit-fever/data-collector/pkg/model/config"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"log"
)

//=============================================================================
var client   influxdb2.Client
var queryAPI api.QueryAPI
var writeAPI api.WriteAPI

//=============================================================================

func InitDatastore(cfg *config.Config) {

	log.Println("Starting datastore...")

	client   = influxdb2.NewClient(cfg.Data.Url, cfg.Data.Token)
	queryAPI = client.QueryAPI(cfg.Data.Org)
	writeAPI = client.WriteAPI(cfg.Data.Org, cfg.Data.Bucket)
}

//=============================================================================

//for value := 0; value < 5; value++ {
//tags := map[string]string{
//"tagname1": "tagvalue1",
//}
//fields := map[string]interface{}{
//"field1": value,
//}
//point := write.NewPoint("measurement1", tags, fields, time.Now())
//time.Sleep(1 * time.Second) // separate points by 1 second
//
//if err := writeAPI.WritePoint(context.Background(), point); err != nil {
//log.Fatal(err)
//}
//}





//queryAPI := client.QueryAPI(org)
//query := `from(bucket: "<BUCKET>")
//            |> range(start: -10m)
//            |> filter(fn: (r) => r._measurement == "measurement1")`
//results, err := queryAPI.Query(context.Background(), query)
//if err != nil {
//log.Fatal(err)
//}
//for results.Next() {
//fmt.Println(results.Record())
//}
//if err := results.Err(); err != nil {
//log.Fatal(err)
//}





//query = `from(bucket: "<BUCKET>")
//              |> range(start: -10m)
//              |> filter(fn: (r) => r._measurement == "measurement1")
//              |> mean()`
//results, err = queryAPI.Query(context.Background(), query)
//if err != nil {
//log.Fatal(err)
//}
//for results.Next() {
//fmt.Println(results.Record())
//}
//if err := results.Err(); err != nil {
//log.Fatal(err)
//}
