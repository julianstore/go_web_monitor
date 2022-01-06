package main

import (
	"log"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	"github.com/webmonitor/web-monitor/watcher"
)


func main() {

	config := configs.GetConfig()

	instance, err := watcher.New(time.Duration(config.TimeInterval)*time.Hour)

	if err != nil {
		log.Fatalln(err)
	}

	if err := instance.Run(config); err != nil {
		log.Fatalln("failed to run tasks:", err)
	}



}

