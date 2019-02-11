package models

type BaseConfigInfo struct {
	CamUuid string `json:"camUuid"`
	Name    string `json:"name"`
	Input   string `json:"input"`
}

type BaseConfigResponse struct {
	BaseConfig []BaseConfigInfo `json:"baseConfig"`
}

type Hub struct {
	Model     string `json:"model"`
	Serial    string `json:"serial"`
	Mac       string `json:"mac"`
	Agency    string `json:"agency"`
	CpuSerial string `json:"cpuSerial"`
}
