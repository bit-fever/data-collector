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

package ds

import (
	"math"
	"time"
)

//=============================================================================

type TimeSlotFunction func(t time.Time) time.Time

//=============================================================================

type DataAggregator struct {
	currDp       *DataPoint
	dataPoints   []*DataPoint
	timeSlotFunc TimeSlotFunction
}

//=============================================================================

func NewDataAggregator(f TimeSlotFunction) *DataAggregator {
	da := &DataAggregator{}
	da.dataPoints   = []*DataPoint{}
	da.timeSlotFunc = f
	return da
}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func (a *DataAggregator) Add(dp *DataPoint) {
	//--- Handle the no aggregation case

	if a.timeSlotFunc == nil {
		a.dataPoints = append(a.dataPoints, dp)
		return
	}

	//--- Aggregation required

	if a.currDp == nil {
		a.currDp = a.createInitialDataPoint(dp)
	} else {
		if a.currDp.Time.Equal(a.timeSlotFunc(dp.Time)) {
			a.Merge(dp)
		} else {
			a.dataPoints = append(a.dataPoints, a.currDp)
			a.currDp = a.createInitialDataPoint(dp)
		}
	}
}

//=============================================================================

func (a *DataAggregator) Flush() {
	if a.currDp != nil {
		a.dataPoints = append(a.dataPoints, a.currDp)
		a.currDp     = nil
	}
}

//=============================================================================

func (a *DataAggregator) DataPoints() []*DataPoint {
	return a.dataPoints
}

//=============================================================================

func (a *DataAggregator) Aggregate(daDes *DataAggregator) {
	for _,dp := range a.DataPoints() {
		daDes.Add(dp)
	}

	daDes.Flush()
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func (a *DataAggregator) createInitialDataPoint(dp *DataPoint) *DataPoint {
	return &DataPoint{
		Time      : a.timeSlotFunc(dp.Time),
		Open      : dp.Open,
		High      : dp.High,
		Low       : dp.Low,
		Close     : dp.Close,
		UpVolume  : dp.UpVolume,
		DownVolume: dp.DownVolume,
	}
}

//=============================================================================

func (a *DataAggregator) Merge(dp *DataPoint) {
	cp := a.currDp

	cp.High        = math.Max(cp.High, dp.High)
	cp.Low         = math.Min(cp.Low,  dp.Low)
	cp.Close       = dp.Close
	cp.UpVolume   += dp.UpVolume
	cp.DownVolume += dp.DownVolume
}

//=============================================================================

func TimeSlotFunction5m(dpTime time.Time) time.Time {
	mins := dpTime.Minute()

	if mins ==  0 { return dpTime }
	if mins <=  5 { return dpTime.Add(time.Minute * time.Duration( 5-mins)) }
	if mins <= 10 { return dpTime.Add(time.Minute * time.Duration(10-mins)) }
	if mins <= 15 { return dpTime.Add(time.Minute * time.Duration(15-mins)) }
	if mins <= 20 { return dpTime.Add(time.Minute * time.Duration(20-mins)) }
	if mins <= 25 { return dpTime.Add(time.Minute * time.Duration(25-mins)) }
	if mins <= 30 { return dpTime.Add(time.Minute * time.Duration(30-mins)) }
	if mins <= 35 { return dpTime.Add(time.Minute * time.Duration(35-mins)) }
	if mins <= 40 { return dpTime.Add(time.Minute * time.Duration(40-mins)) }
	if mins <= 45 { return dpTime.Add(time.Minute * time.Duration(45-mins)) }
	if mins <= 50 { return dpTime.Add(time.Minute * time.Duration(50-mins)) }
	if mins <= 55 { return dpTime.Add(time.Minute * time.Duration(55-mins)) }

	return dpTime.Add(time.Minute * time.Duration(60-mins))
}

//=============================================================================

func TimeSlotFunction10m(dpTime time.Time) time.Time {
	mins := dpTime.Minute()

	if mins ==  0 { return dpTime }
	if mins <= 10 { return dpTime.Add(time.Minute * time.Duration(10-mins)) }
	if mins <= 20 { return dpTime.Add(time.Minute * time.Duration(20-mins)) }
	if mins <= 30 { return dpTime.Add(time.Minute * time.Duration(30-mins)) }
	if mins <= 40 { return dpTime.Add(time.Minute * time.Duration(40-mins)) }
	if mins <= 50 { return dpTime.Add(time.Minute * time.Duration(50-mins)) }

	return dpTime.Add(time.Minute * time.Duration(60-mins))
}

//=============================================================================

func TimeSlotFunction15m(dpTime time.Time) time.Time {
	mins := dpTime.Minute()

	if mins ==  0 { return dpTime }
	if mins <= 15 { return dpTime.Add(time.Minute * time.Duration(15-mins)) }
	if mins <= 30 { return dpTime.Add(time.Minute * time.Duration(30-mins)) }
	if mins <= 45 { return dpTime.Add(time.Minute * time.Duration(45-mins)) }

	return dpTime.Add(time.Minute * time.Duration(60-mins))
}

//=============================================================================

func TimeSlotFunction30m(dpTime time.Time) time.Time {
	mins := dpTime.Minute()

	if mins ==  0 { return dpTime }
	if mins <= 30 { return dpTime.Add(time.Minute * time.Duration(30-mins)) }

	return dpTime.Add(time.Minute * time.Duration(60-mins))
}

//=============================================================================

func TimeSlotFunction60m(dpTime time.Time) time.Time {
	mins := dpTime.Minute()

	if mins ==  0 { return dpTime }

	return dpTime.Add(time.Minute * time.Duration(60-mins))
}

//=============================================================================
