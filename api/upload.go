package api

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"hub-video-decoder/config"
	"net/http"
	"os"
	"time"
)

const (
	PathApiPostImage = "%s/tracker/%s"
)

func PostData(path string, body []byte) error {
	client := &http.Client{
		Timeout: RequestTimeout * time.Second,
	}

	req, err := http.NewRequest("POST", path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", config.KakaCamHubAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.GetAppConfig().Token))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}

	return nil
}

func PostImage(filename string, camUuid string, image64 []byte) {
	path := fmt.Sprintf(PathApiPostImage, config.GetAppConfig().Server, camUuid)
	err := PostData(path, image64)
	if err != nil {
		glog.Info(err)
	}
	os.Remove(filename)
}
