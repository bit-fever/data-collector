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
	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

func getInstrumentData(c *auth.Context) {
	id, err   := c.GetIdFromUrl()
	from      := c.GetParamAsString("from",      "")
	to        := c.GetParamAsString("to",        "")
	timeframe := c.GetParamAsString("timeframe", "5m")
	timezone  := c.GetParamAsString("timezone",  "UTC")

	var result []*ds.DataPoint
	var config *ds.DataConfig

	if err == nil {
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			cfg, err := business.CreateDataConfig(tx, id)
			config = cfg
			return err
		})

		if err == nil {
			config.Timeframe = timeframe
			result, err = business.GetInstrumentDataById(c, from, to, timezone, config)
			if err == nil {
				_=c.ReturnObject(result)
				return
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================
