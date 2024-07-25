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

package db

import (
	"github.com/bit-fever/core/req"
	"gorm.io/gorm"
)

//=============================================================================

func GetInstrumentById(tx *gorm.DB, id uint) (*Instrument, error) {
	var list []Instrument
	res := tx.Find(&list, id)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	if len(list) == 1 {
		return &list[0], nil
	}

	return nil, nil
}

//=============================================================================

func GetInstrumentsBySourceId(tx *gorm.DB, sourceId uint) (*[]Instrument, error) {
	var list []Instrument

	filter := map[string]any{}
	filter["source_id"] = sourceId

	res := tx.
			Where(filter).
			Joins("JOIN product pd ON pd.id = product_id").
			Order("expiration_date").
			Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetInstrumentBySymbol(tx *gorm.DB, productId uint, symbol string) (*Instrument, error) {
	filter := map[string]any{}
	filter["product_id"] = productId
	filter["symbol"]     = symbol

	var list []Instrument
	res := tx.Where(filter).Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	if len(list) == 1 {
		return &list[0], nil
	}

	return nil, nil
}

//=============================================================================

func AddInstrument(tx *gorm.DB, i *Instrument) error {
	return tx.Create(i).Error
}

//=============================================================================

func UpdateInstrument(tx *gorm.DB, i *Instrument) error {
	return tx.Save(i).Error
}

//=============================================================================

func AddUploadJob(tx *gorm.DB, job *UploadJob) error {
	return tx.Create(job).Error
}

//=============================================================================

func UpdateUploadJob(tx *gorm.DB, job *UploadJob) error {
	return tx.Save(job).Error
}

//=============================================================================
