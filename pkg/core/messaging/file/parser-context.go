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
	"io"
	"time"

	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

type ParserContext struct {
	Reader     io.Reader
	Config     *ds.DataConfig
	Location   *time.Location
	Job        *db.IngestionJob
	Block      *db.DataBlock
	DataRange  *DataRange
	DataAggreg *ds.DataAggregator

	//--- Private stuff

	dataPoints []*ds.DataPoint
	currBytes  int64
}

//=============================================================================
//===
//=== Constructor
//===
//=============================================================================

func NewParserContext(file io.Reader, config *ds.DataConfig, loc *time.Location, job *db.IngestionJob, b *db.DataBlock) *ParserContext {
	c := &ParserContext{
		Reader  : file,
		Config  : config,
		Location: loc,
		Job     : job,
		Block   : b,
	}

	c.dataPoints = []*ds.DataPoint{}
	c.DataRange  = &DataRange{}
	c.DataAggreg = ds.NewDataAggregator(ds.TimeSlotFunction5m)

	return c
}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func (c *ParserContext) SaveDataPoint(dp *ds.DataPoint, bytes int) error {
	c.dataPoints = append(c.dataPoints, dp)
	c.Job.Records++
	c.currBytes += int64(bytes)

	if c.Job.Records % 8192 == 0 {
		if err := ds.SetDataPoints(c.dataPoints, c.Config); err != nil {
			return err
		}
		c.dataPoints = []*ds.DataPoint{}
	}

	updateDataRange(dp.Time, c.DataRange)
	c.DataAggreg.Add(dp)

	return c.updateProgress()
}

//=============================================================================

func (c *ParserContext) Flush() error {
	c.DataAggreg.Flush()
	return ds.SetDataPoints(c.dataPoints, c.Config)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func (c *ParserContext) updateProgress() error {
	curProgress := int8(c.currBytes * 100 / c.Job.Bytes)

	if c.Block.Progress != curProgress {
		c.Block.Progress = curProgress

		return db.RunInTransaction(func(tx *gorm.DB) error {
			err := db.UpdateDataBlock(tx, c.Block)
			if err != nil {
				return err
			}

			return db.UpdateIngestionJob(tx, c.Job)
		})
	}

	return nil
}

//=============================================================================
