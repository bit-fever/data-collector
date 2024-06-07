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

package business

import (
	"bufio"
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
	"mime/multipart"
	"time"
)

//=============================================================================

func GetInstrumentDataByProductId(tx *gorm.DB, c *auth.Context, id uint)(*[]db.InstrumentData, error) {
	return db.GetInstrumentsByDataId(tx, id)
}

//=============================================================================

func PrepareForUpload(tx *gorm.DB, c *auth.Context, productId uint, spec *DatafileUploadSpec) (*db.InstrumentData, error) {
	c.Log.Info("PrepareForUpload: Creating instrument for a product for data", "productId", productId, "symbol", spec.Symbol)

	_, err := getProductDataAndCheckAccess(tx, c, productId, "PrepareForUpload")
	if err != nil {
		return nil, err
	}

	data, err := db.GetInstrumentBySymbol(tx, productId, spec.Symbol)

	if err != nil {
		return nil, err
	}

	if data == nil {
		data = &db.InstrumentData{
			ProductDataId : productId,
			Symbol        : spec.Symbol,
			Name          : spec.Name,
			ExpirationDate: spec.ExpirationDate,
			IsContinuous  : spec.ExpirationDate == 0,
		}

		err = db.AddInstrumentData(tx, data)
		if err != nil {
			return nil, err
		}
	} else {
		data.Name           = spec.Name
		data.ExpirationDate = spec.ExpirationDate
		data.IsContinuous   = spec.ExpirationDate == 0

		err = db.UpdateInstrumentData(tx, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

//=============================================================================

func UploadInstrumentData(c *auth.Context, instrData *db.InstrumentData, part *multipart.Part) (*DatafileUploadResponse, error) {
	scanner := bufio.NewScanner(part)

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		time.Sleep(time.Microsecond * 5)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	c.Log.Info("Done.")

	return &DatafileUploadResponse{}, nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func getProductDataAndCheckAccess(tx *gorm.DB, c *auth.Context, id uint, function string) (*db.ProductData, error) {
	pd, err := db.GetProductDataById(tx, id)

	if err != nil {
		c.Log.Error(function +": Could not retrieve product for data", "error", err.Error())
		return nil, err
	}

	if pd == nil {
		c.Log.Error(function +": Product for data was not found", "id", id)
		return nil, req.NewNotFoundError("Product for data was not found: %v", id)
	}

	if ! c.Session.IsAdmin() {
		if pd.Username != c.Session.Username {
			c.Log.Error(function +": Product for data not owned by user", "id", id)
			return nil, req.NewForbiddenError("Product for data is not owned by user: %v", id)
		}
	}

	return pd, nil
}

//=============================================================================
