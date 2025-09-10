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
	"sync"

	"github.com/bit-fever/data-collector/pkg/db"
)

//=============================================================================

type InventoryCache struct {
	sync.RWMutex
	adapters map[string]*AdapterCache
}

//=============================================================================

func newInventoryCache() *InventoryCache {
	ic := InventoryCache{
		adapters: make(map[string]*AdapterCache),
	}

	return &ic
}

//=============================================================================
//===
//=== API methods
//===
//=============================================================================

func (ic *InventoryCache) getDataBlock(systemCode, root, symbol string) *db.DataBlock {
	ic.RLock()
	ac, found := ic.adapters[systemCode]
	ic.RUnlock()

	if found {
		return ac.getDataBlock(root, symbol)
	}

	return nil
}

//=============================================================================

func (ic *InventoryCache) addDataBlock(db *db.DataBlock) {
	ac := ic.getOrCreate(db.SystemCode)
	ac.addDataBlock(db)
}

//=============================================================================

func (ic *InventoryCache) addScheduledJob(sj *ScheduledJob) {
	ac := ic.getOrCreate(sj.block.SystemCode)
	ac.addScheduledJob(sj)
}

//=============================================================================

func (ic *InventoryCache) setConnection(systemCode, username, connCode string, connected bool) {
	ac := ic.getOrCreate(systemCode)
	ac.setConnection(username, connCode, connected)
}

//=============================================================================

func (ic *InventoryCache) disconnectAll() {
	ic.RLock()

	for _,ac := range ic.adapters {
		ac.disconnectAll()
	}

	ic.RUnlock()
}

//=============================================================================

func (ic *InventoryCache) schedule(maxJobs int, e Executor) {
	ic.RLock()
	ic.RUnlock()

	for _,ac := range ic.adapters {
		maxJobs = ac.schedule(maxJobs, e)

		if maxJobs == 0 {
			return
		}
	}
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func (ic *InventoryCache) getOrCreate(systemCode string) *AdapterCache {
	ic.Lock()

	ac, found := ic.adapters[systemCode]
	if !found {
		ac = NewAdapterCache(systemCode)
		ic.adapters[systemCode] = ac
	}

	ic.Unlock()
	return ac
}

//=============================================================================
