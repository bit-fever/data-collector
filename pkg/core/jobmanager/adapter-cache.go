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
	"sync"

	"github.com/bit-fever/data-collector/pkg/db"
)

//=============================================================================

type AdapterCache struct {
	sync.RWMutex
	systemCode  string
	products    map[string]*ProductCache
	connections map[string]*UserConnection
	waitingJobs []*ScheduledJob
	runningJobs []*ScheduledJob
	inErrorJobs []*ScheduledJob
}

//=============================================================================

func NewAdapterCache(systemCode string) *AdapterCache {
	return &AdapterCache{
		systemCode :  systemCode,
		products   : make(map[string]*ProductCache),
		connections: make(map[string]*UserConnection),
	}
}

//=============================================================================
//===
//=== API methods
//===
//=============================================================================

func (ac *AdapterCache) getDataBlock(root string, symbol string) *db.DataBlock {
	ac.RLock()
	pc, found := ac.products[root]
	ac.RUnlock()

	if found {
		return pc.getDataBlock(symbol)
	}

	return nil
}

//=============================================================================

func (ac *AdapterCache) addDataBlock(db *db.DataBlock) {
	ac.Lock()

	pc, found := ac.products[db.Root]
	if !found {
		pc = NewProductCache(db.Root)
		ac.products[db.Root] = pc
	}

	ac.Unlock()
	pc.addDataBlock(db)
}

//=============================================================================

func (ac *AdapterCache) addScheduledJob(sj *ScheduledJob, r Resumer) {
	ac.Lock()

	root := sj.block.Root
	pc, found := ac.products[root]
	if !found {
		pc = NewProductCache(root)
		ac.products[root] = pc
	}

	if sj.job.Status == db.DJStatusWaiting {
		ac.waitingJobs = append(ac.waitingJobs, sj)
	} else if sj.job.Status == db.DJStatusRunning {
		uc,ok := ac.connections[sj.job.UserConnection]
		if ok {
			if uc.connected {
				ac.runningJobs = append(ac.runningJobs, sj)
				uc.allocateToJob(sj)
				//--- Resume job
				r(ac,uc)
			} else {
				//--- Job was running but now the connection is down
				sj.job.UserConnection = ""
				ac.waitingJobs = append(ac.waitingJobs, sj)
			}
		} else {
			blk := sj.block
			slog.Error("Fatal: UserConnection not found. Discarding job", "systemCode", blk.SystemCode, "root", blk.Root, "symbol", blk.Symbol, "jobId", sj.job.Id)
		}
	} else if sj.job.Status == db.DJStatusError {
		ac.inErrorJobs = append(ac.inErrorJobs, sj)
	}

	ac.Unlock()
	pc.addDataBlock(sj.block)
}

//=============================================================================

func (ac *AdapterCache) setConnection(username, connCode string, connected bool) {
	uc := newUserConnection(username,connCode,connected)

	ac.Lock()

	oldUc,found := ac.connections[uc.key()]
	if !found {
		ac.connections[uc.key()] = uc
	} else {
		//--- If we don't create new UserConnection objects we preserve pointers inside the data structure
		oldUc.connected = connected
	}

	ac.Unlock()
}

//=============================================================================

func (ac *AdapterCache) disconnectAll() {
	ac.RLock()

	for _, uc := range ac.connections {
		uc.connected = false
	}

	ac.RUnlock()
}

//=============================================================================

func (ac *AdapterCache) schedule(maxJobs int, e Executor) int {
	ac.Lock()
	defer ac.Unlock()

	maxJobs = maxJobs - len(ac.runningJobs)

	for _, uc := range ac.connections {
		if maxJobs > 0 {
			idx,job := ac.getJobToRun()
			if !uc.isAllocated() && uc.connected && job != nil {
				uc.allocateToJob(job)
				if !e(ac, uc) {
					//--- If the executor cannot start the job, let's abort the entire process
					return 0
				}

				ac.runningJobs = append  (ac.runningJobs, job)
				ac.waitingJobs = removeAt(ac.waitingJobs, idx)
				maxJobs--
			}
		} else {
			break
		}
	}

	return maxJobs
}

//=============================================================================

func (ac *AdapterCache) freeConnection(uc *UserConnection, jobInSleeping bool, jobInError bool) {
	ac.Lock()
	defer ac.Unlock()

	ac.runningJobs = removeElem(ac.runningJobs, uc.scheduledJob)

	if jobInSleeping {
		ac.waitingJobs = append(ac.waitingJobs, uc.scheduledJob)
	}

	if jobInError {
		ac.inErrorJobs = append(ac.inErrorJobs, uc.scheduledJob)
	}

	uc.deallocate()
}

//=============================================================================

func (ac *AdapterCache) getJobToRun() (int,*ScheduledJob) {
	var job *ScheduledJob
	var idx int

	for i, sj := range ac.waitingJobs {
		if sj.IsSchedulable() {
			if job == nil {
				job = sj
				idx = i
			} else {
				if job.job.Priority < sj.job.Priority {
					job = sj
					idx = i
				}
			}
		}
	}

	return idx,job
}

//=============================================================================
//===
//=== General methods
//===
//=============================================================================

func removeAt(s []*ScheduledJob, i int) []*ScheduledJob {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

//=============================================================================

func removeElem(s []*ScheduledJob, sj *ScheduledJob) []*ScheduledJob {
	n := 0
	for _, x := range s {
		if x!=sj {
			s[n] = x
			n++
		}
	}

	return s[:n]
}

//=============================================================================
