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

package business

import (
	"errors"
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

func GetBiasConfigsByAnalysisId(tx *gorm.DB, c *auth.Context, baId uint) (*[]*BiasConfig, error) {
	list, err := db.GetBiasConfigsByAnalysisId(tx, baId)

	if err != nil {
		return nil, err
	}

	var result []*BiasConfig

	for _, dbc := range *list {
		bc := &BiasConfig{}
		bc.FromBiasConfig(&dbc)
		result = append(result, bc)
	}

	return &result, nil
}

//=============================================================================

func AddBiasConfig(tx *gorm.DB, c *auth.Context, baId uint, bcs *BiasConfigSpec) (*db.BiasConfig, error) {
	c.Log.Info("AddBiasConfig: Adding a new bias config", "baId", baId)

	if err:=checkBiasConfigSpec(c, bcs); err != nil {
		return nil, err
	}

	bc := bcs.ToBiasConfig()
	bc.BiasAnalysisId = baId
	err := db.AddBiasConfig(tx, bc)

	if err != nil {
		c.Log.Error("AddBiasConfig: Could not add a new bias config", "error", err.Error())
		return nil, err
	}

	c.Log.Info("AddBiasConfig: Bias config added", "baId", baId, "id", bc.Id)
	return bc, err
}

//=============================================================================

func UpdateBiasConfig(tx *gorm.DB, c *auth.Context, baId uint, id uint, bcs *BiasConfigSpec) (*db.BiasConfig, error) {
	c.Log.Info("UpdateBiasConfig: Updating a bias config", "id", id, "baId", baId)

	if err:=checkBiasConfigSpec(c, bcs); err != nil {
		return nil, err
	}

	bc := bcs.ToBiasConfig()
	bc.Id             = id
	bc.BiasAnalysisId = baId

	err := db.UpdateBiasConfig(tx, bc)
	if err != nil {
		c.Log.Error("UpdateBiasConfig: Could not update a bias config", "error", err.Error())
		return nil, err
	}

	c.Log.Info("UpdateBiasConfig: Bias config updated", "id", bc.Id, "baId", bc.BiasAnalysisId)
	return bc, err
}

//=============================================================================

func DeleteBiasConfig(tx *gorm.DB, c *auth.Context, baId uint, id uint) (bool, error) {
	c.Log.Info("DeleteBiasConfig: Deleting a bias config", "id", id, "baId", baId)

	err := db.DeleteBiasConfig(tx, id)
	if err != nil {
		c.Log.Error("DeleteBiasConfig: Could not delete a bias config", "error", err.Error())
		return false, err
	}

	c.Log.Info("DeleteBiasConfig: Bias config deleted", "id", id, "baId", baId)
	return true, err
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func checkBiasConfigSpec(c *auth.Context, bcs *BiasConfigSpec) error {
	if bcs.Operation != 0 && bcs.Operation != 1 {
		err := errors.New("operation can only be 0 (for long) or 1 (for short)")
		c.Log.Error("checkBiasConfigSpec: Invalid bias config spec", "error", err.Error())
		return err
	}

	return nil
}

//=============================================================================
