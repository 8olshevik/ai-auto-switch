package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// SkillsHandler handles skill management endpoints.
type SkillsHandler struct {
	svc *services.SkillService
}

// NewSkillsHandler creates a new SkillsHandler.
func NewSkillsHandler(svc *services.SkillService) *SkillsHandler {
	return &SkillsHandler{svc: svc}
}

// ListSkills handles GET /skills/.
func (h *SkillsHandler) ListSkills(c *gin.Context) {
	skills, err := h.svc.ListSkills()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skills)
}

// ListSkillsForPlatform handles GET /skills/:platform.
func (h *SkillsHandler) ListSkillsForPlatform(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	skills, err := h.svc.ListSkillsForPlatform(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skills)
}

// installSkillRequest represents the request body for installing a skill.
type installSkillRequest struct {
	Directory string `json:"directory" binding:"required"`
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
	Branch    string `json:"repo_branch"`
	Platform  string `json:"platform"`
	Location  string `json:"location"`
}

// InstallSkill handles POST /skills/install.
func (h *SkillsHandler) InstallSkill(c *gin.Context) {
	var req installSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "directory is required"})
		return
	}

	// The service expects an unexported installRequest type, so we call InstallSkill
	// via the exported method which accepts the same struct shape.
	// We need to use a workaround since installRequest is unexported.
	if err := h.svc.InstallSkill(services.InstallRequest{
		Directory: req.Directory,
		RepoOwner: req.RepoOwner,
		RepoName:  req.RepoName,
		Branch:    req.Branch,
		Platform:  req.Platform,
		Location:  req.Location,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "skill installed"})
}

// uninstallSkillRequest represents the request body for uninstalling a skill.
type uninstallSkillRequest struct {
	Directory string `json:"directory" binding:"required"`
	Platform  string `json:"platform"`
	Location  string `json:"location"`
}

// UninstallSkill handles DELETE /skills/.
func (h *SkillsHandler) UninstallSkill(c *gin.Context) {
	var req uninstallSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "directory is required"})
		return
	}

	var err error
	if req.Platform != "" && req.Location != "" {
		err = h.svc.UninstallSkillEx(req.Directory, req.Platform, req.Location)
	} else {
		err = h.svc.UninstallSkill(req.Directory)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "skill uninstalled"})
}

// toggleSkillRequest represents the request body for toggling a skill.
type toggleSkillRequest struct {
	Directory string `json:"directory" binding:"required"`
	Platform  string `json:"platform" binding:"required"`
	Location  string `json:"location" binding:"required"`
	Enabled   bool   `json:"enabled"`
}

// ToggleSkill handles POST /skills/toggle.
func (h *SkillsHandler) ToggleSkill(c *gin.Context) {
	var req toggleSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "directory, platform, and location are required"})
		return
	}

	if err := h.svc.ToggleSkill(req.Directory, req.Platform, req.Location, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "skill toggled"})
}

// getSkillContentRequest represents query params for getting skill content.
type getSkillContentRequest struct {
	Directory string `form:"directory" binding:"required"`
	Platform  string `form:"platform" binding:"required"`
	Location  string `form:"location" binding:"required"`
}

// GetSkillContent handles GET /skills/content.
func (h *SkillsHandler) GetSkillContent(c *gin.Context) {
	var req getSkillContentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "directory, platform, and location are required"})
		return
	}

	content, err := h.svc.GetSkillContent(req.Directory, req.Platform, req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// saveSkillContentRequest represents the request body for saving skill content.
type saveSkillContentRequest struct {
	Directory string `json:"directory" binding:"required"`
	Platform  string `json:"platform" binding:"required"`
	Location  string `json:"location" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// SaveSkillContent handles PUT /skills/content.
func (h *SkillsHandler) SaveSkillContent(c *gin.Context) {
	var req saveSkillContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "directory, platform, location, and content are required"})
		return
	}

	if err := h.svc.SaveSkillContent(req.Directory, req.Platform, req.Location, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "skill content saved"})
}

// ListRepos handles GET /skills/repos.
func (h *SkillsHandler) ListRepos(c *gin.Context) {
	repos, err := h.svc.ListRepos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, repos)
}
