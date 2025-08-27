//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

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

package jobmanager

import (
	"log/slog"
	"time"

	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"github.com/bit-fever/data-collector/pkg/platform"
)

//=============================================================================

type InstrumentDownLoadJob struct {

}

//=============================================================================

func (i *InstrumentDownLoadJob) execute(jc *JobContext) error {
	uc  := jc.userConnection
	sj  := uc.scheduledJob
	blk := sj.block
	job := sj.job

	slog.Info("DownloadJob: Starting job", "systemCode", blk.SystemCode, "root", blk.Root, "symbol", blk.Symbol, "jobId", job.Id, "resuming", jc.resuming)

	for job.LoadFrom <= job.LoadTo {
		err := processDay(jc, uc, blk, job)
		if err != nil {
			return err
		}

		job.LoadFrom = job.LoadFrom.AddDays(1)

		if job.LoadFrom.IsToday(time.UTC) {
			jc.GoToSleep()
			return nil
		}
	}

	slog.Info("DownloadJob: Ending job", "systemCode", blk.SystemCode, "root", blk.Root, "symbol", blk.Symbol, "jobId", job.Id)
	return nil
}

//=============================================================================

func processDay(jc *JobContext, uc *UserConnection, blk *db.DataBlock, job *db.DownloadJob) error {
	bars,err := platform.GetPriceBars(uc.username, uc.connectionCode, blk.Symbol, job.LoadFrom)
	if err == nil {
		job.CurrDay++

		if !bars.NoData {
			err = storeBars(blk, bars.Bars)
			if err == nil {
				err = updateStatus(jc, blk, job)
			}
		}
	}

	return err
}

//=============================================================================

func storeBars(blk *db.DataBlock, bars []*platform.PriceBar) error {
	var dataPoints []*ds.DataPoint
	var dataAggreg = ds.NewDataAggregator(ds.TimeSlotFunction5m)

	config := &ds.DataConfig{
		UserTable: false,
		Selector : blk.SystemCode,
		Timeframe: "1m",
		Symbol   : blk.Symbol,
		Timezone : "UTC",
	}

	for _, bar := range bars {
		dp := &ds.DataPoint{
			Time        : bar.TimeStamp,
			Open        : bar.Open,
			High        : bar.High,
			Low         : bar.Low,
			Close       : bar.Close,
			UpVolume    : bar.UpVolume,
			DownVolume  : bar.DownVolume,
			UpTicks     : bar.UpTicks,
			DownTicks   : bar.DownTicks,
			OpenInterest: bar.OpenInterest,
		}

		dataPoints = append(dataPoints, dp)
		dataAggreg.Add(dp)
	}

	err := ds.SetDataPoints(dataPoints, config)
	if err != nil {
		return err
	}

	dataAggreg.Flush()
	return ds.BuildAggregates(dataAggreg, config)
}

//=============================================================================

func updateStatus(jc *JobContext, blk *db.DataBlock, job *db.DownloadJob) error {
	if blk.DataFrom.IsNil() {
		blk.DataFrom = job.LoadFrom
	}

	if blk.DataTo.IsNil() || blk.DataTo < job.LoadFrom {
		blk.DataTo = job.LoadFrom
	}

	blk.Progress = int8(job.CurrDay * 100 / job.TotDays)

	return jc.UpdateJob(db.DBStatusLoading, db.DJStatusRunning, "")
}

//=============================================================================
