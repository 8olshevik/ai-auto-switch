package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeswitch/internal/api"
	"codeswitch/internal/config"
	"codeswitch/internal/ws"
	"codeswitch/services"
)

func main() {
	// 1. Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ 配置验证失败: %v", err)
	}
	log.Println("✅ 配置已加载")

	// 2. Initialize database
	if err := services.InitDatabase(); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	log.Println("✅ 数据库已初始化")

	// 2.1 Create default admin user on first startup
	if err := services.CreateDefaultAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("❌ 创建默认管理员失败: %v", err)
	}

	// 3. Initialize database write queue
	if err := services.InitGlobalDBQueue(); err != nil {
		log.Fatalf("❌ 初始化数据库队列失败: %v", err)
	}
	log.Println("✅ 数据库写入队列已启动")

	// 4. Initialize services
	providerService := services.NewProviderService()
	settingsService := services.NewSettingsService()
	autoStartService := services.NewAutoStartService()
	appSettings := services.NewAppSettingsService(autoStartService)
	notificationService := services.NewNotificationService(appSettings)
	blacklistService := services.NewBlacklistService(settingsService, notificationService)
	geminiService := services.NewGeminiService(fmt.Sprintf("%s:%d", cfg.ProxyListenAddr, cfg.ProxyPort))

	proxyAddr := fmt.Sprintf(":%d", cfg.ProxyPort)
	providerRelay := services.NewProviderRelayService(providerService, geminiService, blacklistService, notificationService, appSettings, proxyAddr)

	healthCheckService := services.NewHealthCheckService(providerService, blacklistService, settingsService)
	if err := healthCheckService.Start(); err != nil {
		log.Fatalf("❌ 初始化健康检查服务失败: %v", err)
	}

	log.Println("✅ 服务实例已创建")

	// 5. Start proxy relay service
	go func() {
		if err := providerRelay.Start(); err != nil {
			log.Printf("⚠️  代理服务启动失败: %v", err)
		}
	}()
	log.Printf("✅ 代理服务启动中 (监听 %s:%d)", cfg.ProxyListenAddr, cfg.ProxyPort)

	// 6. Start background tasks
	blacklistStopChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := blacklistService.AutoRecoverExpired(); err != nil {
					log.Printf("⚠️  自动恢复黑名单失败: %v", err)
				}
			case <-blacklistStopChan:
				log.Println("✅ 黑名单定时器已停止")
				return
			}
		}
	}()

	// Start health check auto-polling after a short delay
	go func() {
		time.Sleep(3 * time.Second)
		settings, err := appSettings.GetAppSettings()
		autoEnabled := true
		if err != nil {
			log.Printf("⚠️  读取应用设置失败（使用默认值）: %v", err)
		} else {
			autoEnabled = settings.AutoConnectivityTest
		}
		if autoEnabled {
			healthCheckService.SetAutoAvailabilityPolling(true)
			log.Println("✅ 自动可用性监控已启动")
		} else {
			log.Println("ℹ️  自动可用性监控已禁用（可在设置中开启）")
		}
	}()

	// 7. Initialize WebSocket Hub for real-time event broadcasting
	hub := ws.NewHub()
	go hub.Run()
	log.Println("✅ WebSocket Hub 已启动")

	// Wire up Hub to services for event broadcasting
	providerRelay.SetWSHub(hub)
	healthCheckService.SetWSHub(hub)

	// 8. Set up Gin HTTP server with full router configuration
	logService := services.NewLogService()
	mcpService := services.NewMCPService()
	cliConfigService := services.NewCliConfigService(fmt.Sprintf("%s:%d", cfg.ProxyListenAddr, cfg.ProxyPort))
	promptService := services.NewPromptService()
	skillService := services.NewSkillService()
	speedTestService := services.NewSpeedTestService()
	importService := services.NewImportService(providerService, mcpService)
	updateService := services.NewUpdateService("")

	router := api.SetupRouterWithServices(cfg, &api.RouterServices{
		ProviderSvc:    providerService,
		ProviderRelay:  providerRelay,
		LogSvc:         logService,
		SettingsSvc:    settingsService,
		AppSettingsSvc: appSettings,
		MCPSvc:         mcpService,
		CLIConfigSvc:   cliConfigService,
		PromptSvc:      promptService,
		SkillSvc:       skillService,
		HealthCheckSvc: healthCheckService,
		SpeedTestSvc:   speedTestService,
		ImportSvc:      importService,
		UpdateSvc:      updateService,
		BlacklistSvc:   blacklistService,
		Hub:            hub,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 9. Start HTTP server in a goroutine
	go func() {
		log.Printf("✅ HTTP 服务已启动 (端口 %d)", cfg.Port)
		log.Printf("   代理端口: %d", cfg.ProxyPort)
		log.Printf("   日志级别: %s", cfg.LogLevel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP 服务启动失败: %v", err)
		}
	}()

	// 10. Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 收到关闭信号，正在优雅关闭...")

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  HTTP 服务关闭超时: %v", err)
	}
	log.Println("✅ HTTP 服务已关闭")

	// Stop background tasks
	close(blacklistStopChan)
	healthCheckService.StopBackgroundPolling()
	log.Println("✅ 健康检查服务已停止")

	// Stop proxy relay
	_ = providerRelay.Stop()
	log.Println("✅ 代理服务已停止")

	// Shutdown database write queue
	if err := services.ShutdownGlobalDBQueue(10 * time.Second); err != nil {
		log.Printf("⚠️  队列关闭超时: %v", err)
	} else {
		stats1 := services.GetGlobalDBQueueStats()
		log.Printf("✅ 单次队列已关闭，统计：成功=%d 失败=%d 平均延迟=%.2fms",
			stats1.SuccessWrites, stats1.FailedWrites, stats1.AvgLatencyMs)
		stats2 := services.GetGlobalDBQueueLogsStats()
		log.Printf("✅ 批量队列已关闭，统计：成功=%d 失败=%d 平均延迟=%.2fms 批次=%d",
			stats2.SuccessWrites, stats2.FailedWrites, stats2.AvgLatencyMs, stats2.BatchCommits)
	}

	log.Println("✅ 所有服务已停止，应用退出")
}
