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
	"errors"
	"log/slog"
	"os"
	"strconv"

	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

var cache *InventoryCache = newInventoryCache()

//=============================================================================

func Init(cfg *app.Config) {
	slog.Info("JobManager: Initializing cache...")

	err := initCache()
	if err != nil {
		slog.Error("Fatal: Cannot initialize Job manager. ", "error", err.Error())
		os.Exit(1)
	}

	startScheduler()
}

//=============================================================================
//===
//=== Public functions
//===
//=============================================================================

func NewScheduledJob(block *db.DataBlock, job *db.DownloadJob) *ScheduledJob {
	return &ScheduledJob{block, job}
}

//=============================================================================

func GetDataBlock(systemCode, root, symbol string) *db.DataBlock {
	return cache.getDataBlock(systemCode, root, symbol)
}

//=============================================================================

func AddScheduledJobs(jobs []*ScheduledJob) {
	for _, job := range jobs {
		cache.addScheduledJob(job,nil)
	}
}

//=============================================================================

func SetConnection(systemCode, username, connCode string, connected bool) {
	cache.setConnection(systemCode, username, connCode, connected)
}

//=============================================================================

func DisconnectAll() {
	cache.disconnectAll()
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func initCache() error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		blocksMap,err := loadDataBlocks(tx)
		if err != nil {
			return err
		}

		err = loadDataProducts(tx)
		if err != nil {
			return err
		}

		err = loadDownloadJobs(tx, blocksMap)
		if err != nil {
			return err
		}

		return nil
	})
}

//=============================================================================

func loadDataBlocks(tx *gorm.DB) (map[uint]*db.DataBlock,error) {
	list,err := db.GetGlobalDataBlocks(tx)
	if err != nil {
		return nil,err
	}

	for _,d := range *list {
		cache.addDataBlock(&d)
	}

	blockMap := convertToMap(list)

	return blockMap,nil
}

//=============================================================================

func convertToMap(list *[]db.DataBlock) map[uint]*db.DataBlock {
	res := make(map[uint]*db.DataBlock)

	for _, b := range *list {
		res[b.Id] = &b
	}

	return res
}

//=============================================================================

func loadDataProducts(tx *gorm.DB) error {
	filter :=map[string]any{
		"supports_multiple_data":false,
	}
	products,err := db.GetDataProducts(tx, filter,0,5000)
	if err == nil {
		for _,dp := range *products {
			cache.setConnection(dp.SystemCode, dp.Username, dp.ConnectionCode, dp.Connected)
		}
	}

	return err
}

//=============================================================================

func loadDownloadJobs(tx *gorm.DB, blocksMap map[uint]*db.DataBlock) error {
	jobs,err := db.GetActiveDownloadJobs(tx)
	if err == nil {
		for _, job := range *jobs {
			block,found := blocksMap[job.DataBlockId]
			if !found {
				return errors.New("DataBlock not found! --> id:"+ strconv.Itoa(int(job.DataBlockId)))
			}

			sj := NewScheduledJob(block, &job)
			cache.addScheduledJob(sj,resumer)
		}
	}

	return err
}

//=============================================================================
