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
	"os"

	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

var cache *InventoryCache

//=============================================================================

func Init(cfg *app.Config) {
	slog.Info("JobManager: Initializing cache...")

	err := initCache()
	if err != nil {
		slog.Error("Fatal: Cannot initialize Job manager. ", "error", err.Error())
		os.Exit(1)
	}
}

//=============================================================================
//===
//=== Public functions
//===
//=============================================================================

func GetDataBlock(systemCode, root, symbol string) *db.DataBlock {
	return cache.getDataBlock(systemCode, root, symbol)
}

//=============================================================================

func AddDataBlock(block *db.DataBlock) {
	cache.addDataBlock(block)
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func initCache() error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		list,err := db.GetGlobalDataBlocks(tx)
		if err != nil {
			return err
		}

		cache = NewInventoryCache()

		for _,d := range *list {
			cache.addDataBlock(&d)
		}

		return nil
	})
}

//=============================================================================
