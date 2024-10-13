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
	"github.com/bit-fever/data-collector/pkg/core"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"strconv"
	"time"
)

//=============================================================================
//===
//=== BiasTrade
//===
//=============================================================================

const (
	ExitConditionNormal    =  0
	ExitConditionStop      = -1
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

	stopValue     float64
	profitValue   float64
}

//=============================================================================

func NewBiasTrade(currDp, prevDp *ds.DataPoint, btc *BacktestedConfig) *BiasTrade {
	entryValue  := prevDp.Close
	stopValue   := 0.0
	profitValue := 0.0

	stopDelta   := btc.spec.StopLoss   / float64(btc.brokerProduct.PointValue)
	profitDelta := btc.spec.TakeProfit / float64(btc.brokerProduct.PointValue)

	switch btc.BiasConfig.Operation {
		case 0:
			stopValue   = entryValue - stopDelta
			profitValue = entryValue + profitDelta
		case 1:
			stopValue   = entryValue + stopDelta
			profitValue = entryValue - profitDelta

		default:
			panic("Unknown trade operation: "+ strconv.Itoa(int(btc.currTrade.Operation)))
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
			case 0: return currDp.Low  <= bt.stopValue
			case 1: return currDp.High >= bt.stopValue

			default: panic("Unknown trade operation: "+ strconv.Itoa(int(bt.Operation)))
		}
	}

	return false
}

//=============================================================================

func (bt *BiasTrade) IsInProfit(currDp *ds.DataPoint) bool {
	if bt.profitValue != 0 {
		switch bt.Operation {
			case 0: return currDp.High >= bt.profitValue
			case 1: return currDp.Low  <= bt.profitValue

			default: panic("Unknown trade operation: "+ strconv.Itoa(int(bt.Operation)))
		}
	}

	return false
}

//=============================================================================

func (bt *BiasTrade) Close(dp *ds.DataPoint, bp *db.BrokerProduct, exitCondition int8) {
	var exitValue float64

	switch exitCondition {
		case ExitConditionNormal: exitValue = dp.Close
		case ExitConditionStop:   exitValue = bt.stopValue
		case ExitConditionProfit: exitValue = bt.profitValue
	}

	bt.ExitTime   = dp.Time
	bt.ExitValue  = exitValue

	bt.GrossProfit = (bt.ExitValue - bt.EntryValue) * float64(bp.PointValue)

	if bt.Operation == 1 {
		bt.GrossProfit *= -1
	}

	//--- We have 2 trades: 1 to enter and 1 to exit the market
	bt.NetProfit   = core.Trunc2d(bt.GrossProfit - 2 * float64(bp.CostPerTrade))
	bt.GrossProfit = core.Trunc2d(bt.GrossProfit)
}

//=============================================================================
