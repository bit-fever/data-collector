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
	"time"

	"github.com/algotiqa/core/auth"
	"github.com/algotiqa/core/req"
	"github.com/algotiqa/data-collector/pkg/core"
	"github.com/algotiqa/data-collector/pkg/db"
	"github.com/algotiqa/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

type BiasSummaryResponse struct {
	BiasAnalysis  *db.BiasAnalysis     `json:"biasAnalysis"`
	BrokerProduct *db.BrokerProduct    `json:"brokerProduct"`
	Result        [7]*DataPointDowList `json:"result"`
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
	slot := (dpd.Hour*60 + dpd.Min) / 30
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
		Year:  int16(dpd.Year),
		Month: int8(dpd.Month),
		Day:   int8(dpd.Day),
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

func GetBiasSummaryInfo(tx *gorm.DB, c *auth.Context, id uint, sessionConfig string) (*BiasSummaryResponse, *core.QueryConfig, error) {
	c.Log.Info("GetBiasSummary: Getting bias analysis", "id", id)

	ba, err := db.GetBiasAnalysisById(tx, id)
	if err != nil {
		return nil, nil, err
	}

	if ba == nil {
		return nil, nil, req.NewNotFoundError("Bias analysis not found")
	}

	var config *core.QueryConfig
	config, err = CreateQueryConfig(tx, ba.DataInstrumentId, sessionConfig)
	if err != nil {
		return nil, nil, err
	}

	var bp *db.BrokerProduct
	bp, err = db.GetBrokerProductById(tx, ba.BrokerProductId)
	if err != nil {
		c.Log.Error("GetBiasSummaryInfo: Could not retrieve broker product", "error", err.Error())
		return nil, nil, err
	}

	c.Log.Info("GetBiasSummary: Found bias analysis", "id", id, "name", ba.Name)

	return &BiasSummaryResponse{
		BiasAnalysis:  ba,
		BrokerProduct: bp,
		Result:        [7]*DataPointDowList{},
	}, config, nil
}

//=============================================================================

func GetBiasSummaryData(c *auth.Context, spec *QuerySpec, bsr *BiasSummaryResponse) error {
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

	for i, dpCurr := range dataPoints {
		if i > 0 {
			dpPrev  := dataPoints[i-1]
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

	y, m, d := slotTime.Date()
	hour := slotTime.Hour()
	mins := slotTime.Minute()
	dow  := slotTime.Weekday()

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
