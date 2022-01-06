package watcher

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/webmonitor/web-monitor/configs"
)


func Authenticate(client *http.Client, config *configs.Config) (string, error) {

	url := config.VCenterIP + "/rest/com/vmware/cis/session"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(config.VCenterUserName, config.VCenterUserPwd)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	return res["value"], nil
}

func GetDataStoreList(client *http.Client, sessionID string, config *configs.Config) (string, error) {

	url := config.VCenterIP + "/rest/vcenter/datastore"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var res map[string][]map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	res_data := res["value"]

	return res_data[0]["datastore"], nil

}

func CreateLibrary(client *http.Client, sessionID, datatStoreID string, config *configs.Config) (string, error) {

	libraryData := map[string]interface{}{
		"create_spec": map[string]interface{}{
			"name":        "OS Lib",
			"description": "Latest OS Lib",
			"publish_info": map[string]interface{}{
				"authentication_method": "NONE",
				"persist_json_enabled":  false,
				"published":             false,
			},
			"storage_backings": []map[string]interface{}{
				{

					"datastore_id": datatStoreID,
					"type":         "DATASTORE",
				},
			},
			"type": "LOCAL",
		},
	}

	json_data, err := json.Marshal(libraryData)
	if err != nil {
		return "", err
	}

	url := config.VCenterIP + "/rest/com/vmware/content/local-library"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	return res["value"], nil

}

func AddItemToLibrary(client *http.Client, sessionID, libraryID, fileName string, config *configs.Config) (string, error) {

	libraryData := map[string]interface{}{
		"create_spec": map[string]string{
			"description": "ova file of Latest OS",
			"library_id":  libraryID,
			"type":        "ovf",
			"name":        fileName,
		},
	}

	json_data, err := json.Marshal(libraryData)
	if err != nil {
		return "", err
	}

	url := config.VCenterIP + "/rest/com/vmware/content/library/item"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	return res["value"], nil

}

func CreateUpdateSession(client *http.Client, sessionID, libraryItem string, config *configs.Config) (string, error) {

	sessionData := map[string]interface{}{
		"create_spec": map[string]string{
			"library_item_id": libraryItem,
		},
	}

	json_data, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}

	url := config.VCenterIP + "/rest/com/vmware/content/library/item/update-session"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	return res["value"], nil

}

func GetEndPoint(client *http.Client, fileName, sessionID, updateSession string, config *configs.Config) (string, error) {


	size, err := GetFileSize(fileName)
	if err != nil {
		return "", err
	}

	sessionData := map[string]interface{}{
		"file_spec": map[string]interface{}{
			"name":        fileName,
			"size":        strconv.Itoa(size),
			"source_type": "PUSH",
		},
	}

	json_data, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}

	url := config.VCenterIP + "/rest/com/vmware/content/library/item/updatesession/file/id:" + updateSession + "?~action=add"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res map[string]map[string]map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	res_value := res["value"]["upload_endpoint"]["uri"]

	return res_value.(string), nil

}

func UploadFile(client *http.Client, fileName, uploadEndpoint, sessionID string) (status bool, err error) {


	 values := map[string]io.Reader{
        "file":  mustOpen(fileName), // lets assume its this file
    }

	var b bytes.Buffer
    w := multipart.NewWriter(&b)
    for key, r := range values {
        var fw io.Writer
        if x, ok := r.(io.Closer); ok {
            defer x.Close()
        }
        // Add an image file
        if x, ok := r.(*os.File); ok {
            if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
                return false, err
            }
        } 
        if _, err = io.Copy(fw, r); err != nil {
                return false, err
        }

    }
    // Don't forget to close the multipart writer.
    // If you don't close it, your request will be missing the terminating boundary.
    w.Close()
	

	req, err := http.NewRequest("PUT", uploadEndpoint, &b)
	if err != nil {
		return false, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}

func Validate(client *http.Client,  sessionID, updateSession string, config *configs.Config) (string, error) {

	url := config.VCenterIP + "/rest/com/vmware/content/library/item/updatesession/file/id:" + updateSession + "?~action=validate"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("vmware-api-session-id", sessionID)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var res map[string]map[string]map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	res_value := res["value"]["upload_endpoint"]["uri"]

	return res_value.(string), nil

}

func mustOpen(f string) *os.File {
    r, err := os.Open(f)
    if err != nil {
        panic(err)
    }
    return r
}
