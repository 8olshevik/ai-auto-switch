package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"codeswitch/internal/api/handlers"
	"codeswitch/internal/api/middleware"
	"codeswitch/internal/config"
	"codeswitch/internal/ws"
	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// RouterServices holds all service instances needed by the router.
type RouterServices struct {
	ProviderSvc    *services.ProviderService
	ProviderRelay  *services.ProviderRelayService
	LogSvc         *services.LogService
	SettingsSvc    *services.SettingsService
	AppSettingsSvc *services.AppSettingsService
	MCPSvc         *services.MCPService
	CLIConfigSvc   *services.CliConfigService
	PromptSvc      *services.PromptService
	SkillSvc       *services.SkillService
	HealthCheckSvc *services.HealthCheckService
	SpeedTestSvc   *services.SpeedTestService
	ImportSvc      *services.ImportService
	UpdateSvc      *services.UpdateService
	BlacklistSvc   *services.BlacklistService
	Hub            *ws.Hub
}

// SetupRouter creates and configures the Gin engine with all middleware and route groups.
// Deprecated: Use SetupRouterWithServices for full service injection.
func SetupRouter(cfg *config.AppConfig, providerSvc *services.ProviderService, providerRelay *services.ProviderRelayService, logSvc *services.LogService, settingsSvc *services.SettingsService, appSettingsSvc *services.AppSettingsService, mcpSvc *services.MCPService) *gin.Engine {
	return SetupRouterWithServices(cfg, &RouterServices{
		ProviderSvc:    providerSvc,
		ProviderRelay:  providerRelay,
		LogSvc:         logSvc,
		SettingsSvc:    settingsSvc,
		AppSettingsSvc: appSettingsSvc,
		MCPSvc:         mcpSvc,
	})
}

