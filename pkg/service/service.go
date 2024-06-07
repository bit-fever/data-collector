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

package service

import (
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/auth/roles"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/data-collector/pkg/app"
	"github.com/gin-gonic/gin"
	"log/slog"
)

//=============================================================================

func Init(router *gin.Engine, cfg *app.Config, logger *slog.Logger) {

	ctrl := auth.NewOidcController(cfg.Authentication.Authority, req.GetClient("bf"), logger, cfg)

	router.GET ("/api/collector/v1/data",                         ctrl.Secure(getData,      roles.Admin_User_Service))
	router.POST("/api/collector/v1/data/:id/filters",             ctrl.Secure(ingestData,   roles.Admin_User_Service))

	router.GET ("/api/collector/v1/product-data/:id/instruments", ctrl.Secure(getInstrumentDataByProductId, roles.Admin_User_Service))
	router.POST("/api/collector/v1/product-data/:id/instruments", ctrl.Secure(uploadInstrumentData,         roles.Admin_User_Service))
}

//=============================================================================
