//=============================================================================
//===
//=== Copyright (C) 2024-present Andrea Carboni
//===
//=== This source code is licensed under the Elastic License 2.0 (ELv2) available at:
//=== https://github.com/algotiqa/docs/blob/main/LICENSE.md
//=== By using this file, you agree to the terms and conditions of that license.
//=============================================================================


package business

import (
	"errors"
	"strconv"
	"time"

	"github.com/algotiqa/data-collector/pkg/core"
	"github.com/algotiqa/data-collector/pkg/ds"
	"github.com/algotiqa/types"
)

//=============================================================================

const HardLimit = 3000000

//=============================================================================

type QuerySpec struct {
	Id        uint
	From      string
	To        string
	DaysBack  string
	Timezone  string
	Timeframe string
	Reduction string
	Limit     string
	Config    *core.QueryConfig
}

//=============================================================================

type QueryParams struct {
	TargetLoc  *time.Location
	ProductLoc *time.Location
	From       *time.Time
	To         *time.Time
	Reduction  int
	Limit      int
	Timeframe  int
	Aggregator ds.DataAggregator
}

//=============================================================================

func NewQueryParams(spec *QuerySpec) (*QueryParams, error) {
	targLoc, err := getLocation(spec.Timezone, spec.Config)
	if err != nil {
		return nil, errors.New("Bad 'timezone': " + spec.Timezone + " (" + err.Error() + ")")
	}

	prodLoc, err := time.LoadLocation(spec.Config.DataProduct.Timezone)
	if err != nil {
		return nil, errors.New("Bad product timezone: " + spec.Config.DataProduct.Timezone + " (" + err.Error() + ")")
	}

	daysBack, err := parseBackDays(spec.DaysBack)
	if err != nil {
		return nil, errors.New("Bad 'backDays': " + spec.DaysBack + " (" + err.Error() + ")")
	}

	var from, to *time.Time

	if daysBack > 0 {
		now := time.Now()
		back := now.Add(-time.Hour * 24 * time.Duration(daysBack))
		from = &back
		to = &now
	} else {
		from, err = parseTime(spec.From, targLoc)
		if err != nil {
			return nil, errors.New("Bad 'from': " + spec.From + " (" + err.Error() + ")")
		}

		to, err = parseTime(spec.To, targLoc)
		if err != nil {
			return nil, errors.New("Bad 'to': " + spec.From + " (" + err.Error() + ")")
		}
	}

	timeframe, err := parseTimeframe(spec.Timeframe)
	if err != nil {
		return nil, errors.New("Bad 'timeframe': " + spec.Timeframe + " (" + err.Error() + ")")
	}

	da := buildDataAggregator(timeframe, spec.Config.TradingSession)

	red, err := parseReduction(spec.Reduction)
	if err != nil {
		return nil, errors.New("Bad 'reduction': " + spec.Reduction + " (" + err.Error() + ")")
	}

	lim, err := parseLimit(spec.Limit)
	if err != nil {
		return nil, errors.New("Bad 'limit': " + spec.Limit + " (" + err.Error() + ")")
	}

	return &QueryParams{
		From      : from,
		To        : to,
		TargetLoc : targLoc,
		ProductLoc: prodLoc,
		Reduction : red,
		Limit     : lim,
		Timeframe : timeframe,
		Aggregator: da,
	}, nil
}

//=============================================================================

func getLocation(timezone string, config *core.QueryConfig) (*time.Location, error) {
	if timezone == "" || timezone == "exchange" {
		timezone = config.DataProduct.Timezone
	}

	return time.LoadLocation(timezone)
}

//=============================================================================

func parseBackDays(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	days, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	if days < 0 || days > 10000 {
		return 0, errors.New("allowed range is [0..10000]")
	}

	return days, nil
}

//=============================================================================

func parseTimeframe(value string) (int, error) {
	if value == "" {
		return 0, errors.New("value is missing")
	}

	tf, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	if tf < 1 || tf > 1440 {
		return 0, errors.New("allowed range is [1..1440]")
	}

	return tf, nil
}

//=============================================================================

func parseTime(t string, loc *time.Location) (*time.Time, error) {
	if len(t) == 0 {
		return nil, nil
	}

	date, err := time.ParseInLocation(time.DateTime, t, loc)
	if err == nil {
		date = date.UTC()
	}

	return &date, err
}

//=============================================================================

func parseReduction(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	red, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	if red == 0 {
		return red, nil
	}

	if red < 100 || red > 100000 {
		return 0, errors.New("allowed range is [100..100000]")
	}

	return red, nil
}

//=============================================================================

func parseLimit(value string) (int, error) {
	if value == "" {
		return HardLimit, nil
	}

	lim, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	if lim == 0 {
		return HardLimit, nil
	}

	if lim < 1 || lim > HardLimit {
		return 0, errors.New("allowed range is [1.."+ strconv.Itoa(HardLimit) +"]")
	}

	return lim, nil
}

//=============================================================================

func buildDataAggregator(timeframe int, session *types.TradingSession) ds.DataAggregator {
	if timeframe == 1440 {
		return ds.NewDailyAggregator(session)
	}

	granularity := session.Granularity()

	if  (timeframe % 60 == 0) && (granularity == 60) {
		return ds.NewStandardAggregator(session, granularity, timeframe)
	}

	if  (timeframe % 15 == 0) && (granularity >= 15) {
		return ds.NewStandardAggregator(session, 15, timeframe)
	}

	if  (timeframe % 5 == 0) && (granularity >=  5) {
		return ds.NewStandardAggregator(session, 5, timeframe)
	}

	return ds.NewStandardAggregator(session, 1, timeframe)
}

//=============================================================================
