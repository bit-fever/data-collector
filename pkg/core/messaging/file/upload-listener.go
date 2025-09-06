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

package file

import (
	"log/slog"
	"time"

	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

func Upload(job *db.IngestionJob) bool {
	//--- Wait 2 secs to allow the commit to complete
	time.Sleep(time.Second *2)

	slog.Info("HandleFileUpload: Uploading data file into datastore", "filename", job.Filename)
	var context *ParserContext

	block,err := setDataBlockInLoading(job)
	if err == nil {
		context,err = ingestDatafile(job,block)
		if err == nil {
			err = setDataBlockInProcessing(job, block, context.DataRange)
			if err == nil {
				slog.Info("HandleFileUpload: Calculating aggregates", "filename", job.Filename)
				err = calcAggregates(context)
				if err == nil {
					err = setBlockInReady(block)
					if err == nil {
						slog.Info("HandleFileUpload: Operation complete", "filename", job.Filename)
						_=ds.DeleteDataFile(job.Filename)
						return true
					}
				}
			}
		}
	}

	slog.Error("HandleFileUpload: Raised error while processing message", "filename", job.Filename, "error", err.Error())
	setJobInError(err, job, block)
	_=ds.DeleteDataFile(job.Filename)
	return true
}

//=============================================================================

func setDataBlockInLoading(job *db.IngestionJob) (*db.DataBlock, error) {
	var b *db.DataBlock
	var err error

	err = db.RunInTransaction(func(tx *gorm.DB) error {
		b,err = db.GetDataBlockById(tx, job.DataBlockId)
		if err != nil {
			return err
		}
		b.Status   = db.DBStatusLoading
		b.Progress = 0

		return db.UpdateDataBlock(tx, b)
	})

	return b,err
}

//=============================================================================

func ingestDatafile(job *db.IngestionJob, b *db.DataBlock) (*ParserContext, error) {
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

	context := NewParserContext(file, config, loc, job, b)
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
	var config *business.DataConfig

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		cfg, err := business.CreateDataConfig(tx, id)
		config = cfg
		return err
	})

	return &config.DataConfig, err
}

//=============================================================================

func setDataBlockInProcessing(job *db.IngestionJob, b *db.DataBlock, dr *DataRange) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		if b.DataFrom.IsNil() || b.DataFrom > dr.FromDay {
			b.DataFrom = dr.FromDay
		}

		if b.DataTo.IsNil() || b.DataTo < dr.ToDay {
			b.DataTo = dr.ToDay
		}

		b.Status  = db.DBStatusProcessing
		err := db.UpdateDataBlock(tx, b)
		if err != nil {
			return err
		}

		return db.UpdateIngestionJob(tx, job)
	})
}

//=============================================================================

func setBlockInReady(block *db.DataBlock) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		block.Status  = db.DBStatusReady
		block.Progress= 100

		return db.UpdateDataBlock(tx, block)
	})
}

//=============================================================================

func setJobInError(err error, job *db.IngestionJob, block *db.DataBlock) {
	_ = db.RunInTransaction(func(tx *gorm.DB) error {
		block.Status = db.DBStatusError
		job  .Error  = err.Error()
		_ = db.UpdateDataBlock(tx, block)
		_ = db.UpdateIngestionJob(tx, job)

		return nil
	})
}

//=============================================================================

func calcAggregates(context *ParserContext) error {
	da5m   := context.DataAggreg
	config := context.Config

	return ds.BuildAggregates(da5m, config)
}

//=============================================================================
