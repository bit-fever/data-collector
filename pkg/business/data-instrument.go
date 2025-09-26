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
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/core/jobmanager"
	"github.com/bit-fever/data-collector/pkg/core/process/invloader"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

var DefaultFrom = time.Date(2000,1,1,0,0,0,0, time.UTC)
var DefaultTo   = time.Date(3000,1,1,0,0,0,0, time.UTC)

//=============================================================================

func GetDataInstruments(tx *gorm.DB, c *auth.Context) (*[]db.DataInstrumentFull, error) {
	filter := map[string]any{}
	filter["username"] = c.Session.Username

	return db.GetDataInstrumentsFull(tx, filter)
}

//=============================================================================

func CreateDataConfig(tx *gorm.DB, id uint) (*DataConfig, error) {
	var p *db.DataProduct

	i, err := db.GetDataInstrumentById(tx, id)
	if err == nil {
		p, err = db.GetDataProductById(tx, i.DataProductId)
		if err == nil {
			var instruments *[]db.DataInstrument
			if i.VirtualInstrument {
				instruments,err = db.GetRollingDataInstrumentsByProductIdFast(tx, p.Id, p.Months)
			}

			return createConfig(i, p, instruments), nil
		}
	}

	return nil, err
}

//=============================================================================

func GetDataInstrumentById(tx *gorm.DB, c *auth.Context, id uint, details bool) (*DataInstrumentExt, error) {
	c.Log.Info("GetDataInstrumentById: Getting a data instrument", "id", id)

	di, err := db.GetDataInstrumentById(tx, id)
	if err != nil {
		return nil, err
	}

	if details {
		//--- Add details (if any)
	}

	die := DataInstrumentExt{
		DataInstrument: *di,
	}

	return &die, nil
}

//=============================================================================

func GetDataInstrumentDataById(c *auth.Context, spec *DataInstrumentDataSpec)(*DataInstrumentDataResponse, error) {
	params,err := parseInstrumentDataParams(spec)
	if err != nil {
		return nil, req.NewBadRequestError(err.Error())
	}

	var dataPoints []*ds.DataPoint

	start := time.Now()
	dataPoints, err = getDataPoints(params, spec.Config)
	durQ := time.Now().Sub(start).Seconds()
	lenQ := len(dataPoints)
	if err != nil {
		return nil, err
	}

	noDataForVirtual := dataPoints == nil

	start = time.Now()
	reduced := false
	dataPoints,reduced = reduceDataPoints(dataPoints, params.Reduction)
	durR := time.Now().Sub(start).Seconds()
	lenR := len(dataPoints)

	c.Log.Info("GetDataInstrumentDataById: Query stats", "durationQ", durQ, "recordsQ", lenQ, "durationR", durR, "recordsR", lenR)

	return &DataInstrumentDataResponse{
		Id              : spec.Id,
		Symbol          : spec.Config.DataConfig.Symbol,
		From            : params.From.Format(time.DateTime),
		To              : params.To.Format(time.DateTime),
		Timeframe       : spec.Config.DataConfig.Timeframe,
		Timezone        : params.Location.String(),
		Reduction       : params.Reduction,
		Reduced         : reduced,
		NoDataForVirtual: noDataForVirtual,
		Records         : len(dataPoints),
		DataPoints      : dataPoints,
	}, nil
}

//=============================================================================
//TODO: user should own the instrument in order to reload (or limited to admins)

