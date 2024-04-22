package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type ArgoCDAppList struct {
	Apps      []ArgoCDApp `json:"items"`
	TimeStamp time.Time
}

type ArgoCDApp struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Status struct {
		// Resources []ArgoResources `json:"resources"`
		Sync struct {
			Status   string `json:"status"`
			Revision string `json:"revision"`
		} `json:"sync"`
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
	} `json:"status"`
}

/* For possible future use
type ArgoResources struct {
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Health    struct {
		Status string `json:"status,omitempty"`
	} `json:"health,omitempty"`
}
*/

func getArgoApps(AppList *ArgoCDAppList) {
	argocdUrl := os.Getenv("ARGOCD_URL")
	argocdUser := os.Getenv("ARGOCD_USER")
	argocdPass := os.Getenv("ARGOCD_PASS")
	// TODO: Validate inputs

	client := resty.New()

	resp, _ := client.R().
		SetBody(map[string]interface{}{"username": argocdUser, "password": argocdPass}).
		Post(fmt.Sprintf("%s/api/v1/session", argocdUrl))
	// TODO: Error check login

	for {
		// TODO: Allow setting sync interval via ENV
		if time.Now().Sub(AppList.TimeStamp).Seconds() < 120 {
			time.Sleep(time.Second*120 - time.Now().Sub(AppList.TimeStamp))
		}

		resp, _ = client.R().Get(fmt.Sprintf("%s/api/v1/applications", argocdUrl))
		// TODO: Error check repsonse

		json.Unmarshal(resp.Body(), &AppList)
		// TODO: Error check applist

		AppList.TimeStamp = time.Now()
		for _, app := range AppList.Apps {
			img, _ := os.Create(fmt.Sprintf("img/%s.svg", app.Metadata.Name))
			defer img.Close()
			resp, _ := http.Get(fmt.Sprintf("%s/api/badge?name=%s&revision=true", argocdUrl, app.Metadata.Name))
			defer resp.Body.Close()
			io.Copy(img, resp.Body)
		}
	}
}

func dateFormat(t time.Time) string {
	return t.Format(time.Kitchen)
}

func main() {

	gin.SetMode(gin.ReleaseMode)

	var appList ArgoCDAppList

	os.MkdirAll("./img", os.ModePerm)

	go getArgoApps(&appList)

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
