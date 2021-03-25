package repoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.amplifyedge.org/booty-v2/internal/downloader"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/reposerver"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
	chunkSize      = 1 << 20
)

func AuthCli(srvAddr, user, password string) error {
	hc := http.Client{
		Timeout: defaultTimeout,
	}
	ar := reposerver.AuthRequest{
		User:     user,
		Password: password,
	}
	body, err := json.Marshal(&ar)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(body)

	r, err := http.NewRequest(http.MethodGet, srvAddr, buf)
	if err != nil {
		return err
	}
	r.Header.Set("accept", "application/json")

	res, err := hc.Do(r)
	if err != nil {
		return err
	}

	m := map[string]interface{}{}
	if err = json.NewDecoder(r.Body).Decode(&m); err != nil {
		return err
	}

	saveTokenPath := filepath.Join(osutil.GetDataDir(), "upload-token")
	if err = ioutil.WriteFile(saveTokenPath, []byte(m["token"].(string)), 0600); err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, res.Body)
	if err != nil {
		return err
	}

	return res.Body.Close()
}

func detectContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func UploadCli(srvAddr, filename string) (string, error) {
	hc := http.Client{
		Timeout: defaultTimeout,
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filetype, err := detectContentType(file)
	if err != nil {
		return "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filetype, filepath.Base(file.Name()))
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	writer.Close()
	request, err := http.NewRequest("POST", srvAddr+"/upload/", body)

	if err != nil {
		return "", err
	}

	fileStats, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := strconv.FormatInt(fileStats.Size(), 10)

	// read saved token
	saveTokenPath := filepath.Join(osutil.GetDataDir(), "upload-token")
	tkn, err := ioutil.ReadFile(saveTokenPath)
	if err != nil {
		return "", err
	}

	request.Header.Add("Tus-Resumable", "1.0.0")
	request.Header.Add("Upload-Offset", "0")
	request.Header.Add("Content-Length", fileSize)
	request.Header.Add("Upload-Length", fileSize)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", "Bearer "+string(tkn))

	response, err := hc.Do(request)

	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		fmt.Println("There was an error with the request ", response.StatusCode)
	} else {
		r := regexp.MustCompile(`files\/upload\/(.*?)\+`)
		subStrings := r.FindStringSubmatch(response.Header["Location"][0])
		uploadedFileId := subStrings[1]
		fmt.Println("stored File Name "+uploadedFileId, " status code: ", response.StatusCode)
		return uploadedFileId, nil
	}

	return "", fmt.Errorf("error uploading file")
}

func DownloadCli(serverAddr, fileId, targetDir string) error {
	dlUrl := fmt.Sprintf("%s/dl/%s", serverAddr, fileId)
	return downloader.Download(dlUrl, targetDir)
}