func ReloadDataInstrumentData(tx *gorm.DB, c *auth.Context, id uint) (*db.DownloadJob, *db.DataBlock, error) {
	//--- Data instrument

	di,err := db.GetDataInstrumentById(tx, id)
	if err != nil {
		return nil,nil,req.NewServerErrorByError(err)
	}
	if di == nil {
		return nil,nil,req.NewNotFoundError("Data instrument was not found. Id=", id)
	}

	//--- Data product

	var dp *db.DataProduct
	dp,err = db.GetDataProductById(tx, di.DataProductId)
	if err != nil {
		return nil,nil,req.NewServerErrorByError(err)
	}
	if dp == nil {
		return nil,nil,req.NewNotFoundError("Data product was not found. Id=", di.DataProductId)
	}

	//--- Data block

	blk := jobmanager.GetDataBlock(dp.SystemCode, dp.Symbol, di.Symbol)
	if blk == nil {
		slog.Error("ReloadDataInstrumentData: Data block was not found", "symbol", di.Symbol, "root", dp.Symbol)
		return nil,nil,req.NewNotFoundError("Data block was not found. Symbol=", di.Symbol)
	}

	//--- Check status

	if blk.Status != db.DBStatusEmpty && blk.Status != db.DBStatusReady {
		return nil,nil,req.NewBadRequestError("Data instrument must be READY or EMPTY. Id=", id)
	}

	//--- Update data instrument

	di.RolloverDate  = nil
	di.RolloverDelta = 0
	di.RolloverStatus= db.DIRollStatusWaiting

	err = db.UpdateDataInstrument(tx, di)
	if err != nil {
		return nil,nil,req.NewServerErrorByError(err)
	}

	//--- Update data block

	blk.Status   = db.DBStatusWaiting
	blk.DataFrom = 0
	blk.DataTo   = 0
	blk.Progress = 0

	err = db.UpdateDataBlock(tx, blk)
	if err != nil {
		return nil,nil,req.NewServerErrorByError(err)
	}

	//--- Add download job

	job := invloader.CreateDownloadJob(di, blk, 100, dp.Timezone)

	return job, blk, db.AddDownloadJob(tx, job)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func createConfig(i *db.DataInstrument, p *db.DataProduct, instruments *[]db.DataInstrument) *DataConfig {
	var selector  any
	var userTable bool

	if p.SupportsMultipleData {
		userTable = true
		selector  = i.Id
	} else {
		userTable = false
		selector  = p.SystemCode
	}

	return &DataConfig{
		DataConfig: ds.DataConfig{
			UserTable: userTable,
			Timeframe: "1m",
			Selector : selector,
			Symbol   : i.Symbol,
		},
		Timezone         : p.Timezone,
		VirtualInstrument: i.VirtualInstrument,
		Instruments      : instruments,
	}
}

//=============================================================================

func parseInstrumentDataParams(spec *DataInstrumentDataSpec) (*DataInstrumentDataParams, error) {
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

	prLoc,err := time.LoadLocation(spec.Config.Timezone)
	if err != nil {
		return nil, errors.New("Bad product timezone: "+ spec.Config.Timezone)
	}

	da, err3 := buildDataAggregator(&spec.Config.DataConfig, prLoc)
	if err3 != nil {
		return nil, errors.New("Bad timeframe: "+ spec.Config.DataConfig.Timeframe +" ("+ err3.Error() +")")
	}

	red, err := parseReduction(spec.Reduction)

	if err != nil {
		return nil, errors.New("Bad reduction: "+ spec.Reduction +" ("+ err.Error() +")")
	}

	return &DataInstrumentDataParams{
		Location  : loc,
		From      : from.UTC(),
		To        : to.UTC(),
		Reduction : red,
		Aggregator: da,
	}, nil
}

//=============================================================================

func getLocation(timezone string, config *DataConfig) (*time.Location, error) {
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

func buildDataAggregator(config *ds.DataConfig, productLoc *time.Location) (*ds.DataAggregator, error) {
	tf := config.Timeframe

	if tf=="1m" || tf=="5m" || tf=="15m" || tf=="60m" || tf=="1440m" {
		return ds.NewDataAggregator(nil, productLoc), nil
	}

	if tf=="10m" {
		config.Timeframe = "5m"
		return ds.NewDataAggregator(ds.TimeSlotFunction10m, productLoc), nil
	}
	if tf=="30m" {
		config.Timeframe = "15m"
		return ds.NewDataAggregator(ds.TimeSlotFunction30m, productLoc), nil
	}

	return nil, errors.New("allowed values are 1m, 5m, 10m, 15m, 30m, 60m")
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

	if red == 0 {
		return red, nil
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
			currDp.High        = math.Max(currDp.High, dp.High)
			currDp.Low         = math.Min(currDp.Low,  dp.Low)
			currDp.Close       = dp.Close
			currDp.UpVolume   += dp.UpVolume
			currDp.DownVolume += dp.DownVolume
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
//===
//=== Query splitting
//===
//=============================================================================

func getDataPoints(params *DataInstrumentDataParams, config *DataConfig) ([]*ds.DataPoint,error) {
	if !config.VirtualInstrument {
		err := ds.GetDataPoints(params.From, params.To, &config.DataConfig, params.Location, params.Aggregator)
		return params.Aggregator.DataPoints(),err
	}

	//--- Querying the virtual instrument. We need to split into several queries

	chunks := calcInstrumentListToQuery(params.From, params.To, config.Instruments)
	if chunks == nil {
		return nil,nil
	}

	cumulateDeltas(chunks)

	from    := params.From
	dconfig := &config.DataConfig
	aggreg  := ds.NewDataAggregator(nil, nil)

	for i,c := range *chunks {
		to := c.RolloverDate
		if i==len(*chunks)-1 {
			to = params.To
		}

		dconfig.Symbol = c.Symbol
		err := ds.GetDataPoints(from, to, dconfig, params.Location, params.Aggregator)
		if err != nil {
			return nil,err
		}
		shiftDataPoints(params.Aggregator, aggreg, c.Delta)
		from = to.Add(time.Second*30)
	}

	return aggreg.DataPoints(),nil
}

//=============================================================================

func calcInstrumentListToQuery(from,to time.Time, list *[]db.DataInstrument) *[]*QueryChunk {
	var res []*QueryChunk

	for _,di:= range *list {
		if di.RolloverStatus == db.DIRollStatusNoMatch || di.RolloverStatus == db.DIRollStatusNoData {
			continue
		}

		if di.RolloverStatus == db.DIRollStatusReady  {
			if from.Compare(*di.RolloverDate) <= 0 {
				res = append(res, buildQueryChunk(&di))
			}

			//--- Is this the last instrument that contains data?

			if to.Compare(*di.RolloverDate) <= 0 {
				return &res
			}
		}

		if di.RolloverStatus == db.DIRollStatusWaiting {
			//--- We arrived to the last instrument that we suppose to be sleeping
			//--- This is an assumption as we don't join with data_block to get the block's status

			res = append(res, buildQueryChunk(&di))
			return &res
		}
	}

	return nil
}

//=============================================================================

func buildQueryChunk(di *db.DataInstrument) *QueryChunk {
	rollDate := *di.ExpirationDate
	if di.RolloverDate != nil {
		rollDate = *di.RolloverDate
	}
	return &QueryChunk{
		Symbol      : di.Symbol,
		RolloverDate: rollDate,
		Delta       : di.RolloverDelta,
	}
}

//=============================================================================

type QueryChunk struct {
	Symbol       string
	RolloverDate time.Time
	Delta        float64
}

//=============================================================================

func cumulateDeltas(chunks *[]*QueryChunk) {
	index := len(*chunks) - 3

	for index >= 0 {
		curr := (*chunks)[index]
		next := (*chunks)[index+1]
		curr.Delta += next.Delta
		index--
	}
}

//=============================================================================

func shiftDataPoints(source,destin *ds.DataAggregator, delta float64) {
	for _,dp := range source.DataPoints() {
		dp.Open  += delta
		dp.High  += delta
		dp.Low   += delta
		dp.Close += delta

		destin.Add(dp)
	}

	source.Clear()
	destin.Flush()
}

//=============================================================================
