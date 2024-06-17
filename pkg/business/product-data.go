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
	"errors"
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/business/upload"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"gorm.io/gorm"
	"mime/multipart"
	"time"
)

//=============================================================================

func GetInstrumentDataBySourceId(tx *gorm.DB, c *auth.Context, sourceId uint)(*[]db.InstrumentData, error) {
	return db.GetInstrumentsBySourceId(tx, sourceId)
}

//=============================================================================

func PrepareForUpload(tx *gorm.DB, c *auth.Context, sourceId uint, spec *DatafileUploadSpec) (*db.InstrumentData, *db.ProductData, error) {
	c.Log.Info("PrepareForUpload: Creating instrument for a product for data", "sourceId", sourceId, "symbol", spec.Symbol)

	pd, err := getProductDataAndCheckAccess(tx, c, sourceId, "PrepareForUpload")
	if err != nil {
		return nil, nil,err
	}

	data, err := db.GetInstrumentBySymbol(tx, pd.Id, spec.Symbol)

	if err != nil {
		return nil, nil, err
	}

	if data == nil {
		data = &db.InstrumentData{
			ProductDataId : pd.Id,
			Symbol        : spec.Symbol,
			Name          : spec.Name,
			IsContinuous  : spec.Continuous,
		}

		err = db.AddInstrumentData(tx, data)
		if err != nil {
			return nil, nil, err
		}
	} else {
		data.Name         = spec.Name
		data.IsContinuous = spec.Continuous

		err = db.UpdateInstrumentData(tx, data)
		if err != nil {
			return nil, nil, err
		}
	}

	return data, pd, nil
}

//=============================================================================

func UploadInstrumentData(c *auth.Context, spec *DatafileUploadSpec, instrData *db.InstrumentData, prodData *db.ProductData, part *multipart.Part) (*DatafileUploadResponse, error) {
	parser, err  := upload.NewParser(spec.Parser)
	if err != nil {
		return nil, err
	}

	loc,err := retrieveLocation(spec, prodData)
	if err != nil {
		return nil, err
	}

	config := createConfig(instrData, prodData)
	_, err = parser.Parse(part, config, loc)
	_ = part.Close()

	if err != nil {
		c.Log.Info("UploadInstrumentData: Parser error --> "+ err.Error())
		return nil, err
	}

	//--- Update stats info

	c.Log.Info("UploadInstrumentData: Upload completed")
	return &DatafileUploadResponse{}, nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func getProductDataAndCheckAccess(tx *gorm.DB, c *auth.Context, sourceId uint, function string) (*db.ProductData, error) {
	pd, err := db.GetProductDataBySourceId(tx, sourceId)

	if err != nil {
		c.Log.Error(function +": Could not retrieve product for data", "error", err.Error())
		return nil, err
	}

	if pd == nil {
		c.Log.Error(function +": Product for data was not found", "sourceId", sourceId)
		return nil, req.NewNotFoundError("Product for data was not found: %v", sourceId)
	}

	if ! c.Session.IsAdmin() {
		if pd.Username != c.Session.Username {
			c.Log.Error(function +": Product for data not owned by user", "sourceId", sourceId)
			return nil, req.NewForbiddenError("Product for data is not owned by user: %v", sourceId)
		}
	}

	return pd, nil
}

//=============================================================================

func retrieveLocation(spec *DatafileUploadSpec, pd *db.ProductData) (*time.Location, error){
	if spec.Timezone == "gmt" {
		return time.UTC, nil
	}

	if spec.Timezone == "exc" {
		return time.LoadLocation(pd.Timezone)
	}

	return nil, errors.New("Unknown timezone type: "+ spec.Timezone)
}

//=============================================================================

func createConfig(instrData *db.InstrumentData, prodData *db.ProductData) *ds.DataConfig {
	config := ds.DataConfig{
		SystemCode    : prodData.SystemCode,
		ConnectionCode: prodData.ConnectionCode,
		Username      : prodData.Username,
		Symbol        : instrData.Symbol,
	}

	if ! prodData.SupportsMultipleData {
		config.ConnectionCode = "*"
		config.Username       = "*"
	}

	return &config
}

//=============================================================================
