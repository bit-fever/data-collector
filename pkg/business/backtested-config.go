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
	"math"
	"time"
)

//=============================================================================

const SlotNum int = 48*5

//=============================================================================
//===
//=== BacktestedConfig
//===
//=============================================================================

type BacktestedConfig struct {
	BiasConfig    *BiasConfig           `json:"biasConfig"`
	GrossProfit   float64               `json:"grossProfit"`
	NetProfit     float64               `json:"netProfit"`
	GrossAvgTrade float64               `json:"grossAvgTrade"`
	NetAvgTrade   float64               `json:"netAvgTrade"`
	Trades        []*BiasTrade          `json:"biasTrades"`
	Sequences     []*TriggeringSequence `json:"sequences"`
	Equity        *Equity               `json:"equity"`
	ProfitDistrib *ProfitDistribution   `json:"profitDistrib"`

	//------------------------

	currTrade     *BiasTrade
	excludedSet   *ExcludedSet
	brokerProduct *db.BrokerProduct
	spec          *BiasBacktestSpec
}

//=============================================================================
//=== Constructor
//=============================================================================

func NewBacktestedConfig(bc *BiasConfig, bp *db.BrokerProduct, spec *BiasBacktestSpec) (*BacktestedConfig, error) {
	es, err := NewExcludedSet(bc.Excludes)

	return &BacktestedConfig{
		BiasConfig   : bc,
		Trades       : []*BiasTrade{},
		Sequences    : []*TriggeringSequence{},
		excludedSet  : es,
		brokerProduct: bp,
		spec         : spec,
	}, err
}

//=============================================================================
//=== Methods
//=============================================================================

func (btc *BacktestedConfig) RunBacktest(ti *TimeInfo, currDp *ds.DataPoint, prevDp *ds.DataPoint, i int, dataPoints []*ds.DataPoint) {
	if btc.currTrade == nil {
		//--- Check if we can start a new trade

		if btc.IsDayAllowed(ti) {
			if btc.IsStartOfTrade(ti) {
				btc.StartTrade(currDp, prevDp, i, dataPoints)
			}
		}
	}

	//--- The previous block of code could have started a new 1-slot trade

	if btc.currTrade != nil {
		//--- Check if we need to exit from current trade

		if btc.IsEndOfTrade(ti) {
			btc.EndTrade(currDp, ExitConditionNormal)
		} else if btc.currTrade.IsInStopLoss(currDp) {
			btc.EndTrade(currDp, ExitConditionStop)
		} else if btc.currTrade.IsInProfit(currDp) {
			btc.EndTrade(currDp, ExitConditionProfit)
		}
	}
}

//=============================================================================

func (btc *BacktestedConfig) IsDayAllowed(ti *TimeInfo) bool {
	return btc.BiasConfig.Months[ti.month -1] && !btc.excludedSet.ShouldBeExcluded(ti.month, ti.year)
}

//=============================================================================

func (btc *BacktestedConfig) IsStartOfTrade(ti *TimeInfo) bool {
	return ti.dayOfWeek == btc.BiasConfig.StartDay && ti.slot == btc.BiasConfig.StartSlot
}

//=============================================================================

func (btc *BacktestedConfig) IsEndOfTrade(ti *TimeInfo) bool {
	return ti.dayOfWeek == btc.BiasConfig.EndDay && ti.slot == btc.BiasConfig.EndSlot
}

//=============================================================================

func (btc *BacktestedConfig) StartTrade(currDp, prevDp *ds.DataPoint, index int, dataPoints []*ds.DataPoint) {
	btc.Sequences = append(btc.Sequences, NewTriggeringSequence(index, dataPoints, SlotNum))
	btc.currTrade = NewBiasTrade(currDp, prevDp, btc)
}

//=============================================================================

func (btc *BacktestedConfig) EndTrade(currDp *ds.DataPoint, exitCondition int8) {
	btc.currTrade.Close(currDp, btc.brokerProduct, exitCondition)

	btc.GrossProfit += btc.currTrade.GrossProfit
	btc.NetProfit   += btc.currTrade.NetProfit

	btc.Trades    = append(btc.Trades, btc.currTrade)
	btc.currTrade = nil
}

//=============================================================================

