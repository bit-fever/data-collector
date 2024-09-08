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
	"time"
)

//=============================================================================
//===
//=== Upload spec & response
//===
//=============================================================================

type DatafileUploadSpec struct {
	Symbol        string `json:"symbol"       binding:"required"`
	Name          string `json:"name"         binding:"required"`
	Continuous    bool   `json:"continuous"   binding:"required"`
	FileTimezone  string `json:"fileTimezone" binding:"required"`
	Parser        string `json:"parser"       binding:"required"`
}

//=============================================================================

type DatafileUploadResponse struct {
	Duration int   `json:"duration"`
	Bytes    int64 `json:"bytes"`
}

//=============================================================================
//=== Get data request & response
//=============================================================================

type DataInstrumentDataSpec struct {
	Id        uint
	From      string
	To        string
	Timezone  string
	Reduction string
	Config    *ds.DataConfig
}

//=============================================================================

type DataInstrumentDataParams struct {
	Location  *time.Location
	From       time.Time
	To         time.Time
	Reduction  int
	Aggregator *ds.DataAggregator
}

//=============================================================================

type DataInstrumentDataResponse struct {
	Id          uint            `json:"id"`
	Symbol      string          `json:"symbol"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Timeframe   string          `json:"timeframe"`
	Timezone    string          `json:"timezone"`
	Reduction   int             `json:"reduction,omitempty"`
	Reduced     bool            `json:"reduced"`
	Records     int             `json:"records"`
	DataPoints  []*ds.DataPoint `json:"dataPoints"`
}

//=============================================================================

type DataInstrumentExt struct {
	db.DataInstrument
}

//=============================================================================
//=== Bias analysis
//=============================================================================

type BiasAnalysisSpec struct {
	DataInstrumentId  uint    `json:"dataInstrumentId"`
	BrokerProductId   uint    `json:"brokerProductId"`
	Name              string  `json:"name"`
	Notes             string  `json:"notes"`
}

//=============================================================================

type BiasAnalysisExt struct {
	db.BiasAnalysis
	DataInstrument db.DataInstrument  `json:"dataInstrument"`
	BrokerProduct  db.BrokerProduct   `json:"brokerProduct"`
	Configs        *[]db.BiasConfig   `json:"configs"`
}

//=============================================================================

type BiasSummaryResponse struct {
	BiasAnalysis  *db.BiasAnalysis     `json:"biasAnalysis"`
	BrokerProduct *db.BrokerProduct    `json:"brokerProduct"`
	Result        [7]*DataPointDowList `json:"result"`
	config        *ds.DataConfig
}

//-----------------------------------------------------------------------------

func (r *BiasSummaryResponse) Add(dpd *DataPointDelta) {
	dpdl := r.Result[dpd.Dow]
	if dpdl == nil {
		dpdl = &DataPointDowList{
			Slots: [48]*DataPointSlotList{},
		}

		r.Result[dpd.Dow] = dpdl
	}

	dpdl.Add(dpd)
}

//=============================================================================

type DataPointDowList struct {
	Slots [48]*DataPointSlotList `json:"slots"`
}

//-----------------------------------------------------------------------------

func (l *DataPointDowList) Add(dpd *DataPointDelta) {
	slot := (dpd.Hour * 60 + dpd.Min) / 30
	dpsl := l.Slots[slot]
	if dpsl == nil {
		dpsl = &DataPointSlotList{
			List: []*DataPointEntry{},
		}

		l.Slots[slot] = dpsl
	}

	l.Slots[slot].Add(dpd)
}

//=============================================================================

type DataPointSlotList struct {
	List []*DataPointEntry `json:"list"`
}

//-----------------------------------------------------------------------------

func (l *DataPointSlotList) Add(dpd *DataPointDelta) {
	dpe := &DataPointEntry{
		Year : int16(dpd.Year),
		Month: int8(dpd.Month),
		Day  : int8(dpd.Day),
		Delta: dpd.Delta,
	}

	l.List = append(l.List, dpe)
}

//=============================================================================

type DataPointEntry struct {
	Year  int16   `json:"year"`
	Month int8    `json:"month"`
	Day   int8    `json:"day"`
	Delta float64 `json:"delta"`
}

//=============================================================================

type BiasConfigSpec struct {
	StartDay        int16    `json:"startDay"`
	StartSlot       int16    `json:"startSlot"`
	EndDay          int16    `json:"endDay"`
	EndSlot         int16    `json:"endSlot"`
	Months          []bool   `json:"months"`
	Excludes        []string `json:"excludes"`
	Operation       int8     `json:"operation"`
	GrossProfit     float64  `json:"grossProfit"`
	NetProfit       float64  `json:"netProfit"`
}

//-----------------------------------------------------------------------------

func (bcs *BiasConfigSpec) ToBiasConfig() *db.BiasConfig {
	var bc db.BiasConfig
	bc.StartDay    = bcs.StartDay
	bc.StartSlot   = bcs.StartSlot
	bc.EndDay      = bcs.EndDay
	bc.EndSlot     = bcs.EndSlot
	bc.Months      = core.EncodeMonths(bcs.Months)
	bc.Excludes    = core.EncodeExcludes(bcs.Excludes)
	bc.Operation   = bcs.Operation
	bc.GrossProfit = bcs.GrossProfit
	bc.NetProfit   = bcs.NetProfit

	return &bc
}

//=============================================================================

type BiasConfig struct {
	BiasConfigSpec
	Id              uint `json:"id"`
	BiasAnalysisId  uint `json:"biasAnalysisId"`
}

//-----------------------------------------------------------------------------

func (bc *BiasConfig) FromBiasConfig(dbc *db.BiasConfig) {
	bc.Id             = dbc.Id
	bc.BiasAnalysisId = dbc.BiasAnalysisId
	bc.StartDay       = dbc.StartDay
	bc.StartSlot      = dbc.StartSlot
	bc.EndDay         = dbc.EndDay
	bc.EndSlot        = dbc.EndSlot
	bc.Months         = core.DecodeMonths(dbc.Months)
	bc.Excludes       = core.DecodeExcludes(dbc.Excludes)
	bc.Operation      = dbc.Operation
	bc.GrossProfit    = dbc.GrossProfit
	bc.NetProfit      = dbc.NetProfit
}

//=============================================================================
