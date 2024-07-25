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

func GetInstrumentsBySourceId(tx *gorm.DB, c *auth.Context, sourceId uint)(*[]db.Instrument, error) {
	return db.GetInstrumentsBySourceId(tx, sourceId)
}

//=============================================================================

func AddInstrumentAndJob(tx *gorm.DB, c *auth.Context, sourceId uint, spec *DatafileUploadSpec, filename string, bytes int64) error {
	c.Log.Info("PrepareUploadTask: Creating instrument for a product", "sourceId", sourceId, "symbol", spec.Symbol)

	p, err := getProductAndCheckAccess(tx, c, sourceId, "AddInstrumentAndJob")
	if err != nil {
		return err
	}

	i, err := db.GetInstrumentBySymbol(tx, p.Id, spec.Symbol)
	if err != nil {
		return err
	}

	if i == nil {
		i = &db.Instrument{
			ProductId   : p.Id,
			Symbol      : spec.Symbol,
			Name        : spec.Name,
			IsContinuous: spec.Continuous,
			Status      : db.InstrumentStatusProcessing,
		}

		err = db.AddInstrument(tx, i)
		if err != nil {
			return err
		}
	} else {
		i.Name         = spec.Name
		i.IsContinuous = spec.Continuous
		i.Status       = db.InstrumentStatusProcessing

		err = db.UpdateInstrument(tx, i)
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
		InstrumentId: i.Id,
		Status      : db.UploadJobStatusWaiting,
		Filename    : filename,
		Bytes       : bytes,
		Timezone    : timezone,
		Parser      : spec.Parser,
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

func getProductAndCheckAccess(tx *gorm.DB, c *auth.Context, sourceId uint, function string) (*db.Product, error) {
	p, err := db.GetProductBySourceId(tx, sourceId)

	if err != nil {
		c.Log.Error(function +": Could not retrieve product", "error", err.Error())
		return nil, err
	}

	if p == nil {
		c.Log.Error(function +": Product was not found", "sourceId", sourceId)
		return nil, req.NewNotFoundError("Product was not found: %v", sourceId)
	}

	if ! c.Session.IsAdmin() {
		if p.Username != c.Session.Username {
			c.Log.Error(function +": Product not owned by user", "sourceId", sourceId)
			return nil, req.NewForbiddenError("Product is not owned by user: %v", sourceId)
		}
	}

	return p, nil
}

//=============================================================================

func sendIngestJobMessage(c *auth.Context, job *db.UploadJob) error {
	err := msg.SendMessage(msg.ExCollectorUpload, msg.OriginDb, msg.TypeCreate, msg.SourceUploadJob, job)

	if err != nil {
		c.Log.Error("sendIngestJobMessage: Could not publish the upload message", "error", err.Error())
		return err
	}

	return nil
}

//=============================================================================

func calcTimezone(fileTimezone string, p *db.Product) (string, error){
	if fileTimezone == "utc" {
		return "utc",      nil
	}
	if fileTimezone == "exc" {
		return p.Timezone, nil
	}

	return "", errors.New("Unknown file timezone: "+ fileTimezone)
}

//=============================================================================
