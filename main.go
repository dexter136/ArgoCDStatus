package main

import (
	"log"
	"os"
)

func main() {

	var appList ArgoCDAppList

	err := os.Mkdir("./img", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	go GetArgoApps(&appList)

	webServer("/img", "status.tmpl", &appList, true)

}
