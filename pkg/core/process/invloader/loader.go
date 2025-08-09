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
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/platform"
	"gorm.io/gorm"
	"log/slog"
	"time"
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
	if err != nil {
		slog.Error("processDataProduct: Cannot get instruments from root", "user", dp.Username, "connection", dp.ConnectionCode, "symbol", dp.Symbol, "error", err.Error())
		return
	}

	err = db.RunInTransaction(func(tx *gorm.DB) error {
		for _, instr := range instruments {
			di := db.DataInstrument{
				DataProductId : dp.Id,
				Symbol        : instr.Name,
				Name          : instr.Description,
				ExpirationDate: instr.ExpirationDate,
				Continuous    : instr.Continuous,
				Status        : 0,
			}

			err = db.AddDataInstrument(tx, &di)
			if err != nil {
				return err
			}
		}

		return db.SetDataProductStatus(tx, dp.Id, db.DPStatusFetchingData)
	})

	if err != nil {
		slog.Error("processDataProduct: Cannot add new data instruments", "user", dp.Username, "connection", dp.ConnectionCode, "symbol", dp.Symbol, "error", err.Error())
		return
	}

	slog.Info("processDataProduct: Ending sync process")
}

//=============================================================================
