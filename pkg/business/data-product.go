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
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

func GetDataInstrumentsByProductId(tx *gorm.DB, c *auth.Context, productId uint)(*[]db.DataInstrument, error) {
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

	if i == nil {
		i = &db.DataInstrument{
			DataProductId: p.Id,
			Symbol       : spec.Symbol,
			Name         : spec.Name,
			IsContinuous : spec.Continuous,
			Status       : db.InstrumentStatusProcessing,
		}

		err = db.AddDataInstrument(tx, i)
		if err != nil {
			return err
		}
	} else {
		i.Name         = spec.Name
		i.IsContinuous = spec.Continuous
		i.Status       = db.InstrumentStatusProcessing

		err = db.UpdateDataInstrument(tx, i)
		if err != nil {
			return err
		}
	}

	timezone,err := calcTimezone(spec.FileTimezone, p)
	if err != nil {
		return err
	}

	//--- Add upload job

	job := &db.UploadJob{
		DataInstrumentId: i.Id,
		Status          : db.UploadJobStatusWaiting,
		Filename        : filename,
		Bytes           : bytes,
		Timezone        : timezone,
		Parser          : spec.Parser,
	}

	err = db.AddUploadJob(tx, job)
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

func sendIngestJobMessage(c *auth.Context, job *db.UploadJob) error {
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
