package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type FPOHandler struct{ svc services.FPOServiceInterface }

func NewFPOHandler(s services.FPOServiceInterface) *FPOHandler { return &FPOHandler{svc: s} }

// POST /fpo
func (h *FPOHandler) Create(c *gin.Context) {
	var f models.FPO
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, resp("invalid input", err.Error()))
		return
	}
	if err := h.svc.Create(&f); err != nil {
		c.JSON(http.StatusInternalServerError, resp("create failed", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, resp("created", f))
}

// GET /fpo/:reg_no
func (h *FPOHandler) Get(c *gin.Context) {
	regNo := c.Param("reg_no")
	f, err := h.svc.Get(regNo)
	if err != nil {
		c.JSON(http.StatusNotFound, resp("not found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp("ok", f))
}

// PUT /fpo/:reg_no
func (h *FPOHandler) Update(c *gin.Context) {
	regNo := c.Param("reg_no")

	var f models.FPO
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, resp("invalid input", err.Error()))
		return
	}

	// Ensure the URL wins over the body so the client cannot update the PK
	f.FpoRegNo = regNo

	if err := h.svc.Update(&f); err != nil {
		c.JSON(http.StatusInternalServerError, resp("update failed", err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp("updated", f))
}

// DELETE /fpo/:reg_no
func (h *FPOHandler) Delete(c *gin.Context) {
	regNo := c.Param("reg_no")
	if err := h.svc.Delete(regNo); err != nil {
		c.JSON(http.StatusInternalServerError, resp("delete failed", err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp("deleted", nil))
}

func (h *FPOHandler) List(c *gin.Context) {
	fpos, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp("fetch failed", err.Error()))
		return
	}
	c.JSON(http.StatusOK, resp("ok", fpos))
}

/* small helper */
func resp(msg string, data interface{}) gin.H {
	return gin.H{
		"status":    http.StatusOK,
		"success":   true,
		"message":   msg,
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}
