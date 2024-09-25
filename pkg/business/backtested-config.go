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
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"time"
)

//=============================================================================

const SlotNum int = 48*5

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
//=== BiasTrade
//===
//=============================================================================

type BiasTrade struct {
	EntryTime     time.Time `json:"entryTime"`
	EntryValue    float64   `json:"entryValue"`
	ExitTime      time.Time `json:"exitTime"`
	ExitValue     float64   `json:"exitValue"`
	Operation     int8      `json:"operation"`
	GrossProfit   float64   `json:"grossProfit"`
	NetProfit     float64   `json:"netProfit"`
}

//=============================================================================

func (bt *BiasTrade) Close(dp *ds.DataPoint, bp *db.BrokerProduct) {
	bt.ExitTime   = dp.Time
	bt.ExitValue  = dp.Close
	bt.GrossProfit = (bt.ExitValue - bt.EntryValue) * float64(bp.PointValue)

	if bt.Operation == 1 {
		bt.GrossProfit *= -1
	}

	//--- We have 2 trades: 1 to enter and 1 to exit the market
	bt.NetProfit = bt.GrossProfit - 2 * float64(bp.CostPerTrade)
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

	//------------------------

	currTrade     *BiasTrade
	excludedSet   *ExcludedSet
	brokerProduct *db.BrokerProduct
}

//=============================================================================

func NewBacktestedConfig(bc *BiasConfig, bp *db.BrokerProduct) (*BacktestedConfig, error) {
	es, err := NewExcludedSet(bc.Excludes)

	return &BacktestedConfig{
		BiasConfig   : bc,
		Trades       : []*BiasTrade{},
		Sequences    : []*TriggeringSequence{},
		excludedSet  : es,
		brokerProduct: bp,
	}, err
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
	btc.currTrade = &BiasTrade{
		EntryTime : currDp.Time.Add(-time.Minute * 30),
		EntryValue: prevDp.Close,
		Operation: btc.BiasConfig.Operation,
	}
}

//=============================================================================

func (btc *BacktestedConfig) EndTrade(currDp *ds.DataPoint) {
	btc.currTrade.Close(currDp, btc.brokerProduct)

	btc.GrossProfit += btc.currTrade.GrossProfit
	btc.NetProfit   += btc.currTrade.NetProfit

	btc.Trades    = append(btc.Trades, btc.currTrade)
	btc.currTrade = nil
}

//=============================================================================

func (btc *BacktestedConfig) Finish() {
	numTrades := len(btc.Trades)

	if numTrades > 0 {
		btc.GrossAvgTrade = btc.GrossProfit / float64(numTrades)
		btc.NetAvgTrade   = btc.NetProfit   / float64(numTrades)
	}
}

//=============================================================================
