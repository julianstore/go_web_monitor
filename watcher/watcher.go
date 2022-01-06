package watcher

import (
	"context"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	models "github.com/webmonitor/web-monitor/models"
)

type Watcher struct {
	WatchInterval time.Duration
	Tasks map[string]context.CancelFunc
}

func New(watchInterval time.Duration) (*Watcher, error) {

	w := Watcher{
		WatchInterval: watchInterval,
		Tasks:         map[string]context.CancelFunc{},
	}

	return &w, nil
}

func (w *Watcher) Run(config *configs.Config) error {
	var task models.Task

	task.URL = config.WebSiteURL

	w.NewTask(&task, config)

	return nil
}



