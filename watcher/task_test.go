package watcher

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	"github.com/webmonitor/web-monitor/models"
)

// Unit test function for GetLatestOSCodeName
func (w *Watcher) TestGetLatestOSCodeName(test *testing.T)  {

	config := configs.GetConfig()
	
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
    resp, err := c.Get(config.WebSiteURL)

	if err != nil {
		test.Fatalf(`GetLatestOSCodeName Function Failed`)
	} 

	defer resp.Body.Close()

	codeName, err := w.GetLatestOSCodeName(resp)
	if (err != nil) {
		test.Fatalf(`GetLatestOSCodeName Function Failed`)
	} else {
		if (codeName == "Focal") {
			test.Logf(fmt.Sprintf("GetLatestOSCodeName Function OK : FristKey=%s", codeName))
		} else {
			test.Logf(fmt.Sprintf("GetLatestOSCodeName Function Failed : True Value is 'Focal' but returning FristKey=%s", codeName))
		}
	}
}

// Unit test function for GetInfo
func (w *Watcher)  TestGetInfo(test *testing.T)  {

	config := configs.GetConfig()
	
	var task models.Task
	task.URL = config.WebSiteURL

	var firstKey string;
	firstKey,err := w.GetInfo(&task, config); 
	if (err != nil) {
		test.Fatalf(`GetInfo Function Failed`)
	} else {
		if (firstKey == "focal") {
			test.Logf(fmt.Sprintf("GetInfo Function OK : FristKey=%s", firstKey))
		} else {
			test.Logf(fmt.Sprintf("GetInfo Function Failed : True Value is 'focal' but returning FristKey=%s", firstKey))
		}
	}
}


