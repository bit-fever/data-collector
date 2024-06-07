//=============================================================================
/*
Copyright © 2024 Andrea Carboni andrea.carboni71@gmail.com

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

package service

import (
	//"context"
	//"fmt"
	//"github.com/bit-fever/data-collector/pkg/model"
	//"github.com/bit-fever/data-collector/pkg/model/config"
	//"github.com/bit-fever/data-collector/pkg/model/config/data"
	//influx "github.com/influxdata/influxdb-client-go/v2"
	//"github.com/spf13/viper"
	//"net/http"
	"github.com/bit-fever/core/auth"
)

//=============================================================================

func getData(c *auth.Context) {
	//dataMap := viper.GetStringMapString(config.Data)
	//
	//org := dataMap[data.Org]
	//url := dataMap[data.Url]
	//bucket := dataMap[data.Bucket]
	//token := dataMap[data.Token]
	//
	//queryAPI := client.QueryAPI(org)
	//query := fmt.Sprintf(`from(bucket: "%v") |> range(start: -1d)`, bucket)
	//result, err := queryAPI.Query(context.Background(), query)
	//if err != nil {
	//	panic(err)
	//}
	//for result.Next() {
	//	record := result.Record()
	//	fmt.Printf("%v %v: %v=%v\n", record.Time(), record.Measurement(), record.Field(), record.Value())
	//}
	//client.Close()
	//
	//c.IndentedJSON(http.StatusOK, []model.SymbolData{})
}

//=============================================================================
