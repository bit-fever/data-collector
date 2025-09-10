//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

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

package jobmanager

import (
	"log/slog"
	"time"

	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/data-collector/pkg/core/messaging/rollover"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

type Job interface {
	execute(jc *JobContext) error
}

//=============================================================================

type JobContext struct {
	userConnection *UserConnection
	cache          *AdapterCache
	resuming       bool
	sleeping       bool
}

//=============================================================================

func NewJobContext(uc *UserConnection, cache *AdapterCache, resuming bool) *JobContext {
	return &JobContext{
		userConnection: uc,
		cache         : cache,
		resuming      : resuming,
	}
}

//=============================================================================

func (jc *JobContext) GoToSleep() {
	jc.sleeping = true
}

//=============================================================================

func (jc *JobContext) UpdateJob(blkStatus db.DBStatus, jobStatus db.DJStatus, jobErr string, sendRecalcMsg bool) error {
	return db.RunInTransaction(func(tx *gorm.DB) error {
		blk := jc.userConnection.scheduledJob.block
		job := jc.userConnection.scheduledJob.job

		oldBlkStatus := blk.Status
		oldJobStatus := job.Status

		blk.Status = blkStatus
		job.Status = jobStatus
		job.Error  = jobErr

		err := db.UpdateDataBlock(tx, blk)
		if err == nil {
			err = db.UpdateDownloadJob(tx, job)
			if err == nil && sendRecalcMsg {
				err = jc.sendRollRecalcMessage()
			}
		}

		if err != nil {
			blk.Status = oldBlkStatus
			job.Status = oldJobStatus
		}

		return err
	})
}

//=============================================================================

func (jc *JobContext) EndJob() error {
	err := db.RunInTransaction(func(tx *gorm.DB) error {
		blk := jc.userConnection.scheduledJob.block
		job := jc.userConnection.scheduledJob.job

		oldBlk := *blk

		blk.Status = db.DBStatusReady

		if blk.DataFrom.IsNil() || blk.DataTo.IsNil() {
			blk.Status = db.DBStatusEmpty
		}

		err := db.UpdateDataBlock(tx, blk)
		if err == nil {
			err = db.DeleteDownloadJob(tx, job.Id)
			if err == nil {
				err = jc.sendRollRecalcMessage()
			}
		}

		if err != nil {
			blk.Status   = oldBlk.Status
			blk.DataFrom = oldBlk.DataFrom
			blk.DataTo   = oldBlk.DataTo
			job.UserConnection = ""
		}

		return err
	})

	jc.cache.freeConnection(jc.userConnection, err != nil)

	return err
}

//=============================================================================

func (jc *JobContext) AbortJob(jobErr string) error {
	jc.userConnection.scheduledJob.job.UserConnection = ""
	err := jc.UpdateJob(db.DBStatusError, db.DJStatusError, jobErr, false)

	now := time.Now()
	jc.userConnection.scheduledJob.lastError = &now

	jc.cache.freeConnection(jc.userConnection,true)

	return err
}

//=============================================================================

func (jc *JobContext) SleepJob() error {
	jc.userConnection.scheduledJob.job.UserConnection = ""
	err := jc.UpdateJob(db.DBStatusSleeping, db.DJStatusWaiting, "", true)

	jc.cache.freeConnection(jc.userConnection,true)

	return err
}

//=============================================================================

func (jc *JobContext) sendRollRecalcMessage() error {
	sj := jc.userConnection.scheduledJob

	job := &rollover.RecalcJob{
		DataBlockId: sj.block.Id,
	}

	err := msg.SendMessage(msg.ExCollector, msg.SourceRollRecalcJob, msg.TypeCreate, job)

	if err != nil {
		slog.Error("sendRollRecalcJobMessage: Could not publish the upload message", "error", err.Error())
	}

	return err
}

//=============================================================================
