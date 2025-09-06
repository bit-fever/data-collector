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

func GetDataInstrumentsByProductId(tx *gorm.DB, pId uint) (*[]DataInstrumentExt, error) {
	var list []DataInstrumentExt

	filter := map[string]any{}
	filter["data_product_id"] = pId

	res := tx.
		Select("data_instrument.*, db.status, db.data_from, db.data_to, db.progress ").
		Joins("LEFT JOIN data_block db ON db.id = data_block_id").
		Where(filter).
		Order("symbol").
		Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetRollingDataInstrumentsByProductId(tx *gorm.DB, pId uint) (*[]DataInstrumentExt, error) {
	var list []DataInstrumentExt

	filter := map[string]any{}
	filter["data_product_id"] = pId
	filter["continuous"]      = 0

	res := tx.
		Select("data_instrument.*, db.status, db.data_from, db.data_to, db.progress ").
		Joins("JOIN data_block db ON db.id = data_block_id").
		Where(filter).
		Order("expiration_date").
		Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetDataInstrumentsFull(tx *gorm.DB, filter map[string]any) (*[]DataInstrumentFull, error) {
	var list []DataInstrumentFull

	res := tx.
		Select("data_instrument.*, dp.symbol as product_symbol, dp.system_code, dp.connection_code").
		Joins("JOIN data_product dp ON dp.id = data_product_id").
		Where(filter).
		Order("name").
		Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetDataInstrumentById(tx *gorm.DB, id uint) (*DataInstrument, error) {
	var list []DataInstrument
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

func GetDataInstrumentBySymbol(tx *gorm.DB, productId uint, symbol string) (*DataInstrument, error) {
	filter := map[string]any{}
	filter["data_product_id"] = productId
	filter["symbol"]          = symbol

	var list []DataInstrument
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

func AddDataInstrument(tx *gorm.DB, i *DataInstrument) error {
	return tx.Create(i).Error
}

//=============================================================================

func UpdateDataInstrument(tx *gorm.DB, i *DataInstrument) error {
	return tx.Save(i).Error
}

//=============================================================================
