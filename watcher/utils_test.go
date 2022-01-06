package watcher

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	"github.com/webmonitor/web-monitor/models"
	"golang.org/x/net/html"
)

// Unit test function for ExtractInfo
func TestExtractInfo(test *testing.T)  {

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
		test.Fatalf(`ExtractInfo Function Failed`)
	} 

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fatalf(`ExtractInfo Function Failed`)
	}
	defer resp.Body.Close()

	// Parse page as HTML.
	doc, err := html.Parse(bytes.NewBuffer(body))
	if err != nil {
		test.Fatalf(`ExtractInfo Function Failed`)
	}

	versionList, codeNameList := ExtractInfo(doc);

	if (versionList == nil || codeNameList == nil) {
		test.Fatalf(`ExtractInfo Function Failed`)
	} else {
		test.Logf("ExtractInfo Function OK")

	}

}

// Unit test function for FindLatestIndex
func TestFindLatestIndex(test *testing.T){

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
		test.Fatalf(`FindLatestIndex Function Failed`)
	} 
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fatalf(`FindLatestIndex Function Failed`)
	}

	// Parse page as HTML.
	doc, err := html.Parse(bytes.NewBuffer(body))
	if err != nil {
		test.Fatalf(`FindLatestIndex Function Failed`)
	}

	versionList, _ := ExtractInfo(doc);
	
	index := FindLatestIndex(versionList)
	if index == 5 {
		test.Logf(fmt.Sprintf("FindLatestIndex Function OK : Index=%d", index))
	} else {
		test.Fatalf(`FindLatestIndex Function Failed`)
	}
}

// Unit test function for GetChecksum
func TestGetChecksum(test *testing.T) {
	checkSum, err := GetChecksum("ReadMe.docx")
	if err != nil {
		test.Fatalf(`GetChecksum Function Failed`)
	} else {
		if checkSum == "680b1871eb10046fc64b59b806ad683f1bb83af01cc73564579e4df84f61b135" {
			test.Logf("GetChecksum Function OK")
		} else {
			test.Fatalf(fmt.Sprintf("GetChecksum Function Failed : %s",checkSum))
		}
	}
}

// Unit test function for GetFileSize
func TestGetFileSize(test *testing.T) {
	fileSize, err := GetFileSize("ReadMe.docx")
	if err != nil {
		test.Fatalf(`GetFileSize Function Failed`)
	} else {
		if fileSize == 17751 {
			test.Logf(fmt.Sprintf("GetFileSize Function OK : Size of config file =%d", fileSize))
		} else {
			test.Fatalf(fmt.Sprintf("GetFileSize Function Failed : %d",fileSize))
		}

	}
}

// Unit test function for CreateFile
func TestCreateFile(t *testing.T) {
	var cxt context.Context
	path := "/test_createfile.txt"
	fileName, err := CreateFile(cxt, path, true)
	if err != nil {
		t.Fatalf(`Create File Failed`)
	} else {
		t.Logf(fmt.Sprintf("CreateFile Function OK : %s", fileName))
		newFileName, err := CreateFile(cxt, fileName, false)
		var wanted = fmt.Sprintf("File already exists: %s", path)
		if err.Error() != wanted {
			t.Fatalf(fmt.Sprintf("%s %s", "CreateFile Function Failed", err.Error()))
		} else {
			t.Logf(fmt.Sprintf("Can not Create File OK : %s", newFileName))
		}
	}
}

// Unit test function for DeleteFile
func TestDeleteFile(t *testing.T) {
	var path string
	path = "/1.txt"
	var err = DeleteFile(path)
	if err == nil {
		t.Fatalf("DeleteFile Function failed; No such file 1.txt , but function deletes")
	}

	var cxt context.Context
	path = "/test_deletefile.txt"
	fileName, err := CreateFile(cxt, path, true)
	if err == nil {
		t.Fatalf("Create File Function failed")
	}

	err = DeleteFile(fileName)
	if err == nil {
		t.Logf("DeleteFile Function OK")
	} else {
		t.Fatalf("DeleteFile Function Failed")
	}

}

// Unit test function for WriteFile
func TestWriteFile(t *testing.T) {
	var cxt context.Context
	var path = "/1.txt"
	var writeData = "GoTestData"
	var err = WriteFile(path, writeData)
	if err == nil {
		t.Fatalf("WriteFile Function failed; No such file 1.txt , but function writes")
	}
	path = "/test_writefile.txt"
	fileName, err := CreateFile(cxt, path, true)
	if err == nil {
		t.Fatalf("Create File Function failed")
	}
	err = WriteFile(fileName, writeData)
	if err != nil {
		t.Fatalf("WriteFile Function Failed")
	} else {
		var file, err = os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			t.Fatalf(fmt.Sprintf("Write File Function Failed %s ", err.Error()))
		} else {
			data, err := ioutil.ReadAll(file)
			if err != nil {
				t.Fatalf(fmt.Sprintf("Write File Function Failed %s ", err.Error()))
			} else {
				if string(data) == writeData {
					t.Logf("WriteFile Function OK")
				} else {
					t.Fatalf(fmt.Sprintf("Write File Function Failed write = %s result = %s ", writeData, string(data)))
				}
			}
		}
	}
}

// Unit test function for Download File
func (w *Watcher) TestDownloadFile(t *testing.T) {
	var cxt context.Context
	config := configs.GetConfig()

	var task models.Task

	task.URL = config.WebSiteURL

	var firstKey string;
	firstKey,err := w.GetInfo(&task, config); 
	if err != nil {
		t.Fatalf("DownloadFile Function Failed ")
	}

	fileName := firstKey+"-server-cloudimg-amd64.ova"

	downloadUrl := config.DownloadBaseURL+firstKey+"/current/"+fileName;

	err = DownloadFile(cxt, fileName, downloadUrl, config)
	if err != nil {
		t.Fatalf("DownloadFile Function Failed ")
	} else {
		t.Logf("WriteFile Function OK")
	}
}


// Unit test function for ValidateInfo
func TestValidateInfo(test *testing.T)  {

	config := configs.GetConfig()
	
	status := ValidateInfo("validate_info.conf", config); 
	if (status) {
		test.Fatalf(`ValidateInfo Function Failed`)
	} else {
		test.Logf("ValidateInfo Function OK")
	}
}

// Unit test function for UpdateInfo
func TestUpdateInfo(test *testing.T)  {

	err := UpdateInfo("validate_info.conf"); 
	if (err != nil) {
		test.Fatalf(`UpdateInfo Function Failed`)
	} else {
		test.Logf("UpdateInfo Function OK")
	}
}


// Unit test function for UpdateInfo
func TestUploadToVCenter(test *testing.T)  {

	err := UpdateInfo("validate_info.conf"); 
	if (err != nil) {
		test.Fatalf(`UpdateInfo Function Failed`)
	} else {
		test.Logf("UpdateInfo Function OK")
	}
}


	

