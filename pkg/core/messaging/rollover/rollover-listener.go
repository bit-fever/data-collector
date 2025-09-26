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

package rollover

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

func Recalc(job *RecalcJob) bool {
	if job.DataProductId != 0 {
		return recalcForProduct(job.DataProductId)
	} else {
		list,err := getProductsToRecalc(job.DataBlockId)
		if err == nil {
			for _, id := range *list {
				ok := recalcForProduct(id)
				if !ok {
					return false
				}
			}
		}
	}

	return true
}

//=============================================================================

func getProductsToRecalc(blockId uint) (*[]uint,error) {
	var list *[]uint

	err2 := db.RunInTransaction(func(tx *gorm.DB) error {
		var err error
		list,err = db.GetDataProductsByBlockId(tx, blockId)
		return err
	})

	if err2 != nil {
		return nil,err2
	}

	return list,nil
}

//=============================================================================

func recalcForProduct(id uint) bool {
	slog.Info("recalcForProduct: Starting rollover recalc", "dpId", id)

	dp,instruments,err := getIntrumentSet(id)
	if err == nil {
		var updated []*db.DataInstrumentExt
		var curr,next *db.DataInstrumentExt

		for i:=0; i<len(*instruments)-1; i++ {
			curr = &(*instruments)[i]
			next = &(*instruments)[i+1]

			var toUpdate bool

			//--- Check if we have to calculate the rollover

			shouldRecalc := curr.RolloverDate   == nil ||
							curr.RolloverStatus == db.DIRollStatusNoData ||
							curr.RolloverStatus == db.DIRollStatusNoMatch

			if shouldRecalc {
				if *curr.Status == db.DBStatusReady {
					//--- First block loaded. Check the second one

					if *next.Status == db.DBStatusReady || *next.Status == db.DBStatusSleeping {
						toUpdate, err = calcRollover(dp, curr, next, dp.RolloverTrigger)
						if err != nil {
							break;
						}
					} else if *next.Status == db.DBStatusEmpty {
						toUpdate = setFakeRolloverDate(curr, dp)
					}
				} else if *curr.Status == db.DBStatusEmpty {
					toUpdate = setFakeRolloverDate(curr, dp)
				}

				if toUpdate {
					updated = append(updated, curr)
				}
			}
		}

		err = updateRolledInstruments(updated)

		if err == nil {
			err = recalcVirtualInstrumentStatus(dp, instruments)
		}
	}

	if err != nil {
		slog.Error("recalcForProduct: Operation aborted due to error. Will retry", "dpId", id, "error", err)
	}

	slog.Info("recalcForProduct: Ending rollover recalc", "dpId", id)

	return err == nil
}

//=============================================================================

func getIntrumentSet(dpId uint) (*db.DataProduct,*[]db.DataInstrumentExt,error) {
	var dp   *db.DataProduct
	var list *[]db.DataInstrumentExt

	err2 := db.RunInTransaction(func(tx *gorm.DB) error {
		var err error
		dp,err = db.GetDataProductById(tx, dpId)
		if err == nil {
			if dp == nil {
				err = errors.New("No data product found : "+ strconv.Itoa(int(dpId)))
			} else {
				list,err = db.GetRollingDataInstrumentsByProductId(tx, dpId, dp.Months)
			}
		}
		return err
	})

	if err2 != nil {
		return nil,nil,err2
	}

	return dp,list,nil
}

//=============================================================================

func calcRollover(dp *db.DataProduct, curr, next *db.DataInstrumentExt, rollTrigger db.DPRollTrigger) (bool,error) {
	startRollDate := calcRolloverDate(*curr.ExpirationDate, rollTrigger)

	if *next.Status == db.DBStatusSleeping && time.Now().Sub(startRollDate) <8*time.Hour {
		//--- If the startRollDate is within 8 hours behind now, let's skip
		return false, nil
	}

	return true, calcRolloverDelta(dp, curr, next, startRollDate)
}

//=============================================================================

func calcRolloverDelta(dp *db.DataProduct, curr, next *db.DataInstrumentExt, startRollDate time.Time) error {
	prices1,err1 := getPrices(dp.SystemCode, curr.Symbol, startRollDate)
	prices2,err2 := getPrices(dp.SystemCode, next.Symbol, startRollDate)
	if err1 != nil {
		return errors.New("Failed to get prices from current: "+err1.Error())
	}
	if err2 != nil {
		return errors.New("Failed to get prices from next: "+err2.Error())
	}

	currIdx := 0
	nextIdx := 0

	for currIdx<len(prices1) && nextIdx<len(prices2) {
		p1 := prices1[currIdx]
		p2 := prices2[nextIdx]

		res := p1.Time.Compare(p2.Time)

		if res == -1 {
			currIdx++
		} else if res == 1 {
			nextIdx++
		} else {
			//--- Ok, found the same time. Now calc delta
			curr.RolloverDate   = &p1.Time
			curr.RolloverDelta  = p2.Close - p1.Close
			curr.RolloverStatus = db.DIRollStatusReady
			return nil
		}
	}

	slog.Error("calcRolloverDelta: Cannot find any rollover delta", "dpId", dp.Id, "currId", curr.Id, "nextId", next.Id, "startRollDate", startRollDate)

	curr.RolloverStatus = db.DIRollStatusNoMatch
	curr.RolloverDelta  = 0
	curr.RolloverDate   = &startRollDate

	return nil
}

