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

package invloader

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/bit-fever/core/datatype"
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/bit-fever/data-collector/pkg/core/jobmanager"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/platform"
	"gorm.io/gorm"
)

//=============================================================================

var ticker *time.Ticker

//=============================================================================

func Init(cfg *app.Config) *time.Ticker {
	ticker = time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			run()
		}
	}()

	return ticker
}

//=============================================================================
//===
//=== Inventory loader
//===
//=============================================================================

func run() {
	products,err := getDataProductsToWork()
	if err != nil {
		slog.Error("Cannot retrieve data products to work", "error", err)
		return
	}

	if len(*products) == 0 {
		return
	}

	for _, dp := range *products {
		processDataProduct(&dp)
	}
}

//=============================================================================

func getDataProductsToWork() (*[]db.DataProduct, error) {
	filter := map[string]any{}
	filter["supports_multiple_data"] = false
	filter["connected"]              = true
	filter["status"]                 = db.DPStatusFetchingInventory

	var list *[]db.DataProduct

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		var err error
		list,err = db.GetDataProducts(tx, filter, 0, 1000)
		return err
	})

	return list, err
}

//=============================================================================

func processDataProduct(dp *db.DataProduct) {
	slog.Info("processDataProduct: Start loading inventory for product", "user", dp.Username, "connection", dp.ConnectionCode, "symbol", dp.Symbol)

	instruments,err := platform.GetInstruments(dp.Username, dp.ConnectionCode, dp.Symbol)

	var blocks []*db.DataBlock

	if err != nil {
		err = errors.New("Cannot get instruments from root : "+err.Error())
	} else {
		list := convertInstruments(dp.Id, instruments)

		err = db.RunInTransaction(func(tx *gorm.DB) error {
			blocks,err = addDataInstruments(tx, dp, list)
			if err != nil {
				return errors.New("Cannot add new data instruments : "+ err.Error())
			}

			return db.UpdateDataProductFields(tx, dp.Id, db.DPStatusFetchingData)
		})
	}

	//--- Ending process

	if err != nil {
		slog.Error("processDataProduct: Operation aborted", "user", dp.Username, "connection", dp.ConnectionCode, "symbol", dp.Symbol, "error", err.Error())
	} else {
		for _, ai := range blocks {
			jobmanager.AddDataBlock(ai)
		}

		slog.Info("processDataProduct: End loading inventory for product", "user", dp.Username, "connection", dp.ConnectionCode, "symbol", dp.Symbol)
	}
}

//=============================================================================

func convertInstruments(dpId uint, instruments []platform.Instrument) []*db.DataInstrument {
	var list []*db.DataInstrument

	for _, instr := range instruments {
		di := db.DataInstrument{
			DataProductId : dpId,
			Symbol        : instr.Name,
			Name          : instr.Description,
			ExpirationDate: instr.ExpirationDate,
			Continuous    : instr.Continuous,
			Month         : instr.Month,
		}

		list = append(list, &di)
	}

	return list
}

//=============================================================================

func addDataInstruments(tx *gorm.DB, dp *db.DataProduct, instruments []*db.DataInstrument) ([]*db.DataBlock, error) {
	var blocks []*db.DataBlock

	for _, di := range instruments {
		var isNew bool
		var block *db.DataBlock
		var err error

		if shouldLoad(di, dp.Months) {
			block,isNew,err = getOrCreateDataBlock(tx, dp, di)
			if err != nil {
				return nil,err
			}
		}

		err = db.AddDataInstrument(tx, di)
		if err != nil {
			return nil,err
		}

		if isNew {
			blocks = append(blocks, block)
			err = addDownloadJob(tx, block, di)
			if err != nil {
				return nil,err
			}
		}
	}

	return blocks, nil
}

//=============================================================================

func shouldLoad(di *db.DataInstrument, months string) bool {
	if di.Continuous {
		return false
	}

	return strings.Index(months, di.Month) != -1
}

//=============================================================================

func getOrCreateDataBlock(tx *gorm.DB, dp *db.DataProduct, di *db.DataInstrument) (*db.DataBlock, bool, error) {
	block := jobmanager.GetDataBlock(dp.SystemCode, dp.Symbol, di.Symbol)
	isNew := false

	if block == nil {
		block = &db.DataBlock{
			SystemCode: dp.SystemCode,
			Root      : dp.Symbol,
			Symbol    : di.Symbol,
			Global    : true,
			Status    : db.DBStatusWaiting,
		}

		err := db.AddDataBlock(tx, block)
		if err != nil {
			return nil,false,err
		}

		isNew = true
	}

	di.DataBlockId = &block.Id

	return block,isNew,nil
}

//=============================================================================

func addDownloadJob(tx *gorm.DB, block *db.DataBlock, di *db.DataInstrument) error {
	job := &db.DownloadJob{
		DataInstrumentId: di.Id,
		DataBlockId     : block.Id,
		LoadFrom        : calcLoadFrom(di.ExpirationDate),
		LoadTo          : calcLoadTo(di.ExpirationDate),
	}

	return db.AddDownloadJob(tx, job)
}

//=============================================================================

func calcLoadFrom(expDate *time.Time) datatype.IntDate {
	//--- For every instrument, we consider 1 year of data
	old := expDate.Add(-365 * 24 *time.Hour)
	return datatype.ToIntDate(&old)
}

//=============================================================================

func calcLoadTo(expDate *time.Time) datatype.IntDate {
	return datatype.ToIntDate(expDate)
}

//=============================================================================
