package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func dateFormat(t time.Time) string {
	return t.Format(time.Kitchen)
}

func webServer(imageDirectory, templateFile string, appList *ArgoCDAppList, releaseMode bool) {

	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"dateFormat": dateFormat,
	})

	router.Static(imageDirectory, "./img")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})

	router.LoadHTMLFiles(templateFile)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, templateFile, gin.H{
			"appList": &appList,
		})
	})
	router.Run(":80")
}
