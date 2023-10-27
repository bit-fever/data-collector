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

package db

import (
	"github.com/bit-fever/data-collector/pkg/model/config"
	"gorm.io/driver/mysql"
	"time"

	"gorm.io/gorm"
	"log"
)

//=============================================================================

var dbms *gorm.DB

//=============================================================================

func InitDatabase(cfg *config.Config) {

	log.Println("Starting database...")
	url := cfg.Database.Username + ":" + cfg.Database.Password + "@tcp(" + cfg.Database.Address + ")/" + cfg.Database.Name + "?charset=utf8mb4&parseTime=True"

	dialector := mysql.New(mysql.Config{
		DSN:                       url,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  false,
		DontSupportRenameIndex:    false,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	sqlDB, err := db.DB()
	sqlDB.SetConnMaxLifetime(time.Minute * 3)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)

	dbms = db
}

//=============================================================================

func RunInTransaction(f func(tx *gorm.DB) error) error {
	return dbms.Transaction(f)
}

//=============================================================================