//=============================================================================

func getPrices(systemCode, symbol string, from time.Time) ([]*ds.DataPoint, error){
	config := ds.NewDataConfig(systemCode, symbol,"60m")
	da     := ds.NewDataAggregator(nil,nil)
	to     := from.Add(5 * 24 * time.Hour)

	err    := ds.GetDataPoints(from, to, config, time.UTC, da)
	if err != nil {
		return nil, err
	}

	return da.DataPoints(),nil
}

//=============================================================================

func updateRolledInstruments(list []*db.DataInstrumentExt) error {
	if len(list) == 0 {
		return nil
	}

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		for _, die := range list {
			i := convertInstrument(die)
			err := db.UpdateDataInstrument(tx, i)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("updateRolledInstruments: Failed to update rolled instruments", "error", err)
	}

	return err
}

//=============================================================================

func convertInstrument(die *db.DataInstrumentExt) *db.DataInstrument {
	return &db.DataInstrument{
		Id:             die.Id,
		DataProductId:  die.DataProductId,
		DataBlockId:    die.DataBlockId,
		Symbol:         die.Symbol,
		Name:           die.Name,
		ExpirationDate: die.ExpirationDate,
		RolloverDate:   die.RolloverDate,
		Continuous:     die.Continuous,
		Month:          die.Month,
		RolloverDelta:  die.RolloverDelta,
		RolloverStatus: die.RolloverStatus,
	}
}

//=============================================================================

func setFakeRolloverDate(die *db.DataInstrumentExt, dp *db.DataProduct) bool {
	rollDate          := calcRolloverDate(*die.ExpirationDate, dp.RolloverTrigger)
	die.RolloverDate   = &rollDate
	die.RolloverDelta  = 0
	die.RolloverStatus = db.DIRollStatusNoData

	return true
}

//=============================================================================

func recalcVirtualInstrumentStatus(dp *db.DataProduct, instruments *[]db.DataInstrumentExt) error {
	var emptyDie, noMatchDie *db.DataInstrumentExt

	for _, die := range *instruments {
		status := *die.Status
		if status != db.DBStatusReady && status != db.DBStatusEmpty && status != db.DBStatusSleeping {
			return nil
		}

		if status == db.DBStatusEmpty {
			emptyDie = &die
		}

		if die.RolloverStatus == db.DIRollStatusNoMatch {
			noMatchDie = &die
		}

		if status == db.DBStatusSleeping {
			break
		}
	}

	var vi *db.DataInstrument

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		var err error
		vi,err=db.GetVirtualDataInstrumentByProductId(tx, dp.Id)
		if err == nil {
			if vi == nil {
				return errors.New("Cannot find Virtual Data instrument for product with id: "+ strconv.Itoa(int(dp.Id)))
			}
			err = sendEventToUser(dp, instruments, emptyDie, noMatchDie, vi)
			if err == nil {
				return db.UpdateDataInstrument(tx, vi)
			}
		}
		return err
	})

	if err != nil {
		slog.Error("recalcProductStatus: Failed to set data product status", "error", err)
	} else {
		slog.Info("recalcProductStatus: Data product is ready", "dpId", dp.Id, "root", dp.Symbol)
	}

	return err
}

//=============================================================================

func sendEventToUser(dp *db.DataProduct, instruments *[]db.DataInstrumentExt, empty, noMatch *db.DataInstrumentExt, vi *db.DataInstrument) error {
	if empty == nil && noMatch == nil {
		vi.RolloverStatus = db.DIRollStatusReady
		return msg.SendEventByCode(dp.Username, "dc.dataProduct.ready", map[string]interface{}{
			"root"       : dp.Symbol,
			"virtual"    : vi.Symbol,
			"instruments": len(*instruments),
		})
	}

	if empty != nil {
		vi.RolloverStatus = db.DIRollStatusNoData
		return msg.SendEventByCode(dp.Username, "dc.dataProduct.readyEmpty", map[string]interface{}{
			"root"   : dp.Symbol,
			"symbol" : empty.Symbol,
			"virtual": vi.Symbol,
		})
	}

	vi.RolloverStatus = db.DIRollStatusNoMatch
	return msg.SendEventByCode(dp.Username, "dc.dataProduct.readyNoMatch", map[string]interface{}{
		"root"  : dp.Symbol,
		"symbol": noMatch.Symbol,
		"system": dp.SystemCode,
		"virtual": vi.Symbol,
	})
}

//=============================================================================
