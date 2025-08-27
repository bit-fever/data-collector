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

	"github.com/bit-fever/data-collector/pkg/db"
)

//=============================================================================

const (
	MaxJobs = 4
)

type Executor func(ac *AdapterCache, uc *UserConnection) bool
type Resumer  func(ac *AdapterCache, uc *UserConnection)

//=============================================================================

var ticker *time.Ticker

//=============================================================================

func startScheduler() {
	ticker = time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			run()
		}
	}()
}

//=============================================================================

func run() {
	cache.schedule(MaxJobs, executor)
}

//=============================================================================

func executor(ac *AdapterCache, uc *UserConnection) bool {
	jc := NewJobContext(uc, ac, false)
	err := jc.UpdateJob(db.DBStatusLoading, db.DJStatusRunning, "")
	if err == nil {
		go func() {
			runJob(jc)
		}()
	}

	return err == nil
}

//=============================================================================

func resumer(ac *AdapterCache, uc *UserConnection) {
	go func() {
		//--- Let's wait some time before resuming jobs to allow the boot sequence to complete
		time.Sleep(5 * time.Second)
		jc := NewJobContext(uc, ac, true)
		runJob(jc)
	}()
}

//=============================================================================

func runJob(jc *JobContext) {
	job := &InstrumentDownLoadJob{}
	err := job.execute(jc)

	if err == nil {
		if jc.sleeping {
			slog.Info("DownloadJob: Job sent in sleeping status", "error", err, "jobId", jc.userConnection.scheduledJob.job.Id)
			err = jc.SleepJob()
		} else {
			err = jc.EndJob()
		}
	} else {
		slog.Error("DownloadJob: Encountered an error. Operation was aborted", "error", err, "jobId", jc.userConnection.scheduledJob.job.Id)
		err = jc.AbortJob(err.Error())
	}

	if err != nil {
		slog.Error("Fatal: Cannot end/abort/sleep a job", "jobId", jc.userConnection.scheduledJob.job.Id, "error", err)
	}
}

//=============================================================================
