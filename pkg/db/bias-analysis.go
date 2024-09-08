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

package db

import (
	"github.com/bit-fever/core/req"
	"gorm.io/gorm"
)

//=============================================================================

func GetBiasAnalyses(tx *gorm.DB, filter map[string]any, offset int, limit int) (*[]BiasAnalysisFull, error) {
	var list []BiasAnalysisFull
	res := tx.Where(filter).Offset(offset).Limit(limit).Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetBiasAnalysesFull(tx *gorm.DB, filter map[string]any, offset int, limit int) (*[]BiasAnalysisFull, error) {
	var list []BiasAnalysisFull
	query :=
		"SELECT ba.*, di.symbol as data_symbol, di.name as data_name, bp.symbol as broker_symbol, bp.name as broker_name " +
		"FROM bias_analysis ba " +
		"LEFT JOIN data_instrument di on ba.data_instrument_id = di.id " +
		"LEFT JOIN broker_product bp on ba.broker_product_id = bp.id"

	res := tx.Raw(query).Where(filter).Offset(offset).Limit(limit).Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func GetBiasAnalysisById(tx *gorm.DB, id uint) (*BiasAnalysis, error) {
	var list []BiasAnalysis
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

func AddBiasAnalysis(tx *gorm.DB, ba *BiasAnalysis) error {
	return tx.Create(ba).Error
}

//=============================================================================

func UpdateBiasAnalysis(tx *gorm.DB, ba *BiasAnalysis) error {
	return tx.Save(ba).Error
}

//=============================================================================
//=== Bias configs
//=============================================================================

func GetBiasConfigsByAnalysisId(tx *gorm.DB, id uint) (*[]BiasConfig, error) {
	var list []BiasConfig

	filter := map[string]any{}
	filter["bias_analysis_id"] = id

	res := tx.Where(filter).Order("start_day").Find(&list)

	if res.Error != nil {
		return nil, req.NewServerErrorByError(res.Error)
	}

	return &list, nil
}

//=============================================================================

func AddBiasConfig(tx *gorm.DB, bc *BiasConfig) error {
	return tx.Create(bc).Error
}

//=============================================================================

func UpdateBiasConfig(tx *gorm.DB, bc *BiasConfig) error {
	return tx.Save(bc).Error
}

//=============================================================================

func DeleteBiasConfig(tx *gorm.DB, id uint) error {
	return tx.Delete(&BiasConfig{}, id).Error
}

//=============================================================================
