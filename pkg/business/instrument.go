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
	"math"
	"strconv"
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

func GetInstrumentDataById(c *auth.Context, spec *InstrumentDataSpec)(*InstrumentDataResponse, error) {
	params,err := parseInstrumentDataParams(spec)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	dataPoints,err := ds.GetDataPoints(params.From, params.To, spec.Config, params.Location)
	dur := time.Now().Sub(start).Seconds()
	if err != nil {
		return nil, err
	}

	c.Log.Info("GetInstrumentDataById: Query stats", "duration", dur, "records", len(dataPoints))

	reduced := false
	dataPoints,reduced = reduceDataPoints(dataPoints, params.Reduction)

	return &InstrumentDataResponse{
		Id         : spec.Id,
		Symbol     : spec.Config.Symbol,
		From       : params.From.Format(time.DateTime),
		To         : params.To.Format(time.DateTime),
		Timeframe  : spec.Config.Timeframe,
		Timezone   : params.Location.String(),
		Reduction  : params.Reduction,
		Reduced    : reduced,
		Records    : len(dataPoints),
		DataPoints : dataPoints,
	}, nil
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

func parseInstrumentDataParams(spec *InstrumentDataSpec) (*InstrumentDataParams, error) {
	loc,err := getLocation(spec.Timezone, spec.Config)
	if err != nil {
		return nil, errors.New("Bad timezone: "+ spec.Timezone +" ("+ err.Error() +")")
	}

	from,err1 := parseTime(spec.From, DefaultFrom, loc)
	to,  err2 := parseTime(spec.To,   DefaultTo,   loc)

	if err1 != nil {
		return nil, errors.New("Bad 'from' parameter: "+ spec.From +" ("+ err1.Error() +")")
	}

	if err2 != nil {
		return nil, errors.New("Bad 'to' parameter: "+ spec.To +" ("+ err2.Error() +")")
	}

	if err = checkTimeframe(spec.Config.Timeframe); err != nil {
		return nil, errors.New("Bad timeframe: "+ spec.Config.Timeframe +" ("+ err.Error() +")")
	}

	red, err := parseReduction(spec.Reduction)

	if err != nil {
		return nil, errors.New("Bad reduction: "+ spec.Reduction +" ("+ err.Error() +")")
	}

	return &InstrumentDataParams{
		Location : loc,
		From     : from,
		To       : to,
		Reduction: red,
	}, nil
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

func parseReduction(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	red, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	if red < 100 || red > 100000 {
		return 0, errors.New("allowed range is 100..100000")
	}

	return red,nil
}

//=============================================================================

func reduceDataPoints(dataPoints []*ds.DataPoint, reduction int) ([]*ds.DataPoint, bool) {
	if reduction == 0 || len(dataPoints) <= reduction {
		return dataPoints, false
	}

	shrinkSize := len(dataPoints) / reduction +1

	var list []*ds.DataPoint
	var currDp *ds.DataPoint = nil
	var count = 0

	for _,dp := range dataPoints {
		if currDp == nil {
			currDp = dp
		} else {
			currDp.High    = math.Max(currDp.High, dp.High)
			currDp.Low     = math.Min(currDp.Low,  dp.Low)
			currDp.Close   = dp.Close
			currDp.Volume += dp.Volume
		}

		count++
		if count == shrinkSize {
			list   = append(list, currDp)
			currDp = nil
			count  = 0
		}
	}

	return list, true
}

//=============================================================================
