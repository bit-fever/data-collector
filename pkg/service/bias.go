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
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================
//=== BiasAnalysis
//=============================================================================

func getBiasAnalyses(c *auth.Context) {
	filter := map[string]any{}
	offset, limit, err := c.GetPagingParams()

	if err == nil {
		details, err := c.GetParamAsBool("details", false)

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				list, err := business.GetBiasAnalyses(tx, c, filter, offset, limit, details)

				if err != nil {
					return err
				}

				return c.ReturnList(list, offset, limit, len(*list))
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func getBiasAnalysisById(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		details, err := c.GetParamAsBool("details", false)

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				pb, err := business.GetBiasAnalysisById(tx, c, id, details)

				if err != nil {
					return err
				}

				return c.ReturnObject(&pb)
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func addBiasAnalysis(c *auth.Context) {
	var pds business.BiasAnalysisSpec
	err := c.BindParamsFromBody(&pds)

	if err == nil {
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			ts, err := business.AddBiasAnalysis(tx, c, &pds)

			if err != nil {
				return err
			}

			return c.ReturnObject(ts)
		})
	}

	c.ReturnError(err)
}

//=============================================================================

func updateBiasAnalysis(c *auth.Context) {
	var pds business.BiasAnalysisSpec
	err := c.BindParamsFromBody(&pds)

	if err == nil {
		id,err := c.GetIdFromUrl()

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				ts, err := business.UpdateBiasAnalysis(tx, c, id, &pds)

				if err != nil {
					return err
				}

				return c.ReturnObject(ts)
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================
//=== BiasConfig
//=============================================================================

func getBiasConfigsByAnalysisId(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			list, err := business.GetBiasConfigsByAnalysisId(tx, c, id)

			if err != nil {
				return err
			}

			return c.ReturnList(list, 0, 5000, len(*list))
		})
	}

	c.ReturnError(err)
}

//=============================================================================

func addBiasConfig(c *auth.Context) {
	baId, err := c.GetIdFromUrl()

	if err == nil {
		var bcs business.BiasConfigSpec
		err = c.BindParamsFromBody(&bcs)

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				bc, err := business.AddBiasConfig(tx, c, baId, &bcs)

				if err != nil {
					return err
				}

				return c.ReturnObject(bc)
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func updateBiasConfig(c *auth.Context) {
	baId, err := c.GetIdFromUrl()

	if err == nil {
		var bcId uint
		bcId, err = c.GetId2FromUrl()

		if err == nil {
			var bcs business.BiasConfigSpec
			err = c.BindParamsFromBody(&bcs)

			if err == nil {
				err = db.RunInTransaction(func(tx *gorm.DB) error {
					bc, err := business.UpdateBiasConfig(tx, c, baId, bcId, &bcs)

					if err != nil {
						return err
					}

					return c.ReturnObject(bc)
				})
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func deleteBiasConfig(c *auth.Context) {
	baId, err := c.GetIdFromUrl()

	if err == nil {
		var bcId uint
		bcId, err = c.GetId2FromUrl()

		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				bc, err := business.DeleteBiasConfig(tx, c, baId, bcId)

				if err != nil {
					return err
				}

				return c.ReturnObject(bc)
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================
//=== Summary
//=============================================================================

func getBiasSummary(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		var bsr *business.BiasSummaryResponse

		err = db.RunInTransaction(func(tx *gorm.DB) error {
			bsr, err = business.GetBiasSummaryInfo(tx, c, id)
			return err
		})

		if err == nil {
			err = business.GetBiasSummaryData(c, id, bsr)
			if err == nil {
				_=c.ReturnObject(bsr)
				return
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================
//=== Backtesting
//=============================================================================

func runBacktest(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		var bts *business.BiasBacktestSpec
		err = c.BindParamsFromBody(&bts)

		if err == nil {
			var bbr *business.BiasBacktestResponse

			err = db.RunInTransaction(func(tx *gorm.DB) error {
				bbr, err = business.GetBacktestInfo(tx, c, id)
				return err
			})

			if err == nil {
				err = business.RunBacktest(c, bbr)
				if err == nil {
					_=c.ReturnObject(bbr)
					return
				}
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================
