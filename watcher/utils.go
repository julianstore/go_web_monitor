package watcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/webmonitor/web-monitor/configs"
	"github.com/webmonitor/web-monitor/constants"
	"golang.org/x/net/html"
)

type WriteCounter struct {
	Total   uint64
	Context context.Context
}

// Write wites the number of bytes written to it.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	return n, nil
}

// A simple helper function to iterate over HTML node.
func CrawlDocument(node *html.Node, handler func(node *html.Node) bool) bool {
	if handler(node) {
		return true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if CrawlDocument(child, handler) {
			return true
		}
	}

	return false
}

func CollectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		CollectText(c, buf)
	}
}

// Extract LTS versions and code Names from HTML page.
func ExtractInfo(doc *html.Node) ([]string, []string) {
	var osVersionList []string
	var codeNameList []string
	CrawlDocument(doc, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "strong" {
			text := &bytes.Buffer{}
			CollectText(node, text)
			if strings.Contains(text.String(), "LTS") {
				osVersionList = append(osVersionList, text.String())
				codeNameNode := node.Parent.Parent.Parent.FirstChild.NextSibling.NextSibling.NextSibling.FirstChild.FirstChild
				codeName := &bytes.Buffer{}
				CollectText(codeNameNode, codeName)
				codeNameList = append(codeNameList, codeName.String())
			}
		}

		return false
	})

	return osVersionList, codeNameList
}

// find index of latest version
func FindLatestIndex(versionList []string) (index int) {
	latestMajor, _ := strconv.Atoi(strings.Split(versionList[0], ".")[0])
	latestMinor, _ := strconv.Atoi(strings.Split(versionList[0], ".")[1])
	index = 0
	for i, value := range versionList {
		majorVersion, _ := strconv.Atoi(strings.Split(value, ".")[0])
		if majorVersion > latestMajor {
			latestMajor = majorVersion
			index = i
		} else if majorVersion == latestMajor {
			minorVersion, _ := strconv.Atoi(strings.Split(value, ".")[1])
			if minorVersion > latestMinor {
				latestMinor = minorVersion
				index = i
			}
		}
	}
	return index
}

func DownloadFile(ctx context.Context, fileName string, url string, config *configs.Config) error {

	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		// We use ABSURDLY large keys, and should probably not.
		TLSHandshakeTimeout: 60 * time.Second,
	}
	c := &http.Client{
		Transport: t,
	}
	resp, err := c.Get(url)

	// Get the data
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != constants.ResponseCodeSuccess {
		return fmt.Errorf("FILE NOT FOUND")
	}

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(fileName + ".tmp")
	if err != nil {
		return err
	}

	fmt.Println("Downloading.....")
	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	counter.Context = ctx
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	// Close the file without defer so it can happen before Rename()
	out.Close()

	fmt.Println("Download finished....")

	return os.Rename(fileName+".tmp", fileName)
}

func ValidateInfo(fileName string ,config *configs.Config) (bool) {
	
	isExist := configs.CheckFileExist(constants.InfoFile)
	if !isExist {
		return true
	} else {
		input, err := ioutil.ReadFile(constants.InfoFile)
		if err != nil {
			return false
		}
		checkSum, err := GetChecksum(fileName)
		if err != nil {
			return false
		}

		fileSize, err := GetFileSize(fileName)
		if err != nil {
			return false
		}

		infoStr := fmt.Sprintf("checkSum=%s:fileSize=%d", checkSum,fileSize)

		if strings.Contains(string(input), infoStr) {
			return false
		} else {
			return true
		} 
	}
	
}

func UpdateInfo(fileName string ) ( error) {

	f, err := os.OpenFile(constants.InfoFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
    }

	checkSum, err := GetChecksum(fileName)
	if err != nil {
		return err
	}

	fileSize, err := GetFileSize(fileName)
	if err != nil {
		return err
	}

	infoStr := fmt.Sprintf("checkSum=%s:fileSize=%d", checkSum,fileSize)

	 _, err = f.Write([]byte(infoStr))
    if err != nil {
       return err
    }
	f.Close()
	
	return nil
}

