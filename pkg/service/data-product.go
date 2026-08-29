//=============================================================================
//===
//=== Copyright (C) 2023-present Andrea Carboni
//===
//=== This source code is licensed under the Elastic License 2.0 (ELv2) available at:
//=== https://github.com/algotiqa/docs/blob/main/LICENSE.md
//=== By using this file, you agree to the terms and conditions of that license.
//=============================================================================

package service

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"time"

	"github.com/algotiqa/core/auth"
	"github.com/algotiqa/core/dbms"
	"github.com/algotiqa/data-collector/pkg/business"
	"github.com/algotiqa/data-collector/pkg/core"
	"github.com/algotiqa/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

func getDataInstrumentsByProductId(c *auth.Context) {
	pId, err := c.GetIdFromUrl()

	if err == nil {
		var stored bool
		stored, err = c.GetParamAsBool("stored", false)
		if err == nil {
			err = dbms.RunInTransaction(func(tx *gorm.DB) error {
				list, err := business.GetDataInstrumentsByProductId(c, tx, pId, stored)

				if err != nil {
					return err
				}

				return c.ReturnList(list, 0, len(*list), len(*list))
			})
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func uploadDataInstrumentData(c *auth.Context) {
	productId, err := c.GetIdFromUrl()

	start := time.Now()

	if err == nil {
		var reader *multipart.Reader
		reader, err = c.Gin.Request.MultipartReader()

		if err == nil {
			var part *multipart.Part

			if part, err = reader.NextPart(); err != io.EOF {
				var spec *business.DatafileUploadSpec
				spec, err = retrieveUploadSpec(part)

				if err == nil {
					if part, err = reader.NextPart(); err != io.EOF {
						filename := ""
						var bytes int64
						filename, bytes, err = ds.SaveDatafile(part)
						_ = part.Close()

						if err == nil {
							err = dbms.RunInTransaction(func(tx *gorm.DB) error {
								return business.AddDataInstrumentAndJob(tx, c, productId, spec, filename, bytes)
							})

							if err == nil {
								dur := int(time.Now().Sub(start).Seconds())
								_ = c.ReturnObject(&business.DatafileUploadResponse{
									Duration: dur,
									Bytes:    bytes,
								})
								return
							}
						}
					}
				}
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================

func analyzeDataProduct(c *auth.Context) {
	var result *business.DataProductAnalysisResponse
	var config *core.QueryConfig

	id, err := c.GetIdFromUrl()

	if err == nil {
		err = dbms.RunInTransaction(func(tx *gorm.DB) error {
			sessionConfig := c.GetParamAsString("sessionConfig", "")
			cfg, err1 := business.CreateQueryConfigForProduct(c, tx, id, sessionConfig)
			config = cfg
			return err1
		})

		if err == nil {
			spec   := createQuerySpec(c, id, config)
			atrLen := c.GetParamAsString("atrLen", "")
			result, err = business.AnalyzeProduct(c, spec, atrLen)
			if err == nil {
				_ = c.ReturnObject(result)
				return
			}
		}
	}

	c.ReturnError(err)
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func retrieveUploadSpec(part *multipart.Part) (*business.DatafileUploadSpec, error) {
	bytes, err := io.ReadAll(part)

	if err == nil {
		var spec business.DatafileUploadSpec

		err = json.Unmarshal(bytes, &spec)

		if err == nil {
			err = part.Close()

			if err == nil {
				return &spec, nil
			}
		}
	}

	return nil, err
}

//=============================================================================
