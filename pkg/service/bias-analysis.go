//=============================================================================
//===
//=== Copyright (C) 2024-present Andrea Carboni
//===
//=== This source code is licensed under the Elastic License 2.0 (ELv2) available at:
//=== https://github.com/algotiqa/docs/blob/main/LICENSE.md
//=== By using this file, you agree to the terms and conditions of that license.
//=============================================================================

package service

import (
	"github.com/algotiqa/core/auth"
	"github.com/algotiqa/core/dbms"
	"github.com/algotiqa/data-collector/pkg/business"
	"github.com/algotiqa/data-collector/pkg/core"
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
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
		id, err := c.GetIdFromUrl()

		if err == nil {
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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

func deleteBiasAnalysis(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
			ba, err := business.DeleteBiasAnalysis(tx, c, id)

			if err != nil {
				return err
			}

			return c.ReturnObject(&ba)
		})
	}

	c.ReturnError(err)
}

//=============================================================================
//=== BiasConfig
//=============================================================================

func getBiasConfigsByAnalysisId(c *auth.Context) {
	id, err := c.GetIdFromUrl()

	if err == nil {
		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
				err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
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
		var cfg *core.QueryConfig

		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
			sessionConfig := c.GetParamAsString("sessionConfig", "")
			bsr, cfg, err = business.GetBiasSummaryInfo(tx, c, id, sessionConfig)
			return err
		})

		if err == nil {
			spec := createQuerySpec(c, id, cfg)
			err = business.GetBiasSummaryData(c, spec, bsr)
			if err == nil {
				_ = c.ReturnObject(bsr)
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
		var bbr *business.BiasBacktestResponse
		var cfg *core.QueryConfig
		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
			sessionConfig := c.GetParamAsString("sessionConfig", "")
			bbr, cfg, err = business.GetBacktestInfo(tx, c, id, sessionConfig)
			return err
		})

		if err == nil {
			spec := createQuerySpec(c, id, cfg)
			err = business.RunBacktest(c, spec, bbr)
			if err == nil {
				_ = c.ReturnObject(bbr)
				return
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================
