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

	"github.com/algotiqa/data-collector/pkg/core"
	"github.com/algotiqa/data-collector/pkg/db"
	"github.com/algotiqa/data-collector/pkg/ds"
)

//=============================================================================
//===
//=== BiasTrade
//===
//=============================================================================

const (
	ExitConditionNormal = 0
	ExitConditionStop   = -1
	ExitConditionProfit = +1
)

//-----------------------------------------------------------------------------

type BiasTrade struct {
	EntryTime     time.Time `json:"entryTime"`
	EntryValue    float64   `json:"entryValue"`
	ExitTime      time.Time `json:"exitTime"`
	ExitValue     float64   `json:"exitValue"`
	Operation     int8      `json:"operation"`
	GrossProfit   float64   `json:"grossProfit"`
	NetProfit     float64   `json:"netProfit"`
	ExitCondition int8      `json:"exitCondition"`

	stopValue   float64
	profitValue float64
}

//=============================================================================

func NewBiasTrade(currDp, prevDp *ds.DataPoint, btc *BacktestedConfig, spec *BiasBacktestSpec) *BiasTrade {
	entryValue  := prevDp.Close
	stopValue   := 0.0
	profitValue := 0.0

	stopDelta   := spec.StopLoss / float64(btc.brokerProduct.PointValue)
	profitDelta := spec.TakeProfit / float64(btc.brokerProduct.PointValue)

	switch btc.BiasConfig.Operation {
	case 0:
		stopValue   = entryValue - stopDelta
		profitValue = entryValue + profitDelta
	case 1:
		stopValue   = entryValue + stopDelta
		profitValue = entryValue - profitDelta

	default:
		panic("Unknown trade operation: " + strconv.Itoa(int(btc.currTrade.Operation)))
	}

	if stopDelta == 0 {
		stopValue = 0
	}

	if profitDelta == 0 {
		profitValue = 0
	}

	bt := &BiasTrade{
		EntryTime  : currDp.Time.Add(-time.Minute * 30),
		EntryValue : entryValue,
		Operation  : btc.BiasConfig.Operation,
		stopValue  : stopValue,
		profitValue: profitValue,
	}

	return bt
}

//=============================================================================

func (bt *BiasTrade) IsInStopLoss(currDp *ds.DataPoint) bool {
	if bt.stopValue != 0 {
		switch bt.Operation {
		case 0:
			return currDp.Low <= bt.stopValue
		case 1:
			return currDp.High >= bt.stopValue

		default:
			panic("Unknown trade operation: " + strconv.Itoa(int(bt.Operation)))
		}
	}

	return false
}

//=============================================================================

func (bt *BiasTrade) IsInProfit(currDp *ds.DataPoint) bool {
	if bt.profitValue != 0 {
		switch bt.Operation {
		case 0:
			return currDp.High >= bt.profitValue
		case 1:
			return currDp.Low <= bt.profitValue

		default:
			panic("Unknown trade operation: " + strconv.Itoa(int(bt.Operation)))
		}
	}

	return false
}

//=============================================================================

func (bt *BiasTrade) Close(dp *ds.DataPoint, bp *db.BrokerProduct, exitCondition int8) {
	var exitValue float64

	switch exitCondition {
	case ExitConditionNormal:
		exitValue = dp.Close
	case ExitConditionStop:
		exitValue = bt.stopValue
	case ExitConditionProfit:
		exitValue = bt.profitValue
	}

	bt.ExitTime = dp.Time
	bt.ExitValue = exitValue

	bt.GrossProfit = (bt.ExitValue - bt.EntryValue) * float64(bp.PointValue)

	if bt.Operation == 1 {
		bt.GrossProfit *= -1
	}

	//--- We have 2 trades: 1 to enter and 1 to exit the market
	bt.NetProfit = core.Trunc2d(bt.GrossProfit - 2*float64(bp.CostPerOperation))
	bt.GrossProfit = core.Trunc2d(bt.GrossProfit)
}

//=============================================================================
