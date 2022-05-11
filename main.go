package main

import (
	"github.com/condigolabs/content-marketing/config"
	"github.com/condigolabs/content-marketing/controller"
	"github.com/condigolabs/content-marketing/startup"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var router *gin.Engine

func main() {
	config.LogInit()
	startup.GetIntent()
	startup.GetGenerator()
	defer func() { startup.Close() }()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://cdglb.com" ,"https://cdglb.com", "https://condigolabs.com", "http://condigolabs.com"},
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.GET("/health", func(c *gin.Context) {
		ret := gin.H{"app": "content-marketing-gin"}
		for k, v := range c.Request.Header {
			ret[k] = v
		}
		c.JSON(http.StatusOK, ret)
	})
	router.GET("/templates", func(c *gin.Context) {
		c.HTML(http.StatusOK, "default.tmpl.html", gin.H{"Message": "hello world"})
	})
	controller.InitRouter(router)
	_ = router.Run(":8080")
}
