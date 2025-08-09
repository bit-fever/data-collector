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

package upload

import (
	"encoding/json"
	"github.com/bit-fever/core/datatype"
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

//=============================================================================

func HandleUploadMessage(m *msg.Message) bool {

	slog.Info("New upload message received", "source", m.Source, "type", m.Type)

	if m.Source == msg.SourceUploadJob {
		job := db.UploadJob{}
		err := json.Unmarshal(m.Entity, &job)
		if err != nil {
			slog.Error("Dropping badly formatted message!", "entity", string(m.Entity))
			return true
		}

		if m.Type == msg.TypeCreate {
			return uploadFile(&job)
		}
	}

	slog.Error("Dropping message with unknown source/type!", "source", m.Source, "type", m.Type)
	return true
}

//=============================================================================

func uploadFile(job *db.UploadJob) bool {
	//--- Wait 2 secs to allow the commit to complete
	time.Sleep(time.Second *2)

	slog.Info("uploadFile: Uploading data file into datastore", "filename", job.Filename)
	var context *ParserContext

	err := setJobInAdding(job)
	if err == nil {
		context,err = ingestDatafile(job)
		if err == nil {
			err = updateJob(job, context.DataRange)
			if err == nil {
				slog.Info("uploadFile: Calculating aggregates", "filename", job.Filename)
				err = calcAggregates(context)
				if err == nil {
					err = setJobInReady(job)
					if err == nil {
						slog.Info("uploadFile: Operation complete", "filename", job.Filename)
						_=ds.DeleteDataFile(job.Filename)
						return true
					}
				}
			}
		}
	}

	slog.Error("uploadFile: Raised error while processing message", "filename", job.Filename, "error", err.Error())
	setJobInError(err, job)
	_=ds.DeleteDataFile(job.Filename)
	return true
}

//=============================================================================

func ingestDatafile(job *db.UploadJob) (*ParserContext, error) {
	start := time.Now()

	parser,err := NewParser(job.Parser)
	if err != nil {
		return nil,err
	}

	loc,err := retrieveLocation(job.Timezone)
	if err != nil {
		return nil,err
	}

	config,err := retrieveConfig(job.DataInstrumentId)
	if err != nil {
		return nil,err
	}

	file,err := ds.OpenDatafile(job.Filename)
	if err != nil {
		return nil,err
	}

	context := NewParserContext(file, config, loc, job)
	defer file.Close()

	err = parser.Parse(context)
	if err != nil {
		slog.Error("ingestDatafile: Parser error --> "+ err.Error())
		return nil, err
	}

	//--- Return stats

	end := time.Now()
	dur := end.Sub(start)

	slog.Info("ingestDatafile: Upload complete", "records", job.Records, "duration", dur.Seconds())

	return context, nil
}

//=============================================================================

func retrieveLocation(timezone string) (*time.Location, error){
	if timezone == "utc" {
		return time.UTC, nil
	}

	return time.LoadLocation(timezone)
}

//=============================================================================

func retrieveConfig(id uint) (*ds.DataConfig, error) {
	var config *ds.DataConfig

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		cfg, err := business.CreateDataConfig(tx, id)
		config = cfg
		return err
	})

	return config, err
}

//=============================================================================

func updateJob(job *db.UploadJob, dr *DataRange) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		i, err := db.GetDataInstrumentById(tx, job.DataInstrumentId)
		if err == nil {
			err = updateLoadedPeriod(tx, i, dr)
			if err == nil {
				err = updateUploadJob(tx, job)
			}
		}

		return err
	})
}

//=============================================================================

func updateLoadedPeriod(tx *gorm.DB, i *db.DataInstrument, dr *DataRange) error {
	//--- Update loaded period

	if i.DataFrom == 0 || i.DataFrom > dr.FromDay {
		i.DataFrom = dr.FromDay
	}

	if i.DataTo == 0 || i.DataTo < dr.ToDay {
		i.DataTo = dr.ToDay
	}

	if !i.Continuous {
		i.ExpirationDate = nil

		if i.DataTo != 0 {
			d := datatype.IntDate(i.DataTo)
			t := d.ToDateTime(false, time.UTC)
			i.ExpirationDate = &t
		}
	}
	return db.UpdateDataInstrument(tx, i)
}

//=============================================================================

func updateUploadJob(tx *gorm.DB, job *db.UploadJob) error {
	job.Status  = db.UploadJobStatusAggregating
	job.Progress= 0
	return db.UpdateUploadJob(tx, job)
}

//=============================================================================

func setJobInAdding(job *db.UploadJob) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		job.Status = db.UploadJobStatusAdding
		return db.UpdateUploadJob(tx, job)
	})
}

//=============================================================================

func setJobInReady(job *db.UploadJob) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		var i *db.DataInstrument

		job.Status  = db.UploadJobStatusReady
		job.Progress= 100

		err := db.UpdateUploadJob(tx, job)
		if err == nil {
			i, err = db.GetDataInstrumentById(tx, job.DataInstrumentId)
			if err == nil {
				i.Status = db.InstrumentStatusReady
				err = db.UpdateDataInstrument(tx, i)
			}
		}

		return err
	})
}

//=============================================================================

func setJobInError(err error, job *db.UploadJob) {
	_ = db.RunInTransaction(func(tx *gorm.DB) error {
		job.Status = db.UploadJobStatusError
		job.Error = err.Error()
		_ = db.UpdateUploadJob(tx, job)

		i, err := db.GetDataInstrumentById(tx, job.DataInstrumentId)
		if err == nil {
			i.Status = db.InstrumentStatusError
			_ = db.UpdateDataInstrument(tx, i)
		}

		return nil
	})
}

//=============================================================================

func calcAggregates(context *ParserContext) error {
	da5m   := context.DataAggreg
	config := context.Config
	err := saveAggregate(da5m, config, "5m")

	if err == nil {
		da15m := ds.NewDataAggregator(ds.TimeSlotFunction15m)
		da5m.Aggregate(da15m)
		err = saveAggregate(da15m, config, "15m")
		if err == nil {
			da60m := ds.NewDataAggregator(ds.TimeSlotFunction60m)
			da15m.Aggregate(da60m)
			err = saveAggregate(da60m, config, "60m")
		}
	}

	return err
}

//=============================================================================

func saveAggregate(da *ds.DataAggregator, config *ds.DataConfig, timeframe string) error {
	var dataPoints []*ds.DataPoint
	config.Timeframe = timeframe

	for _,dp := range da.DataPoints() {
		dataPoints = append(dataPoints, dp)

		if len(dataPoints) == 8192 {
			if err := ds.SetDataPoints(dataPoints, config); err != nil {
				return err
			}
			dataPoints = []*ds.DataPoint{}
		}
	}

	return ds.SetDataPoints(dataPoints, config)
}

//=============================================================================
