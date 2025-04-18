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
	"bufio"
	"context"
	"github.com/bit-fever/core"
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"os"
	"time"
)

//=============================================================================

var pool *pgxpool.Pool
var staging string

type Formatter func(dp *DataPoint) any

//=============================================================================

func InitDatastore(cfg *app.Datastore) {

	slog.Info("Starting datastore...")
	url := "postgres://"+ cfg.Username + ":" + cfg.Password + "@" + cfg.Address + "/" + cfg.Name

	p, err := pgxpool.New(context.Background(), url)
	if err != nil {
		core.ExitWithMessage("Failed to connect to the datastore: "+ err.Error())
	}

	pool    = p
	staging = cfg.Staging
}

//=============================================================================
//===
//=== Datafile management
//===
//=============================================================================

func OpenDatafile(filename string) (*os.File, error){
	return os.Open(staging + string(os.PathSeparator) + filename)
}

//=============================================================================

func SaveDatafile(part io.Reader) (string, int64, error) {
	var bytes int64
	filename := uuid.NewString()
	slog.Info("Starting datafile upload", "filename", filename)

	file, err := os.Create(staging + string(os.PathSeparator) + filename)
	if err == nil {
		w := bufio.NewWriter(file)
		bytes, err = io.Copy(w, part)

		if err == nil {
			err = w.Flush()
			if err == nil {
				err = file.Close()
				if err == nil {
					slog.Info("Ending datafile upload", "filename", filename, "bytes", bytes)
					return filename, bytes, nil
				}
			}
		}

		_= file.Close()
		_= os.Remove(filename)
	}

	slog.Info("Error during datafile upload", "filename", filename, "error", err.Error())
	return "", 0, err
}

//=============================================================================

func DeleteDataFile(filename string) error {
	return os.Remove(staging + string(os.PathSeparator) + filename)
}

//=============================================================================
//===
//=== Datapoints get/set
//===
//=============================================================================

func GetDataPoints(from time.Time, to time.Time, config *DataConfig, loc *time.Location, da *DataAggregator) error {
	query := buildGetQuery(config)

	rows, err := pool.Query(context.Background(), query, config.Symbol, config.Selector, from, to)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var dp DataPoint
		err = rows.Scan(&dp.Time, &dp.Open, &dp.High, &dp.Low, &dp.Close, &dp.UpVolume, &dp.DownVolume)

		if err != nil {
			return err
		}

		dp.Time = dp.Time.In(loc)
		da.Add(&dp)
	}

	da.Flush()

	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

//=============================================================================

func SetDataPoints(points []*DataPoint, config *DataConfig) error {
	if len(points) == 0 {
		return nil
	}

	query := buildAddQuery(config)
	batch := &pgx.Batch{}

	for i := range points {
		dp := points[i]
		batch.Queue(query, dp.Time, config.Symbol, config.Selector, dp.Open, dp.High, dp.Low, dp.Close, dp.UpVolume, dp.DownVolume)
	}

	br := pool.SendBatch(context.Background(), batch)
	_, err := br.Exec()
	_ = br.Close()

	return err
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func buildGetQuery(config *DataConfig) string {
	table := "system_data_"
	field := "system_code"

	if config.UserTable {
		table = "user_data_"
		field = "product_id"
	}

	table = table + config.Timeframe

	query := 	"SELECT time, open, high, low, close, up_volume, down_volume FROM "+ table +" "+
				"WHERE symbol = $1 AND "+ field +" = $2 AND time >= $3 AND time <= $4 "+
				"ORDER BY time"

	return query
}

//=============================================================================

func buildAddQuery(config *DataConfig) string {
	table := "system_data_"
	field := "system_code"

	if config.UserTable {
		table = "user_data_"
		field = "product_id"
	}

	table = table + config.Timeframe

	query := 	"INSERT INTO "+ table +"(time, symbol, "+ field +", open, high, low, close, up_volume, down_volume) " +
				"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) " +
				"ON CONFLICT(time, symbol, "+ field +") DO UPDATE SET "+
				"open=excluded.open,"+
				"high=excluded.high,"+
				"low=excluded.low,"+
				"close=excluded.close,"+
				"up_volume=excluded.up_volume,"+
				"down_volume=excluded.down_volume"

	return query
}

//=============================================================================
