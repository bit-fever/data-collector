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
	"github.com/bit-fever/sick-engine/session"
	"gorm.io/gorm"
	"time"
)

//=============================================================================
//===
//=== Structures
//===
//=============================================================================

type BiasBacktestSpec struct {
	StopLoss   float64                  `json:"stopLoss"`
	TakeProfit float64                  `json:"takeProfit"`
	Session    *session.TradingSession  `json:"session"`
}

//=============================================================================

type BiasBacktestResponse struct {
	BiasAnalysis      *db.BiasAnalysis     `json:"biasAnalysis"`
	BrokerProduct     *db.BrokerProduct    `json:"brokerProduct"`
	Spec              *BiasBacktestSpec    `json:"spec"`
	BacktestedConfigs []*BacktestedConfig  `json:"backtestedConfigs"`
	config            *ds.DataConfig
}

//=============================================================================
//===
//=== Functions
//===
//=============================================================================

func GetBacktestInfo(tx *gorm.DB, c *auth.Context, id uint, spec *BiasBacktestSpec) (*BiasBacktestResponse, error) {
	c.Log.Info("GetBacktestInfo: Getting bias analysis and configs for backtest", "id", id)

	ba, err := getBiasAnalysisAndCheckAccess(tx, c, id, "GetBacktestInfo")
	if err != nil {
		return nil, err
	}

	biasConfigs,err2 := GetBiasConfigsByAnalysisId(tx, c, id)
	if err2 != nil {
		c.Log.Error("GetBacktestInfo: Could not retrieve bias configs", "error", err.Error())
		return nil, err2
	}

	var config *ds.DataConfig
	config, err = CreateDataConfig(tx, ba.DataInstrumentId)
	if err != nil {
		c.Log.Error("GetBacktestInfo: Could not create data config", "error", err.Error())
		return nil, err
	}

	var bp *db.BrokerProduct
	bp, err = db.GetBrokerProductById(tx, ba.BrokerProductId)
	if err != nil {
		c.Log.Error("GetBacktestInfo: Could not retrieve broker product", "error", err.Error())
		return nil, err
	}

	var btConfigs []*BacktestedConfig

	for _, bc := range *biasConfigs {
		btc, err := NewBacktestedConfig(bc, bp, spec)
		if err != nil {
			c.Log.Error("GetBacktestInfo: Could not build backtested config", "error", err.Error())
			return nil, err
		}

		btConfigs = append(btConfigs, btc)
	}

	err = checkSpec(c, spec)
	if err != nil {
		return nil,err
	}

	return &BiasBacktestResponse{
		BiasAnalysis     : ba,
		BrokerProduct    : bp,
		Spec             : spec,
		BacktestedConfigs: btConfigs,
		config           : config,
	}, nil
}

//=============================================================================

func RunBacktest(c *auth.Context, bbr *BiasBacktestResponse) error {
	c.Log.Info("RunBacktest: Starting backtest for bias analysis", "id", bbr.BiasAnalysis.Id)

	bbr.config.Timeframe = "15m"

	da   := ds.NewDataAggregator(ds.TimeSlotFunction30m)
	loc,_:= time.LoadLocation(bbr.config.Timezone)
	err  := ds.GetDataPoints(DefaultFrom, DefaultTo, bbr.config, loc, da)

	if err != nil {
		c.Log.Error("RunBacktest: Could not retrieve data points", "error", err.Error())
		return err
	}

	dataPoints := da.DataPoints()

	for i, dp := range dataPoints {
		if i>0 {
			prevDp := dataPoints[i-1]
			ti     := calcTimeInfo(dp)

			for _, btc := range bbr.BacktestedConfigs {
				btc.RunBacktest(ti, dp, prevDp, i, dataPoints)
			}
		}
	}

	for _, btc := range bbr.BacktestedConfigs {
		btc.Finish()
	}

	return nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func checkSpec(c *auth.Context, bts *BiasBacktestSpec) error {
	var err error

	if bts.StopLoss < 0 {
		err = errors.New("stopLoss cannot be negative")
		c.Log.Error("createParams: Invalid stopLoss", "error", err.Error())
		return err
	}

	if bts.TakeProfit < 0 {
		err = errors.New("takeProfit cannot be negative")
		c.Log.Error("createParams: Invalid takeProfit", "error", err.Error())
		return err
	}

	return nil
}

//=============================================================================

func calcTimeInfo(dp *ds.DataPoint) *TimeInfo {
	//--- Calc slot time from destination to take into account leaps when markets
	//--- are closed (i.e. slot 16:00 - 17:30 will have 16:00 instead of 17:00)

	slotTime := dp.Time.Add(-time.Minute * 30)

	year,month,_ := slotTime.Date()
	hour,mins, _ := slotTime.Clock()
	dow          := slotTime.Weekday()
	slot         := (hour * 60 + mins) / 30

	return &TimeInfo{
		dayOfWeek: int16(dow),
		slot     : int16(slot),
		month    : int16(month),
		year     : int16(year),
	}
}

//=============================================================================
