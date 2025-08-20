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
	"strconv"

	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

func GetDataInstrumentsByProductId(tx *gorm.DB, c *auth.Context, productId uint)(*[]db.DataInstrumentExt, error) {
	return db.GetDataInstrumentsByProductId(tx, productId)
}

//=============================================================================

func AddDataInstrumentAndJob(tx *gorm.DB, c *auth.Context, productId uint, spec *DatafileUploadSpec, filename string, bytes int64) error {
	c.Log.Info("AddDataInstrumentAndJob: Creating instrument for a data product", "dataProductId", productId, "symbol", spec.Symbol)

	p, err := getDataProductAndCheckAccess(tx, c, productId, "AddDataInstrumentAndJob")
	if err != nil {
		return err
	}

	i, err := db.GetDataInstrumentBySymbol(tx, p.Id, spec.Symbol)
	if err != nil {
		return err
	}

	var b *db.DataBlock

	if i == nil {
		i,b,err = createDataInstrument(tx, p, spec)
	} else {
		b,err = updateDataInstrument(tx, i, spec)
	}

	timezone,err := calcTimezone(spec.FileTimezone, p)
	if err != nil {
		return err
	}

	//--- Add upload job

	job := &db.IngestionJob{
		DataInstrumentId: i.Id,
		DataBlockId     : b.Id,
		Filename        : filename,
		Bytes           : bytes,
		Timezone        : timezone,
		Parser          : spec.Parser,
	}

	err = db.AddIngestionJob(tx, job)
	if err != nil {
		return err
	}

	return sendIngestJobMessage(c, job)
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func getDataProductAndCheckAccess(tx *gorm.DB, c *auth.Context, id uint, function string) (*db.DataProduct, error) {
	p, err := db.GetDataProductById(tx, id)

	if err != nil {
		c.Log.Error(function +": Could not retrieve data product", "error", err.Error())
		return nil, err
	}

	if p == nil {
		c.Log.Error(function +": Data product was not found", "id", id)
		return nil, req.NewNotFoundError("Data product was not found: %v", id)
	}

	if ! c.Session.IsAdmin() {
		if p.Username != c.Session.Username {
			c.Log.Error(function +": Data product not owned by user", "id", id)
			return nil, req.NewForbiddenError("Data product is not owned by user: %v", id)
		}
	}

	return p, nil
}

//=============================================================================

func createDataInstrument(tx *gorm.DB, p *db.DataProduct, spec *DatafileUploadSpec) (*db.DataInstrument, *db.DataBlock, error) {
	//--- Add its associated DataBlock
	b := &db.DataBlock{
		SystemCode: p.SystemCode,
		Root      : p.Symbol,
		Symbol    : spec.Symbol,
		Status    : db.DBStatusWaiting,
		Global    : false,
		Progress  : 0,
	}

	err := db.AddDataBlock(tx, b)
	if err != nil {
		return nil,nil,err
	}

	//--- Add a new DataInstrument

	i := &db.DataInstrument{
		DataProductId: p.Id,
		DataBlockId  : &b.Id,
		Symbol       : spec.Symbol,
		Name         : spec.Name,
		Continuous   : false,
	}

	err = db.AddDataInstrument(tx, i)

	return i,b,err
}

//=============================================================================

func updateDataInstrument(tx *gorm.DB, i *db.DataInstrument, spec *DatafileUploadSpec) (*db.DataBlock, error) {
	i.Name = spec.Name

	err := db.UpdateDataInstrument(tx, i)
	if err != nil {
		return nil,err
	}

	var b *db.DataBlock
	b,err = db.GetDataBlockById(tx, *i.DataBlockId)
	if err != nil {
		return nil,err
	}

	if b == nil {
		return nil, errors.New("Panic: DataBlock was not found! --> id="+ strconv.Itoa(int(*i.DataBlockId)))
	}

	b.Status   = db.DBStatusWaiting
	b.Progress = 0

	err = db.UpdateDataBlock(tx, b)

	return b,err
}

//=============================================================================

func sendIngestJobMessage(c *auth.Context, job *db.IngestionJob) error {
	err := msg.SendMessage(msg.ExCollector, msg.SourceUploadJob, msg.TypeCreate, job)

	if err != nil {
		c.Log.Error("sendIngestJobMessage: Could not publish the upload message", "error", err.Error())
		return err
	}

	return nil
}

//=============================================================================

func calcTimezone(fileTimezone string, p *db.DataProduct) (string, error){
	if fileTimezone == "utc" {
		return "utc",      nil
	}
	if fileTimezone == "exc" {
		return p.Timezone, nil
	}

	return "", errors.New("Unknown file timezone: "+ fileTimezone)
}

//=============================================================================
