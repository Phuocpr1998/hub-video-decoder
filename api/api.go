package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"hub-video-decoder/config"
	"hub-video-decoder/models"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	PathApiGetBaseConfig      = "%s/core/v1/hubs/base-config"
)

// Stat
const (
	RequestTimeout = 10
)

func GetBaseConfig() (*models.BaseConfigResponse, error) {
	if config.GetAppConfig().RunMode == "dev" {
		return GetBaseConfigMock()
	}

	var err error
	path := fmt.Sprintf(PathApiGetBaseConfig, config.GetAppConfig().Server)
	rr := &models.BaseConfigResponse{}
	if err = GetConfig(path, rr); err != nil {
		glog.Errorf("[Base Conf] %v", err)
		return nil, err
	} else {
		return rr, nil
	}
}

func GetBaseConfigMock() (*models.BaseConfigResponse, error) {
	var err error
	rr := &models.BaseConfigResponse{}
	base, err := ioutil.ReadFile("config.rest.base.json")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(base, rr); err != nil {
		glog.Errorf("error %v", err)
	}

	return rr, nil
}

func GetConfig(path string, ires interface{}) error {
	h := models.Hub{Model: config.GetAppConfig().Model, Mac: config.GetAppConfig().Mac(), Serial: config.GetAppConfig().Serial}
	str, _ := json.Marshal(h)
	client := &http.Client{
		Timeout: RequestTimeout * time.Second,
	}

	req, err := http.NewRequest("POST", path, bytes.NewReader(str))
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, ires)

	return err
}