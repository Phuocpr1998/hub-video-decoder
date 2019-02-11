package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"time"
)


const KakaCamHubAgent = "X-Kakacam-Hub"
const ConfigFile = "config.json"

const AppVersion = "1.0.1"

var appConfig *AppConfig
var rabbitConnString string

const (
	TimeCheckRegister = 10 * time.Second
	TimeConnectRetry  = 10 * time.Second
)

// App Constant
const (
	IdleRetryOpenInput = 15 * time.Second
	ForceReloadConfigInterval = 3 * 60

	IdleSleep = 3 * time.Second
)

const (
	RabbitConnStringProd       = "amqp://uprovcloudcam:yUEq4Dy1!=RE0ukkdq3Q7SjPXkkd4Dydq3Q7@61.28.232.61:5672/vcloudcam,amqp://uprovcloudcam:yUEq4Dy1!=RE0ukkdq3Q7SjPXkkd4Dydq3Q7@61.28.232.36:5672/vcloudcam,amqp://uprovcloudcam:yUEq4Dy1!=RE0ukkdq3Q7SjPXkkd4Dydq3Q7@61.28.232.42:5672/vcloudcam"
	RabbitConnStringProdBigCam = "amqp://uprobigcam:yUEqkkddq3Q74Dy1!=Rq3Q7SjPXkkd4DyE0u@61.28.232.61:5672/bigcam,amqp://uprobigcam:yUEqkkddq3Q74Dy1!=Rq3Q7SjPXkkd4DyE0u@61.28.232.36:5672/bigcam,amqp://uprobigcam:yUEqkkddq3Q74Dy1!=Rq3Q7SjPXkkd4DyE0u@61.28.232.42:5672/bigcam"
	RabbitConnStringProdButton = "amqp://upronutbam:Jm$MOOo$u5I3A$v06Oy01DGxy4WWSJfMmLKnTHR@61.28.232.61:5672/nutbam,amqp://upronutbam:Jm$MOOo$u5I3A$v06Oy01DGxy4WWSJfMmLKnTHR@61.28.232.36:5672/nutbam,amqp://upronutbam:Jm$MOOo$u5I3A$v06Oy01DGxy4WWSJfMmLKnTHR@61.28.232.42:5672/nutbam"
	RabbitConnStringStaging    = "amqp://uvcloudcam:Vng_1234567899@61.28.230.30:5672/vcloudcam,amqp://uvcloudcam:Vng_1234567899@61.28.230.36:5672/vcloudcam"
	RabbitConnStringLocal      = "amqp://guest:guest@localhost:5672"
)

func init() {
	appConfig = &AppConfig{}
	appConfig.Parse()

	if appConfig.RunMode == "prod" {
		switch appConfig.Agency {
		case "bgc":
			rabbitConnString = RabbitConnStringProdBigCam
		case "btn":
			rabbitConnString = RabbitConnStringProdButton
		default:
			rabbitConnString = RabbitConnStringProd
		}
	} else if appConfig.RunMode == "staging" {
		rabbitConnString = RabbitConnStringStaging
	} else {
		rabbitConnString = RabbitConnStringLocal
	}
}

type AppConfig struct {
	Token   string `json:"token"`
	Agency  string `json:"agency"`
	Model   string `json:"model"`
	IntFace string `json:"interface"`
	Server  string `json:"server"`
	Serial  string `json:"serial"`
	RunMode string `json:"run"`
}

//Parse Config
func (config *AppConfig) Parse() error {
	raw, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, config)
}

func GetRabbitConnString() string {
	return rabbitConnString
}

func (config *AppConfig) Save() error {
	str, _ := json.MarshalIndent(config, "", "\t")
	ioutil.WriteFile(ConfigFile, str, 0644)
	return nil
}

func (config *AppConfig) Mac() string {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				addr := i.HardwareAddr.String()
				inf := i.Name
				if strings.Compare(config.IntFace, inf) == 0 {
					return addr
				}
				// break
			}
		}
	}
	return ""
}

func (config *AppConfig) CpuSerial() string {
	serial, err := exec.Command("sh", "-c", "cat /proc/cpuinfo | grep -i serial | awk '{print $3}'").Output()
	if err == nil {
		return string(serial)
	}
	return ""
}
func (config *AppConfig) FactoryReset() {
	GetAppConfig().Token = ""
	GetAppConfig().Serial = ""
	GetAppConfig().Save()
}

func GetAppConfig() *AppConfig {
	return appConfig
}
