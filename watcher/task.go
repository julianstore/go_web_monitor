package watcher

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	"github.com/webmonitor/web-monitor/constants"
	"github.com/webmonitor/web-monitor/models"
	"golang.org/x/net/html"
)

// Add a new task to the list and run it.
func (w *Watcher) NewTask(task *models.Task, config *configs.Config) {
	ctx, _ := context.WithCancel(context.Background())

	if err := w.RunTask(ctx, task, config); err != nil {
		log.Printf("Failed to run task for %s. Error: %s\n", task.URL, err)
	}
}


func (w *Watcher) GetLatestOSCodeName(resp *http.Response) ( string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse page as HTML.
	doc, err := html.Parse(bytes.NewBuffer(body))
	if err != nil {
		return  "", err
	}

	versionList, codeNameList := ExtractInfo(doc);

	index := FindLatestIndex(versionList)
	
	return codeNameList[index], nil
}


func (w *Watcher) GetInfo(task *models.Task,config *configs.Config) ( string, error) {
	t := &http.Transport{
            Dial: (&net.Dialer{
                    Timeout: 60 * time.Second,
                    KeepAlive: 30 * time.Second,
            }).Dial,
            // We use ABSURDLY large keys, and should probably not.
            TLSHandshakeTimeout: 60 * time.Second,
    }
    c := &http.Client{
            Transport: t,
    }
    resp, err := c.Get(task.URL)

	if err != nil {
		return "",fmt.Errorf("failed to fetch task URL: %s", err.Error())
	} else if resp.StatusCode != constants.ResponseCodeSuccess {
		return "",fmt.Errorf("responseCode %d",resp.StatusCode)
	}

	defer resp.Body.Close()

	codeName, err := w.GetLatestOSCodeName(resp)

	firstKey := strings.ToLower(strings.Split(codeName, " ")[0])

	if err != nil {
		return "", fmt.Errorf("failed to get latest code name. Error: %s", err)
	}

	return firstKey,nil
}

// Run the task every X minutes.
func (w *Watcher) RunTask(ctx context.Context, task *models.Task, config *configs.Config) error {
	
	log.Println("Crawling:", task.URL)

	var firstKey string;
	firstKey,err := w.GetInfo(task, config); 
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	fileName := firstKey+"-server-cloudimg-amd64.ova"

	downloadUrl := config.DownloadBaseURL+firstKey+"/current/"+fileName;

	err = DownloadFile(ctx, fileName, downloadUrl, config)

	if err != nil {
		fmt.Println(fmt.Errorf("DOWNLOAD FAILED: %s",err.Error()))
	} else {
		status := ValidateInfo(fileName,config) 
		if status {
			err = UploadToVCenter(ctx, fileName, config)
			if err != nil {
				fmt.Println(fmt.Errorf("UPLOAD FILE FAILED : %s",err.Error()))
			} else {
				err = UpdateInfo(fileName)
				if err != nil {
					fmt.Println(fmt.Errorf("UPDATE INFO FAILED: %s",err.Error()))
				}
			}
		} else {
				fmt.Println("ALREADY UPLOADED FILE")
		}
		
	}

	select {
	case <-time.After(w.WatchInterval):
		return w.RunTask(ctx, task, config)
	case <-ctx.Done():
		log.Println("Stopped task for", task.URL)
		return nil
	}
}
