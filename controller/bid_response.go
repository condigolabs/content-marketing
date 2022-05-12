package controller

import (
	"github.com/condigolabs/content-marketing/services/intent"
	"github.com/condigolabs/content-marketing/startup"
	"github.com/gin-gonic/gin"
	"net/http"
)

func LoadBidRequest(c *gin.Context) {

	service := startup.GetIntent()

	r, err := service.LoadBidRequest(intent.Param{
		Country:   "USA",
		LastHours: 0,
		Locale:    "en-US",
		Value:     "",
		Tag:       "",
		RequestId: "46db40be-ae5f-4de0-9d33-16d877be8005",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, r)
}