// SetupRouterWithServices creates and configures the Gin engine with all services.
func SetupRouterWithServices(cfg *config.AppConfig, svc *RouterServices) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	// Public health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1 route group
	v1 := router.Group("/api/v1")

	// Auth handler
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)

	// Public routes (no auth required)
	authPublic := v1.Group("/auth")
	{
		authPublic.POST("/login", authHandler.Login)
	}

	// Public update routes (no auth required)
	updateHandler := handlers.NewUpdateHandler(svc.UpdateSvc)
	v1.GET("/update/version", updateHandler.GetVersion)

	// Authenticated routes (JWT required)
	authorized := v1.Group("/")
	authorized.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Auth (authenticated endpoints)
		authProtected := authorized.Group("/auth")
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.GET("/me", authHandler.GetMe)
		}

		// Providers management
		if svc.ProviderSvc != nil {
			providerHandler := handlers.NewProviderHandler(svc.ProviderSvc)
			providers := authorized.Group("/providers")
			{
				providers.GET("/:kind", providerHandler.GetProviders)
				providers.POST("/:kind", providerHandler.SaveProviders)
				providers.POST("/:kind/duplicate/:id", providerHandler.DuplicateProvider)
				providers.PUT("/:kind/:id/rename", providerHandler.RenameProvider)
				providers.POST("/:kind/reorder", providerHandler.ReorderProviders)
			}
		}

		// Proxy management
		if svc.ProviderRelay != nil {
			proxyHandler := handlers.NewProxyHandler(svc.ProviderRelay)
			proxy := authorized.Group("/proxy")
			{
				proxy.GET("/status", proxyHandler.GetStatus)
				proxy.POST("/start", proxyHandler.Start)
				proxy.POST("/stop", proxyHandler.Stop)
				proxy.GET("/last-used", proxyHandler.GetLastUsed)
			}
		}

		// Settings
		if svc.SettingsSvc != nil && svc.AppSettingsSvc != nil {
			settingsHandler := handlers.NewSettingsHandler(svc.SettingsSvc, svc.AppSettingsSvc, cfg)
			settings := authorized.Group("/settings")
			{
				settings.GET("/", settingsHandler.GetSettings)
				settings.PUT("/", settingsHandler.UpdateSettings)
				settings.GET("/app", settingsHandler.GetAppSettings)
				settings.PUT("/app", settingsHandler.UpdateAppSettings)
			}
		}

		// Blacklist
		if svc.BlacklistSvc != nil {
			blacklistHandler := handlers.NewBlacklistHandler(svc.BlacklistSvc)
			blacklist := authorized.Group("/blacklist")
			{
				blacklist.GET("/", blacklistHandler.GetStatus)
				blacklist.POST("/recover", blacklistHandler.ManualRecover)
				blacklist.GET("/settings", blacklistHandler.GetSettings)
			}
		}

		// Logs
		if svc.LogSvc != nil {
			logsHandler := handlers.NewLogsHandler(svc.LogSvc)
			logs := authorized.Group("/logs")
			{
				logs.GET("/", logsHandler.ListLogs)
				logs.GET("/stats", logsHandler.GetStats)
				logs.GET("/heatmap", logsHandler.GetHeatmap)
				logs.DELETE("/", logsHandler.ClearLogs)
			}
		}

		// MCP server management
		if svc.MCPSvc != nil {
			mcpHandler := handlers.NewMcpHandler(svc.MCPSvc)
			mcp := authorized.Group("/mcp")
			{
				mcp.GET("/", mcpHandler.List)
				mcp.POST("/", mcpHandler.Create)
				mcp.PUT("/:id", mcpHandler.Update)
				mcp.DELETE("/:id", mcpHandler.Delete)
			}
		}

		// CLI config
		if svc.CLIConfigSvc != nil {
			cliConfigHandler := handlers.NewCLIConfigHandler(svc.CLIConfigSvc)
			cliConfig := authorized.Group("/cli-config")
			{
				cliConfig.GET("/:platform", cliConfigHandler.GetConfig)
				cliConfig.PUT("/:platform", cliConfigHandler.SaveConfig)
				cliConfig.GET("/:platform/snapshots", cliConfigHandler.GetConfigSnapshots)
				cliConfig.PUT("/:platform/file", cliConfigHandler.SaveConfigFileContent)
				cliConfig.GET("/:platform/template", cliConfigHandler.GetTemplate)
				cliConfig.PUT("/:platform/template", cliConfigHandler.SetTemplate)
				cliConfig.POST("/:platform/restore", cliConfigHandler.RestoreDefault)
			}
		}

		// Prompts
		if svc.PromptSvc != nil {
			promptsHandler := handlers.NewPromptsHandler(svc.PromptSvc)
			prompts := authorized.Group("/prompts")
			{
				prompts.GET("/:platform", promptsHandler.GetPrompts)
				prompts.POST("/:platform", promptsHandler.CreatePrompt)
				prompts.PUT("/:platform/:id", promptsHandler.UpdatePrompt)
				prompts.DELETE("/:platform/:id", promptsHandler.DeletePrompt)
				prompts.POST("/:platform/:id/enable", promptsHandler.EnablePrompt)
				prompts.GET("/:platform/file", promptsHandler.GetCurrentFileContent)
			}
		}

		// Skills
		if svc.SkillSvc != nil {
			skillsHandler := handlers.NewSkillsHandler(svc.SkillSvc)
			skills := authorized.Group("/skills")
			{
				skills.GET("/", skillsHandler.ListSkills)
				skills.GET("/:platform", skillsHandler.ListSkillsForPlatform)
				skills.POST("/install", skillsHandler.InstallSkill)
				skills.DELETE("/", skillsHandler.UninstallSkill)
				skills.POST("/toggle", skillsHandler.ToggleSkill)
				skills.GET("/content", skillsHandler.GetSkillContent)
				skills.PUT("/content", skillsHandler.SaveSkillContent)
				skills.GET("/repos", skillsHandler.ListRepos)
			}
		}

		// Health check (authenticated)
		if svc.HealthCheckSvc != nil {
			healthHandler := handlers.NewHealthHandler(svc.HealthCheckSvc)
			health := authorized.Group("/health")
			{
				health.GET("/", healthHandler.GetLatestResults)
				health.POST("/check", healthHandler.RunCheck)
				health.GET("/history", healthHandler.GetHistory)
			}
		}

		// Speed test
		if svc.SpeedTestSvc != nil {
			speedTestHandler := handlers.NewSpeedTestHandler(svc.SpeedTestSvc)
			speedtest := authorized.Group("/speedtest")
			{
				speedtest.POST("/run", speedTestHandler.Run)
			}
		}

		// Import/Export
		if svc.ImportSvc != nil {
			importHandler := handlers.NewImportHandler(svc.ImportSvc)
			importExport := authorized.Group("/import")
			{
				importExport.POST("/config", importHandler.ImportConfig)
				importExport.GET("/export", importHandler.ExportConfig)
				importExport.GET("/status", importHandler.GetStatus)
				importExport.POST("/mcp/parse", importHandler.ParseMCPJSON)
				importExport.POST("/mcp", importHandler.ImportMCPServers)
			}
		}

		// Update
		if svc.UpdateSvc != nil {
			updateHandler := handlers.NewUpdateHandler(svc.UpdateSvc)
			update := authorized.Group("/update")
			{
				update.GET("/check", updateHandler.CheckUpdate)
				update.GET("/state", updateHandler.GetState)
				update.POST("/download", updateHandler.DownloadUpdate)
				update.POST("/install", updateHandler.InstallUpdate)
				update.POST("/dismiss", updateHandler.DismissUpdate)
				update.POST("/cancel", updateHandler.CancelDownload)
			}
		}

		// API Gateway
		if true { // GatewayHandler is always available
			gatewayHandler := handlers.NewGatewayHandler()
			gateway := authorized.Group("/gateway")
			{
				gateway.POST("/keys", gatewayHandler.CreateKey)
				gateway.GET("/keys", gatewayHandler.ListKeys)
				gateway.DELETE("/keys/:id", gatewayHandler.DeleteKey)
				gateway.POST("/keys/:id/toggle", gatewayHandler.ToggleKey)
				gateway.GET("/stats", gatewayHandler.GetStats)
				gateway.PUT("/rate-limit", gatewayHandler.UpdateRateLimit)
			}
		}

		// AI Assistant
		if svc.ProviderRelay != nil && svc.Hub != nil {
			assistantHandler := handlers.NewAssistantHandler(
				svc.ProviderRelay,
				svc.ProviderSvc,
				svc.SettingsSvc,
				svc.AppSettingsSvc,
				svc.LogSvc,
				svc.Hub,
				svc.ProviderRelay.Addr(),
			)
			assistant := authorized.Group("/assistant")
			{
				assistant.POST("/chat", assistantHandler.Chat)
				assistant.GET("/history", assistantHandler.History)
				assistant.DELETE("/history", assistantHandler.ClearHistory)
				assistant.POST("/execute", assistantHandler.Execute)
			}
		}

		// WebSocket — uses query param token auth (not header-based),
		// so it's registered outside the standard auth middleware group.
		// JWT validation is handled inside the handler itself.
	}

	// WebSocket endpoint with custom auth (token via query parameter)
	if svc.Hub != nil {
		wsHandler := handlers.NewWSHandler(svc.Hub, cfg.JWTSecret)
		v1.GET("/ws/events", wsHandler.HandleConnect)
	}

	// Static file serving for frontend dist directory
	distPath := filepath.Join("frontend", "dist")
	if _, err := os.Stat(distPath); err == nil {
		router.Static("/assets", filepath.Join(distPath, "assets"))
		router.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))

		// Catch-all route: serve index.html for Vue Router history mode
		// Any path that doesn't match /api/* or static assets serves the SPA
		router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(distPath, "index.html"))
		})
	} else {
		// No dist directory available — return 404 for non-API routes
		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})
	}

	return router
}