func (btc *BacktestedConfig) Finish() {
	numTrades := len(btc.Trades)

	if numTrades > 0 {
		btc.GrossAvgTrade = core.Trunc2d(btc.GrossProfit / float64(numTrades))
		btc.NetAvgTrade   = core.Trunc2d(btc.NetProfit   / float64(numTrades))
	}

	btc.GrossProfit = math.Trunc(btc.GrossProfit)
	btc.NetProfit   = math.Trunc(btc.NetProfit)

	btc.Equity        = NewEquity(btc.Trades)
	btc.ProfitDistrib = NewProfitDistribution(btc.Trades, float64(btc.brokerProduct.CostPerTrade))
}

//=============================================================================
//===
//=== TimeInfo
//===
//=============================================================================

type TimeInfo struct {
	dayOfWeek  int16
	slot       int16
	month      int16
	year       int16
}

//=============================================================================
//===
//=== TriggeringSequence
//===
//=============================================================================

type TriggeringSequence struct {
	DataPoints []*ds.DataPoint
}

//=============================================================================

func NewTriggeringSequence(index int, dataPoints []*ds.DataPoint, slotNum int) *TriggeringSequence {
	i := index - slotNum
	if i <0 {
		return nil
	}

	//index--

	var points []*ds.DataPoint

	for ; i<index; i++ {
		points = append(points, dataPoints[i])
	}

	return &TriggeringSequence{
		DataPoints: points,
	}
}

//=============================================================================
//===
//=== Equity
//===
//=============================================================================

type Equity struct {
	Time  []time.Time `json:"time"`
	Gross []float64   `json:"gross"`
	Net   []float64   `json:"net"`
}

//=============================================================================

func NewEquity(trades []*BiasTrade) *Equity {
	var timeArray   []time.Time
	var grossEquity []float64
	var netEquity   []float64

	var grossPrev float64 = 0
	var netPrev   float64 = 0

	for _, bt := range trades {
		timeArray = append(timeArray, bt.ExitTime)

		grossEquity = append(grossEquity, grossPrev + bt.GrossProfit)
		netEquity   = append(netEquity,   netPrev   + bt.NetProfit)

		grossPrev = grossPrev + bt.GrossProfit
		netPrev   = netPrev   + bt.NetProfit
	}

	//--- Truncate decimal digits on profits

	for i, _ := range timeArray {
		grossEquity[i] = math.Trunc(grossEquity[i])
		netEquity[i]   = math.Trunc(netEquity  [i])
	}

	return &Equity{
		Time : timeArray,
		Gross: grossEquity,
		Net  : netEquity,
	}
}

//=============================================================================
//===
//=== ProfitDistribution
//===
//=============================================================================

type ProfitDistribution struct {
	NetProfits [16]float64 `json:"netProfits"`
	NumTrades  [16]int     `json:"numTrades"`
	AvgTrades  [16]float64 `json:"avgTrades"`
}

//=============================================================================

func NewProfitDistribution(trades []*BiasTrade, costPerTrade float64) *ProfitDistribution {
	var netProfits[16]float64
	var numTrades [16]int
	var avgTrades [16]float64

	threshold := calcThreshold(costPerTrade)

	for _, bt := range trades {
		found := false

		for i:=0; i<15; i++ {
			if bt.NetProfit < threshold[i] {
				netProfits[i] += bt.NetProfit
				numTrades [i] += 1
				found = true
				break
			}
		}

		if ! found {
			netProfits[15] += bt.NetProfit
			numTrades [15] += 1
		}
	}

	//---trunc profits and calc avgTrade to 2 decimal digits

	for i:=0; i<16; i++ {
		if numTrades[i] != 0 {
			avgTrades[i] = core.Trunc2d(netProfits[i] / float64(numTrades[i]))
		}

		netProfits[i] = math.Round(netProfits[i])
	}

	return &ProfitDistribution{
		NetProfits: netProfits,
		NumTrades : numTrades,
		AvgTrades : avgTrades,
	}
}

//=============================================================================

func calcThreshold(costPerTrade float64) []float64 {
	var threshold []float64

	for i:=6; i>=0; i-- {
		threshold = append(threshold, -math.Pow(2, float64(i)) * costPerTrade)
	}

	threshold = append(threshold, 0)

	for i:=0; i<=6; i++ {
		threshold = append(threshold, math.Pow(2, float64(i)) * costPerTrade)
	}

	return threshold
}

//=============================================================================
