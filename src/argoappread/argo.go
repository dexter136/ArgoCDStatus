package argoappread

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type ArgoCDAppList struct {
	Apps        []ArgoCDApp `json:"items"`
	LastSync    time.Time
	LastAttempt time.Time
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

func saveArgoBadge(appName, argocdUrl string) {
	img, err := os.Create(fmt.Sprintf("img/%s.svg", appName))
	defer img.Close()
	if err != nil {
		log.Printf("Warning: Unable to create file for ArgoCD Badge for %s. Error is %v", appName, err)
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/badge?name=%s&revision=true", argocdUrl, appName))
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Warning: Unable to gather ArgoCD Badge for %s. Error is %v", appName, err)
	}

	io.Copy(img, resp.Body)
}

func GetArgoApps(AppList *ArgoCDAppList) {

	argocdUrl, test := os.LookupEnv("ARGOCD_URL")
	if !test {
		log.Fatal("Fatal Error: ARGOCD_URL is not set.")
	}
	argocdUser, test := os.LookupEnv("ARGOCD_USER")
	if !test {
		log.Fatal("Fatal Error: ARGOCD_URL is not set.")
	}
	argocdPass, test := os.LookupEnv("ARGOCD_PASS")
	if !test {
		log.Fatal("Fatal Error: ARGOCD_URL is not set.")
	}

	var syncRefresh time.Duration = 120 * time.Second
	syncEnv, test := os.LookupEnv("SYNC_REFRESH")
	if test {
		i, err := strconv.Atoi(syncEnv)
		if err != nil {
			log.Fatalf("Fatal Error: SYNC_REFRESH is set but could not be converted to an integer. Error is %v", err)
		}
		syncRefresh = time.Duration(i) * time.Second
	}

	client := resty.New()

	resp, err := client.R().
		SetBody(map[string]interface{}{"username": argocdUser, "password": argocdPass}).
		Post(fmt.Sprintf("%s/api/v1/session", argocdUrl))

	//Require initial login attempt to be successful to allow app to run
	if err != nil {
		log.Fatalf("Fatal Error: Unable to reach ArgoCD. Error is %v", err)
	}
	if resp.StatusCode() != 200 {
		log.Fatalf("Fatal Error: Bad response from ArgoCD. Response is %v", resp.String())
	}

	for {
		if time.Now().Sub(AppList.LastAttempt) < syncRefresh {
			time.Sleep(syncRefresh - time.Now().Sub(AppList.LastAttempt))
		}
		AppList.LastAttempt = time.Now()

		resp, err = client.R().Get(fmt.Sprintf("%s/api/v1/applications", argocdUrl))
		if err != nil {
			log.Printf("Error: Unable to gather applications from ArgoCD. Error is %v", err)
		} else {
			err := json.Unmarshal(resp.Body(), &AppList)

			if err != nil {
				log.Printf("Warning: Unable to convert ArgoCD response to AppList json. Error is %v", err)
			}
			AppList.LastSync = time.Now()
			for _, app := range AppList.Apps {
				saveArgoBadge(app.Metadata.Name, argocdUrl)
			}
		}
	}
}