func UploadToVCenter(ctx context.Context, fileName string, config *configs.Config) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	// get session id
	sessionID, err := Authenticate(client,config)

	if err != nil {
		return fmt.Errorf("Authenticate Failed")
	} else {
		fmt.Println("Authenticating finished successfully")
	}

	// get datastore id
	datatStoreID, err := GetDataStoreList(client, sessionID,config)
	if err != nil {
		return fmt.Errorf("GET DATA STORE LIST FAILED")
	} else {
		fmt.Println("Getting datastore finished successfully")
	}

	// create a new content library
	libraryID, err := CreateLibrary(client, sessionID, datatStoreID,config)
	if err != nil {
		return fmt.Errorf("CREATE LIBRARY FAILED")
	} else {
		fmt.Println("Library created successfully")
	}

	// add item to library
	libraryItemID, err := AddItemToLibrary(client, sessionID, libraryID, fileName,config)
	if err != nil {
		return fmt.Errorf("ADD ITEM TO LIBRARY FAILED")
	} else {
		fmt.Println("Item added to the library successfully")
	}

	// create an update session
	updateSession, err := CreateUpdateSession(client, sessionID, libraryItemID,config)
	if err != nil {
		return fmt.Errorf("CREATE UPDATE SESSION FAILED")
	} else {
		fmt.Println("Update session created successfully")
	}

	// get endpoint for upload
	endPoint, err := GetEndPoint(client, fileName, sessionID, updateSession,config)
	if err != nil {
		return fmt.Errorf("UPDATE SESSION FAILED")
	} else {
		fmt.Println("Get endpoint successfully")
	}

	// upload file
	status, err := UploadFile(client, fileName, endPoint, sessionID)
	if err != nil {
		return fmt.Errorf("UPLOAD FAILED")
	}

	if status {
		fmt.Println("Uploading finished successfully")
	}

	return nil
}

// GetChecksum gets the checksum of file
func GetChecksum(fileName string) (checksum string, err error) {
	checkSum := ""
	_, err = os.Stat(fileName)
	if err == nil {
		f, err := os.Open(fileName)
		if err != nil {
			fmt.Println("ERROR MISSING FILE")
			return "", err
		}
		defer f.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, f); err != nil {
			fmt.Println("ERROR CREATING FILE")
			return "", err
		}

		checkSum = hex.EncodeToString(hasher.Sum(nil))
	} else {
		return "", err
	}

	return checkSum, nil
}

// GetFileSize gets the size of file
func GetFileSize(fileName string) (size int, err error) {
	fi, err := os.Stat(fileName)
	if err == nil {
		size = int(fi.Size())
	} else {
		fmt.Println("ERROR MISSING FILE")
		return 0, err
	}

	return size, nil
}

// CreateFile conf file
// Print success info log on successfully ran command, return error if fail
func CreateFile(ctx context.Context, fileName string, force bool) (string, error) {

	if force {
		var file, err = os.Create(fileName)
		if err != nil {
			return "", fmt.Errorf(fmt.Sprintf("Failed to create file: %v", err))
		}
		defer file.Close()
	} else {
		// check if file exists
		var _, err = os.Stat(fileName)

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create(fileName)
			if err != nil {
				return "", fmt.Errorf(fmt.Sprintf("Failed to create file: %v", err))
			}
			defer file.Close()
		} else {
			return fileName, fmt.Errorf(fmt.Sprintf("File already exists: %s", fileName))
		}
	}


	fmt.Printf("File created: %s\n", fileName)

	return fileName, nil
}

// DeleteFile deletes specified file
func DeleteFile(filePath string) error {
	e := os.Remove(filePath)
	if e != nil {
		return e
	}
	return nil
}

// WriteFile writes a file as data
func WriteFile(fileName string, data string) (err error) {
	// write to file
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)

	return err
}