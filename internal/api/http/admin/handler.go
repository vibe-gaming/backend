package admin

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

//go:embed *.html
var adminFiles embed.FS

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) AdminPage(c *gin.Context) {
	h.serveHTML(c, "admin.html", "admin page")
}

func (h *Handler) CreateBenefitPage(c *gin.Context) {
	h.serveHTML(c, "create-benefit.html", "create benefit page")
}

func (h *Handler) BenefitsListPage(c *gin.Context) {
	h.serveHTML(c, "benefits-list.html", "benefits list page")
}

func (h *Handler) BenefitDetailPage(c *gin.Context) {
	h.serveHTML(c, "benefit-detail.html", "benefit detail page")
}

func (h *Handler) OrganizationsListPage(c *gin.Context) {
	h.serveHTML(c, "organizations-list.html", "organizations list page")
}

func (h *Handler) OrganizationDetailPage(c *gin.Context) {
	h.serveHTML(c, "organization-detail.html", "organization detail page")
}

func (h *Handler) CreateOrganizationPage(c *gin.Context) {
	h.serveHTML(c, "create-organization.html", "create organization page")
}

func (h *Handler) serveHTML(c *gin.Context, filename, pageName string) {
	htmlContent, err := adminFiles.ReadFile(filename)
	if err != nil {
		logger.Error("failed to read "+pageName, zap.Error(err), zap.String("file", filename))
		c.String(http.StatusInternalServerError, "Failed to read "+pageName)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
}
