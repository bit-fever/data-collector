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
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
	"time"
)

//=============================================================================

func GetBiasSummaryInfo(tx *gorm.DB, c *auth.Context, id uint) (*BiasSummaryResponse, error) {
	c.Log.Info("GetBiasSummary: Getting bias analysis", "id", id)

	ba,err := db.GetBiasAnalysisById(tx, id)
	if err != nil {
		return nil, err
	}

	var config *ds.DataConfig
	config, err = CreateDataConfig(tx, ba.DataInstrumentId)
	if err != nil {
		return nil, err
	}

	var bp *db.BrokerProduct
	bp, err = db.GetBrokerProductById(tx, ba.BrokerProductId)
	if err != nil {
		return nil, err
	}

	c.Log.Info("GetBiasSummary: Found bias analysis", "id", id, "name", ba.Name)

	return &BiasSummaryResponse{
		BiasAnalysis : ba,
		BrokerProduct: bp,
		Result       : [7]*DataPointDowList{},
		config       : config,
	}, nil
}

//=============================================================================

func GetBiasSummaryData(c *auth.Context, id uint, bsr *BiasSummaryResponse) error {
	bsr.config.Timeframe = "15m"

	da   := ds.NewDataAggregator(ds.TimeSlotFunction30m)
	loc,_:= time.LoadLocation(bsr.config.Timezone)
	err  := ds.GetDataPoints(DefaultFrom, DefaultTo, bsr.config, loc, da)

	if err != nil {
		return err
	}

	dataPoints := da.DataPoints()

	for i, dpCurr := range dataPoints {
		if i>0 {
			dpPrev  := dataPoints[i -1]
			dpDelta := newDataPointDelta(dpPrev, dpCurr)
			bsr.Add(dpDelta)
		}
	}

	return nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func newDataPointDelta(dpPrev, dpCurr *ds.DataPoint) *DataPointDelta {
	delta := dpCurr.Close - dpPrev.Close

	//--- Calc slot time from destination to take into account leaps when markets
	//--- are closed (i.e. slot 16:00 - 17:30 will have 16:00 instead of 17:00)

	slotTime := dpCurr.Time.Add(-time.Minute * 30)

	y,m,d := slotTime.Date()
	hour  := slotTime.Hour()
	mins  := slotTime.Minute()
	dow   := slotTime.Weekday()

	return &DataPointDelta{
		Year : y,
		Month: int(m),
		Day  : d,
		Hour : hour,
		Min  : mins,
		Delta: delta,
		Dow  : int(dow),
	}
}

//=============================================================================

type DataPointDelta struct {
	Year  int
	Month int
	Day   int
	Hour  int
	Min   int
	Delta float64
	Dow   int
}

//=============================================================================
