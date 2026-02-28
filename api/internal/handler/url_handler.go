package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"encurtador/internal/model"
	"encurtador/internal/service"
)

// urlServicer is the subset of service.URLService the handler depends on,
// defined here so the handler can be tested with a stub.
type urlServicer interface {
	Create(ctx context.Context, req service.CreateRequest) (*service.CreateResult, error)
	Resolve(ctx context.Context, slug string) (*model.CachedURL, error)
	VerifyPassword(ctx context.Context, slug, password string) (string, error)
	ExpireEarly(ctx context.Context, slug, manageToken string) error
	CheckSlug(ctx context.Context, slug string) (available bool, suggestion string, err error)
}

type URLHandler struct {
	svc         urlServicer
	frontendURL string
}

func NewURLHandler(svc urlServicer, frontendURL string) *URLHandler {
	return &URLHandler{svc: svc, frontendURL: frontendURL}
}

type createRequest struct {
	TargetURL string    `json:"target_url" binding:"required"`
	Slug      string    `json:"slug"`
	TTL       model.TTL `json:"ttl"       binding:"required"`
	Password  string    `json:"password"`
}

type createResponse struct {
	Slug        string    `json:"slug"`
	ShortURL    string    `json:"short_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	Protected   bool      `json:"protected"`
	ManageToken string    `json:"manage_token"`
}

func (h *URLHandler) CreateURL(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateHTTPURL(req.TargetURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_url: " + err.Error()})
		return
	}

	result, err := h.svc.Create(c.Request.Context(), service.CreateRequest{
		TargetURL: req.TargetURL,
		Slug:      req.Slug,
		TTL:       req.TTL,
		Password:  req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "slug is taken and no alternative could be found"})
		case errors.Is(err, service.ErrInvalidSlugFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidTTL):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ttl value"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create URL"})
		}
		return
	}

	c.JSON(http.StatusCreated, createResponse{
		Slug:        result.Slug,
		ShortURL:    result.ShortURL,
		ExpiresAt:   result.ExpiresAt,
		Protected:   result.Protected,
		ManageToken: result.ManageToken,
	})
}

func (h *URLHandler) CheckSlug(c *gin.Context) {
	slug := c.Param("slug")

	available, suggestion, err := h.svc.CheckSlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSlugFormat) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check slug"})
		return
	}

	resp := gin.H{"available": available}
	if suggestion != "" {
		resp["suggestion"] = suggestion
	}
	c.JSON(http.StatusOK, resp)
}

func (h *URLHandler) RedirectOrGate(c *gin.Context) {
	slug := c.Param("slug")

	cached, err := h.svc.Resolve(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if cached == nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/404")
		return
	}

	if cached.Protected {
		c.Redirect(http.StatusFound, h.frontendURL+"/gate/"+slug)
		return
	}

	c.Redirect(http.StatusMovedPermanently, cached.TargetURL)
}

type unlockRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *URLHandler) UnlockURL(c *gin.Context) {
	slug := c.Param("slug")

	var req unlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetURL, err := h.svc.VerifyPassword(c.Request.Context(), slug, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if targetURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found or expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"target_url": targetURL})
}

type expireRequest struct {
	ManageToken string `json:"manage_token" binding:"required"`
}

func (h *URLHandler) ExpireURL(c *gin.Context) {
	slug := c.Param("slug")

	var req expireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ExpireEarly(c.Request.Context(), slug, req.ManageToken); err != nil {
		if errors.Is(err, service.ErrInvalidManageToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid manage token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL has been expired"})
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("must be a valid http or https URL")
	}
	return nil
}
