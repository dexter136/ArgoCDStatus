package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dexter136/ArgoCDStatus/v2/src/argoappread"
	"github.com/gin-gonic/gin"
)

func dateFormat(t time.Time) string {
	return t.Format(time.Kitchen)
}

func main() {

	gin.SetMode(gin.ReleaseMode)

	var appList argoappread.ArgoCDAppList

	err := os.Mkdir("./img", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	go argoappread.GetArgoApps(&appList)

	router := gin.Default()

	router.SetFuncMap(template.FuncMap{
		"dateFormat": dateFormat,
	})

	router.Static("/img", "./img")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})

	router.LoadHTMLFiles("./status.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "status.tmpl", gin.H{
			"appList": &appList,
		})
	})
	router.Run(":80")

}
