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

package ds

import (
	"testing"
	"time"
)

var hourly = []DataPoint{
	{Time:p("2021-11-30T01:00:00+00:00"),Open:4663,   High:4666.75,Low:4660.5, Close:4665.75,UpVolume:4504,  DownVolume:3388,  UpTicks:3209,  DownTicks:2640,  OpenInterest:0},
	{Time:p("2021-11-30T02:00:00+00:00"),Open:4666,   High:4667.5, Low:4663.5, Close:4664.5, UpVolume:3348,  DownVolume:3593,  UpTicks:2593,  DownTicks:2788,  OpenInterest:0},
	{Time:p("2021-11-30T03:00:00+00:00"),Open:4664.5, High:4665.5, Low:4656.75,Close:4661,   UpVolume:3727,  DownVolume:4506,  UpTicks:3006,  DownTicks:3486,  OpenInterest:0},
	{Time:p("2021-11-30T04:00:00+00:00"),Open:4660.75,High:4661.25,Low:4653.5, Close:4657.5, UpVolume:2931,  DownVolume:3051,  UpTicks:2411,  DownTicks:2450,  OpenInterest:0},
	{Time:p("2021-11-30T05:00:00+00:00"),Open:4657.5, High:4660,   Low:4653.25,Close:4656.75,UpVolume:2471,  DownVolume:3188,  UpTicks:2124,  DownTicks:2600,  OpenInterest:0},
	{Time:p("2021-11-30T06:00:00+00:00"),Open:4656.5, High:4657,   Low:4606.5, Close:4608.75,UpVolume:21063, DownVolume:27653, UpTicks:17096, DownTicks:20789, OpenInterest:0},
	{Time:p("2021-11-30T07:00:00+00:00"),Open:4608.75,High:4616.5, Low:4582,   Close:4612,   UpVolume:20264, DownVolume:19959, UpTicks:16676, DownTicks:15902, OpenInterest:0},
	{Time:p("2021-11-30T08:00:00+00:00"),Open:4612,   High:4620.25,Low:4603.25,Close:4611.75,UpVolume:16656, DownVolume:14612, UpTicks:12717, DownTicks:12377, OpenInterest:0},
	{Time:p("2021-11-30T09:00:00+00:00"),Open:4611.75,High:4612.25,Low:4595.25,Close:4604,   UpVolume:22645, DownVolume:20515, UpTicks:17148, DownTicks:16005, OpenInterest:0},
	{Time:p("2021-11-30T10:00:00+00:00"),Open:4603.75,High:4604.5, Low:4588.5, Close:4604.25,UpVolume:19422, DownVolume:17861, UpTicks:14200, DownTicks:14281, OpenInterest:0},
	{Time:p("2021-11-30T11:00:00+00:00"),Open:4604,   High:4607.5, Low:4595,   Close:4605.5, UpVolume:13609, DownVolume:12503, UpTicks:10368, DownTicks:9768,  OpenInterest:0},
	{Time:p("2021-11-30T12:00:00+00:00"),Open:4605.5, High:4609.75,Low:4597.5, Close:4604.25,UpVolume:15645, DownVolume:15416, UpTicks:12213, DownTicks:11757, OpenInterest:0},
	{Time:p("2021-11-30T13:00:00+00:00"),Open:4604,   High:4619.5, Low:4603,   Close:4610.75,UpVolume:24786, DownVolume:20003, UpTicks:15376, DownTicks:13792, OpenInterest:0},
	{Time:p("2021-11-30T14:00:00+00:00"),Open:4610.75,High:4626,   Low:4603.75,Close:4625,   UpVolume:31575, DownVolume:26873, UpTicks:20958, DownTicks:18151, OpenInterest:0},
	{Time:p("2021-11-30T15:00:00+00:00"),Open:4625,   High:4633.75,Low:4614.5, Close:4628,   UpVolume:138733,DownVolume:129681,UpTicks:92255, DownTicks:87645, OpenInterest:0},
	{Time:p("2021-11-30T16:00:00+00:00"),Open:4628,   High:4644.25,Low:4583.5, Close:4584,   UpVolume:217997,DownVolume:227786,UpTicks:145243,DownTicks:153114,OpenInterest:0},
	{Time:p("2021-11-30T17:00:00+00:00"),Open:4583.75,High:4602.5, Low:4565.75,Close:4576.75,UpVolume:205186,DownVolume:208658,UpTicks:138530,DownTicks:138482,OpenInterest:0},
	{Time:p("2021-11-30T18:00:00+00:00"),Open:4576.5, High:4584.25,Low:4562,   Close:4575,   UpVolume:123305,DownVolume:121996,UpTicks:81035, DownTicks:81774, OpenInterest:0},
	{Time:p("2021-11-30T19:00:00+00:00"),Open:4574.75,High:4593,   Low:4567.5, Close:4592.5, UpVolume:90243, DownVolume:84177, UpTicks:55940, DownTicks:56369, OpenInterest:0},
	{Time:p("2021-11-30T20:00:00+00:00"),Open:4592.5, High:4597,   Low:4573,   Close:4577.25,UpVolume:88109, DownVolume:90754, UpTicks:56706, DownTicks:58203, OpenInterest:0},
	{Time:p("2021-11-30T21:00:00+00:00"),Open:4577.25,High:4580.5, Low:4557,   Close:4562.75,UpVolume:217907,DownVolume:237060,UpTicks:118521,DownTicks:126568,OpenInterest:0},
	{Time:p("2021-11-30T22:00:00+00:00"),Open:4562.75,High:4585.75,Low:4562.5, Close:4585.5, UpVolume:63647, DownVolume:59388, UpTicks:30973, DownTicks:28647, OpenInterest:0},
	{Time:p("2021-12-01T00:00:00+00:00"),Open:4587.5, High:4595,   Low:4580,   Close:4584.5, UpVolume:9703,  DownVolume:9437,  UpTicks:6937,  DownTicks:6875,  OpenInterest:0},
}

var daily = DataPoint{
	Time:p("2021-12-01T00:00:00+00:00"),Open:4663,High:4667.5,Low:4557,Close:4584.5,UpVolume:1357476,DownVolume:1362058,UpTicks:876235,DownTicks:884463,OpenInterest:0,
}

//=============================================================================

func p(d string) time.Time {
	t, err := time.ParseInLocation(time.RFC3339, d, time.UTC)
	if err != nil { panic(err) }

	return t.In(time.UTC)
}

//=============================================================================

func TestDailyAggregator(t *testing.T) {
	da60m := NewDataAggregator(nil)

	for _, dp := range hourly {
		da60m.Add(&dp)
	}

	da1day := NewDataAggregator(TimeSlotFunction1440m)
	da60m.Aggregate(da1day)

	if len(da1day.dataPoints) != 1 {
		t.Errorf("Too many data points in a daily aggregate. Expected %v but got %v", 1, len(da1day.dataPoints))
	}

	dp := da1day.dataPoints[0]

	if *dp != daily {
		t.Errorf("Data point %v does not match expected value %v", dp, daily)
	}
}

//=============================================================================
