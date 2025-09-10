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

package service

import (
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/core/jobmanager"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

func getDataInstruments(c *auth.Context) {
	err:= db.RunInTransaction(func(tx *gorm.DB) error {
		list, err := business.GetDataInstruments(tx, c)

		if err != nil {
				return err
			}

		return c.ReturnList(list, 0, len(*list), len(*list))
	})

	c.ReturnError(err)
}

//=============================================================================

func getDataInstrumentById(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		var details bool
		details, err = c.GetParamAsBool("details", false)

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				var di *business.DataInstrumentExt
				di, err = business.GetDataInstrumentById(tx, c, id, details)

				if err != nil {
					return err
				}

				return c.ReturnObject(di)
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func getDataInstrumentData(c *auth.Context) {
	var result *business.DataInstrumentDataResponse
	var config *business.DataConfig

	id, err   := c.GetIdFromUrl()
	timeframe := c.GetParamAsString("timeframe",  "5m")

	if err == nil {
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			cfg, err := business.CreateDataConfig(tx, id)
			config = cfg
			return err
		})

		if err == nil {
			config.DataConfig.Timeframe = timeframe
			spec := &business.DataInstrumentDataSpec{
				Id       : id,
				From     : c.GetParamAsString("from",     ""),
				To       : c.GetParamAsString("to",       ""),
				Timezone : c.GetParamAsString("timezone", "UTC"),
				Reduction: c.GetParamAsString("reduction",""),
				Config   : config,
			}
			result, err = business.GetDataInstrumentDataById(c, spec)
			if err == nil {
				_=c.ReturnObject(result)
				return
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func reloadDataInstrumentData(c *auth.Context) {
	id,err := c.GetIdFromUrl()

	if err == nil {
		var job *db.DownloadJob
		var blk *db.DataBlock
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			job,blk,err = business.ReloadDataInstrumentData(tx, c, id)
			return err
		})

		if err == nil {
			sj := jobmanager.NewScheduledJob(blk, job)
			jobmanager.AddScheduledJob(sj)
		}
	}

	c.ReturnError(err)
}

//=============================================================================
