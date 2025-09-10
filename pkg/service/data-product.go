//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

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
	"encoding/json"
	"io"
	"mime/multipart"
	"time"

	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
)

//=============================================================================

func getDataInstrumentsByProductId(c *auth.Context) {
	pId, err := c.GetIdFromUrl()

	if err == nil {
		var stored bool
		stored, err = c.GetParamAsBool("stored", false)
		if err == nil {
			err = db.RunInTransaction(func(tx *gorm.DB) error {
				list, err := business.GetDataInstrumentsByProductId(tx, c, pId, stored)

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
							err = db.RunInTransaction(func(tx *gorm.DB) error {
								return business.AddDataInstrumentAndJob(tx, c, productId, spec, filename, bytes)
							})

							if err == nil {
								dur := int(time.Now().Sub(start).Seconds())
								_ = c.ReturnObject(&business.DatafileUploadResponse{
									Duration: dur,
									Bytes   : bytes,
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
