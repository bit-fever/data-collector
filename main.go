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

package main

import (
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/boot"
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/bit-fever/data-collector/pkg/core/jobmanager"
	"github.com/bit-fever/data-collector/pkg/core/messaging"
	"github.com/bit-fever/data-collector/pkg/core/process"
	"github.com/bit-fever/data-collector/pkg/db"
	"github.com/bit-fever/data-collector/pkg/ds"
	"github.com/bit-fever/data-collector/pkg/platform"
	"github.com/bit-fever/data-collector/pkg/service"
	"log/slog"
)

//=============================================================================

const component = "data-collector"

//=============================================================================

func main() {
	cfg := &app.Config{}
	boot.ReadConfig(component, cfg)
	logger := boot.InitLogger(component, &cfg.Application)
	engine := boot.InitEngine(logger,    &cfg.Application)
	initClients()
	db.InitDatabase(&cfg.Database)
	ds.InitDatastore(&cfg.Datastore)
	platform.Init(&cfg.Platform)
	auth.InitAuthentication(&cfg.Authentication)
	msg.InitMessaging(&cfg.Messaging)
	service.Init(engine, cfg, logger)
	messaging.InitMessageListener()
	process.Init(cfg)
	jobmanager.Init(cfg)
	boot.RunHttpServer(engine, &cfg.Application)
}

//=============================================================================

func initClients() {
	slog.Info("Initializing clients...")
	req.AddClient("bf", "ca.crt", "server.crt", "server.key")
}

//=============================================================================
