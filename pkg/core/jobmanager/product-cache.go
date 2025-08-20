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

type ProductCache struct {
	sync.RWMutex
	root      string
	blocksMap map[string]*db.DataBlock
}

//=============================================================================

func NewProductCache(root string) *ProductCache {
	return &ProductCache{
		root     : root,
		blocksMap: make(map[string]*db.DataBlock),
	}
}

//=============================================================================
//===
//=== API methods
//===
//=============================================================================

func (pc *ProductCache) getDataBlock(symbol string) *db.DataBlock {
	pc.RLock()
	i, found := pc.blocksMap[symbol]
	pc.RUnlock()

	if found {
		return i
	}

	return nil
}

//=============================================================================

func (pc *ProductCache) addDataBlock(db *db.DataBlock) {
	pc.Lock()
	pc.blocksMap[db.Symbol] = db
	pc.Unlock()
}

//=============================================================================
