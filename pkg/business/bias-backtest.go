//=============================================================================
//===
//=== Copyright (C) 2024-present Andrea Carboni
//===
//=== This source code is licensed under the Elastic License 2.0 (ELv2) available at:
//=== https://github.com/algotiqa/docs/blob/main/LICENSE.md
//=== By using this file, you agree to the terms and conditions of that license.
//=============================================================================


package business

import (
	"strconv"
	"time"

	"github.com/algotiqa/core/auth"
	"github.com/algotiqa/core/req"
	"github.com/algotiqa/data-collector/pkg/core"
	"github.com/algotiqa/data-collector/pkg/db"
	"github.com/algotiqa/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================
//===
//=== Structures
//===
//=============================================================================

type BiasBacktestSpec struct {
	StopLoss   float64
	TakeProfit float64
}

//=============================================================================

type BiasBacktestResponse struct {
	BiasAnalysis      *db.BiasAnalysis    `json:"biasAnalysis"`
	BrokerProduct     *db.BrokerProduct   `json:"brokerProduct"`
	Spec              *BiasBacktestSpec   `json:"spec"`
	BacktestedConfigs []*BacktestedConfig `json:"backtestedConfigs"`
}

//=============================================================================
//===
//=== Functions
//===
//=============================================================================

func GetBacktestInfo(tx *gorm.DB, c *auth.Context, id uint, sessionConfig string) (*BiasBacktestResponse, *core.QueryConfig, error) {
	c.Log.Info("GetBacktestInfo: Getting bias analysis and configs for backtest", "id", id)

	ba, err := getBiasAnalysisAndCheckAccess(tx, c, id, "GetBacktestInfo")
	if err != nil {
		return nil, nil, err
	}

	biasConfigs, err2 := GetBiasConfigsByAnalysisId(tx, c, id)
	if err2 != nil {
		c.Log.Error("GetBacktestInfo: Could not retrieve bias configs", "error", err.Error())
		return nil, nil, err2
	}

	var config *core.QueryConfig
	config, err = CreateQueryConfig(tx, ba.DataInstrumentId, sessionConfig)
	if err != nil {
		c.Log.Error("GetBacktestInfo: Could not create data config", "error", err.Error())
		return nil, nil, err
	}

	var bp *db.BrokerProduct
	bp, err = db.GetBrokerProductById(tx, ba.BrokerProductId)
	if err != nil {
		c.Log.Error("GetBacktestInfo: Could not retrieve broker product", "error", err.Error())
		return nil, nil, err
	}

	var btConfigs []*BacktestedConfig

	for _, bc := range *biasConfigs {
		btc, err := NewBacktestedConfig(bc, bp)
		if err != nil {
			c.Log.Error("GetBacktestInfo: Could not build backtested config", "error", err.Error())
			return nil, nil, err
		}

		btConfigs = append(btConfigs, btc)
	}

	spec, err := createSpec(c)
	if err != nil {
		return nil, nil, err
	}

	return &BiasBacktestResponse{
		BiasAnalysis     : ba,
		BrokerProduct    : bp,
		BacktestedConfigs: btConfigs,
		Spec             : spec,
	}, config, nil
}

//=============================================================================

func RunBacktest(c *auth.Context, spec *QuerySpec, bbr *BiasBacktestResponse) error {
	c.Log.Info("RunBacktest: Starting backtest for bias analysis", "id", bbr.BiasAnalysis.Id)

	spec.Timeframe = "30"
	spec.Timezone  = ""

	params, err := NewQueryParams(spec)
	if err != nil {
		return req.NewBadRequestError(err.Error())
	}

	params.Aggregator = ds.NewSimpleAggregator(ds.NewQuantizer15mTo30m())
	params.Reduction  = 0
	params.Limit      = 0

	dataPoints, err := getDataPoints(params, spec.Config)
	if err != nil {
		return err
	}

	for i, dp := range dataPoints {
		if i > 0 {
			prevDp := dataPoints[i-1]
			ti := calcTimeInfo(dp)

			for _, btc := range bbr.BacktestedConfigs {
				btc.RunBacktest(ti, dp, prevDp, i, dataPoints, bbr.Spec)
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

func createSpec(c *auth.Context) (*BiasBacktestSpec, error) {
	stopLoss  ,errL := getMandatoryValue(c,"stopLoss")
	takeProfit,errP := getMandatoryValue(c,"takeProfit")
	if errL != nil {
		return nil,errL
	}
	if errP != nil {
		return nil,errP
	}

	if stopLoss < 0 {
		return nil,req.NewBadRequestError("stopLoss cannot be negative: %v", stopLoss)
	}

	if takeProfit < 0 {
		return nil,req.NewBadRequestError("takeProfit cannot be negative: %v", takeProfit)
	}

	return &BiasBacktestSpec{
		StopLoss  : stopLoss,
		TakeProfit: takeProfit,
	}, nil
}

//=============================================================================

func getMandatoryValue(c *auth.Context, name string) (float64,error) {
	sValue := c.GetParamAsString(name, "")
	if sValue == "" {
		return 0, req.NewBadRequestError("Missing '%v' parameter", name)
	}

	return strconv.ParseFloat(sValue, 64)
}

//=============================================================================

func calcTimeInfo(dp *ds.DataPoint) *TimeInfo {
	//--- Calc slot time from destination to take into account leaps when markets
	//--- are closed (i.e. slot 16:00 - 17:30 will have 16:00 instead of 17:00)

	slotTime := dp.Time.Add(-time.Minute * 30)

	year, month, _ := slotTime.Date()
	hour, mins, _ := slotTime.Clock()
	dow := slotTime.Weekday()
	slot := (hour*60 + mins) / 30

	return &TimeInfo{
		dayOfWeek: int16(dow),
		slot     : int16(slot),
		month    : int16(month),
		year     : int16(year),
	}
}

//=============================================================================
