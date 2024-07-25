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

package business

import (
	"errors"
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
	"time"
)

//=============================================================================

var DefaultFrom = time.Date(2000,1,1,0,0,0,0, time.UTC)
var DefaultTo   = time.Date(3000,1,1,0,0,0,0, time.UTC)

//=============================================================================

func CreateDataConfig(tx *gorm.DB, id uint) (*ds.DataConfig, error) {
	var p *db.Product

	i, err := db.GetInstrumentById(tx, id)
	if err == nil {
		p, err = db.GetProductById(tx, i.ProductId)
		if err == nil {
			return createConfig(i, p), nil
		}
	}

	return nil, err
}

//=============================================================================

func GetInstrumentDataById(c *auth.Context, from string, to string, timezone string, config *ds.DataConfig)([]*ds.DataPoint, error) {

	loc,err := getLocation(timezone, config)
	if err != nil {
		return nil, errors.New("Bad timezone ("+ err.Error() +")")
	}

	fromTime,err1 := parseTime(from, DefaultFrom, loc)
	toTime,  err2 := parseTime(to,   DefaultTo,   loc)

	if err1 != nil {
		return nil, errors.New("Bad 'from' parameter: "+ from +" ("+ err1.Error() +")")
	}

	if err2 != nil {
		return nil, errors.New("Bad 'to' parameter: "+ to +" ("+ err2.Error() +")")
	}

	if err := checkTimeframe(config.Timeframe); err != nil {
		return nil, errors.New("Bad timeframe: "+ config.Timeframe +" ("+ err.Error() +")")
	}

	start := time.Now()
	dataPoints,err := ds.GetDataPoints(fromTime, toTime, config, loc)
	dur := time.Now().Sub(start).Seconds()

	if err == nil {
		c.Log.Info("GetInstrumentDataById: Query stats", "duration", dur, "records", len(dataPoints))
	}

	return dataPoints, err
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func createConfig(i *db.Instrument, p *db.Product) *ds.DataConfig {
	var selector  any
	var userTable bool

	if p.SupportsMultipleData {
		userTable = true
		selector  = i.Id
	} else {
		userTable = false
		selector  = p.SystemCode
	}

	return &ds.DataConfig{
		UserTable: userTable,
		Timeframe: "1m",
		Selector : selector,
		Symbol   : i.Symbol,
		Timezone : p.Timezone,
	}
}

//=============================================================================

func getLocation(timezone string, config *ds.DataConfig) (*time.Location, error) {
	if timezone == "exchange" {
		timezone = config.Timezone
	}

	return time.LoadLocation(timezone)
}

//=============================================================================

func parseTime(t string, defValue time.Time, loc *time.Location) (time.Time, error) {
	if len(t) == 0 {
		return defValue, nil
	}

	return time.ParseInLocation(time.DateTime, t, loc)
}

//=============================================================================

func checkTimeframe(tf string) error {
	if tf=="1m" || tf=="5m" || tf=="15m" || tf=="60m" {
		return nil
	}

	return errors.New("allowed values are 1m, 5m, 15m, 60m")
}

//=============================================================================
