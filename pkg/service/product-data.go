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
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/business"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
)

//=============================================================================

func getInstrumentDataBySourceId(c *auth.Context) {
	pdId, err := c.GetIdFromUrl()

	if err == nil {
		err = db.RunInTransaction(func(tx *gorm.DB) error {
			list, err := business.GetInstrumentDataBySourceId(tx, c, pdId)

			if err != nil {
				return err
			}

			return c.ReturnList(list, 0, len(*list), len(*list))
		})
	}

	c.ReturnError(err)
}

//=============================================================================

func uploadInstrumentData(c *auth.Context) {
	sourceId, err := c.GetIdFromUrl()

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
						var instrData *db.InstrumentData
						var prodData  *db.ProductData

						err = db.RunInTransaction(func(tx *gorm.DB) error {
							instrData, prodData, err = business.PrepareForUpload(tx, c, sourceId, spec)
							return err
						})

						if err == nil {
							var response *business.DatafileUploadResponse
							response, err = business.UploadInstrumentData(c, spec, instrData, prodData, part)

							if err == nil {
								_ = part.Close()
								_ = c.ReturnObject(response)
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
