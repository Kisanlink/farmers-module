package handlers

import (
	"github.com/Kisanlink/farmers-module/services"
)

type CropHandler struct {
	service services.CropCycleServiceInterface
}

func NewCropHandler(service services.CropCycleServiceInterface) *CropHandler {
	return &CropHandler{service: service}
}
