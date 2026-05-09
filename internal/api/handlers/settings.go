package handlers

import (
	"net/http"

	"codeswitch/internal/config"
	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles system settings and app settings HTTP endpoints.
type SettingsHandler struct {
	settings    *services.SettingsService
	appSettings *services.AppSettingsService
	cfg         *config.AppConfig
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(settings *services.SettingsService, appSettings *services.AppSettingsService, cfg *config.AppConfig) *SettingsHandler {
	return &SettingsHandler{
		settings:    settings,
		appSettings: appSettings,
		cfg:         cfg,
	}
}

// settingsResponse represents the JSON response for GET /settings/.
type settingsResponse struct {
	ProxyPort       int    `json:"proxyPort"`
	ProxyListenAddr string `json:"proxyListenAddr"`
	LogLevel        string `json:"logLevel"`
	DatabasePath    string `json:"databasePath"`
	Port            int    `json:"port"`
}

// updateSettingsRequest represents the JSON body for PUT /settings/.
type updateSettingsRequest struct {
	ProxyPort       *int    `json:"proxyPort"`
	ProxyListenAddr *string `json:"proxyListenAddr"`
	LogLevel        *string `json:"logLevel"`
	DatabasePath    *string `json:"databasePath"`
	Port            *int    `json:"port"`
}

// GetSettings handles GET /settings/.
// Returns all system-level settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, settingsResponse{
		ProxyPort:       h.cfg.ProxyPort,
		ProxyListenAddr: h.cfg.ProxyListenAddr,
		LogLevel:        h.cfg.LogLevel,
		DatabasePath:    h.cfg.DatabasePath,
		Port:            h.cfg.Port,
	})
}

// UpdateSettings handles PUT /settings/.
// Validates port ranges (1-65535) and updates system settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate port ranges
	if req.ProxyPort != nil {
		if *req.ProxyPort < 1 || *req.ProxyPort > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "proxyPort must be between 1 and 65535"})
			return
		}
		h.cfg.ProxyPort = *req.ProxyPort
	}

	if req.Port != nil {
		if *req.Port < 1 || *req.Port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "port must be between 1 and 65535"})
			return
		}
		h.cfg.Port = *req.Port
	}

	if req.ProxyListenAddr != nil {
		h.cfg.ProxyListenAddr = *req.ProxyListenAddr
	}

	if req.LogLevel != nil {
		level := *req.LogLevel
		if level != "debug" && level != "info" && level != "warn" && level != "error" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "logLevel must be one of: debug, info, warn, error"})
			return
		}
		h.cfg.LogLevel = level
	}

	if req.DatabasePath != nil {
		if *req.DatabasePath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "databasePath cannot be empty"})
			return
		}
		h.cfg.DatabasePath = *req.DatabasePath
	}

	c.JSON(http.StatusOK, settingsResponse{
		ProxyPort:       h.cfg.ProxyPort,
		ProxyListenAddr: h.cfg.ProxyListenAddr,
		LogLevel:        h.cfg.LogLevel,
		DatabasePath:    h.cfg.DatabasePath,
		Port:            h.cfg.Port,
	})
}

// GetAppSettings handles GET /settings/app.
// Returns the application-level settings (UI preferences, budget, etc.).
func (h *SettingsHandler) GetAppSettings(c *gin.Context) {
	settings, err := h.appSettings.GetAppSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateAppSettings handles PUT /settings/app.
// Updates application-level settings.
func (h *SettingsHandler) UpdateAppSettings(c *gin.Context) {
	var settings services.AppSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	saved, err := h.appSettings.SaveAppSettings(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, saved)
}
