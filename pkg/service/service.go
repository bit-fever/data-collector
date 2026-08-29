//=============================================================================
//===
//=== Copyright (C) 2024-present Andrea Carboni
//===
//=== This source code is licensed under the Elastic License 2.0 (ELv2) available at:
//=== https://github.com/algotiqa/docs/blob/main/LICENSE.md
//=== By using this file, you agree to the terms and conditions of that license.
//=============================================================================


package service

import (
	"log/slog"

	"github.com/algotiqa/core/auth"
	"github.com/algotiqa/core/auth/roles"
	"github.com/algotiqa/core/req"
	"github.com/algotiqa/data-collector/pkg/app"
	"github.com/gin-gonic/gin"
)

//=============================================================================

func Init(router *gin.Engine, cfg *app.Config, logger *slog.Logger) {

	ctrl := auth.NewOidcController(cfg.Authentication.Authority, req.GetDefaultClient(), logger, cfg)

	router.GET  ("/api/collector/v1/config/parsers",                 ctrl.Secure(getParsers, roles.Admin_User_Service))

	router.GET   ("/api/collector/v1/data-instruments",              ctrl.Secure(getDataInstruments,            roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/data-instruments/:id",          ctrl.Secure(getDataInstrumentById,         roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/data-instruments/:id/data",     ctrl.Secure(getDataInstrumentData,         roles.Admin_User_Service))
	router.POST  ("/api/collector/v1/data-instruments/:id/reload",   ctrl.Secure(reloadDataInstrumentData,      roles.Admin_User_Service))

	router.GET   ("/api/collector/v1/data-products/:id/instruments", ctrl.Secure(getDataInstrumentsByProductId, roles.Admin_User_Service))
	router.POST  ("/api/collector/v1/data-products/:id/instruments", ctrl.Secure(uploadDataInstrumentData,      roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/data-products/:id/analysis",    ctrl.Secure(analyzeDataProduct,            roles.Admin_User_Service))

	router.GET   ("/api/collector/v1/bias-analyses",                  ctrl.Secure(getBiasAnalyses,     roles.Admin_User_Service))
	router.POST  ("/api/collector/v1/bias-analyses",                  ctrl.Secure(addBiasAnalysis,     roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/bias-analyses/:id",              ctrl.Secure(getBiasAnalysisById, roles.Admin_User_Service))
	router.PUT   ("/api/collector/v1/bias-analyses/:id",              ctrl.Secure(updateBiasAnalysis,  roles.Admin_User_Service))
	router.DELETE("/api/collector/v1/bias-analyses/:id",              ctrl.Secure(deleteBiasAnalysis,  roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/bias-analyses/:id/summary",      ctrl.Secure(getBiasSummary,      roles.Admin_User_Service))
	router.GET   ("/api/collector/v1/bias-analyses/:id/backtest",     ctrl.Secure(runBacktest,         roles.Admin_User_Service))

	router.GET   ("/api/collector/v1/bias-analyses/:id/configs",      ctrl.Secure(getBiasConfigsByAnalysisId, roles.Admin_User_Service))
	router.POST  ("/api/collector/v1/bias-analyses/:id/configs",      ctrl.Secure(addBiasConfig,              roles.Admin_User_Service))
	router.PUT   ("/api/collector/v1/bias-analyses/:id/configs/:id2", ctrl.Secure(updateBiasConfig,           roles.Admin_User_Service))
	router.DELETE("/api/collector/v1/bias-analyses/:id/configs/:id2", ctrl.Secure(deleteBiasConfig,           roles.Admin_User_Service))
}

//=============================================================================
