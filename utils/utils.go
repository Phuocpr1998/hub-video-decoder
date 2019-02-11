package utils

import (
	"bufio"
	"encoding/base64"
	"errors"
	"github.com/streadway/amqp"
	"hub-video-decoder/config"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func GetCurrentIP() (string, error) {
	itf, err := net.InterfaceByName(config.GetAppConfig().IntFace) //here your interface

	if err != nil {
		return "", err
	}

	item, _ := itf.Addrs()
	var ip net.IP
	for _, addr := range item {
		switch v := addr.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				if v.IP.To4() != nil { //Verify if IP is IPV4
					ip = v.IP
				}
			}
		}
	}

	if ip != nil {
		return ip.String(), nil
	} else {
		return "", errors.New("ssdp.ip.localip")
	}
}

func DialCluster(connConfig string) (*amqp.Connection, error) {
	urls := strings.Split(connConfig, ",")
	var url string
	if len(urls) > 1 {
		rand.Seed(time.Now().UnixNano())

		url = urls[rand.Intn(len(urls)-1)]
	} else {
		url = urls[0]
	}

	return amqp.Dial(url)
}

func Base64Encoder(filename string) (string, error) {
	imgFile, err := os.Open(filename) // open file

	if err != nil {
		return "", err
	}

	defer imgFile.Close()

	// create a new buffer base on file size
	fInfo, _ := imgFile.Stat()
	var size int64 = fInfo.Size()
	buf := make([]byte, size)

	// read file content into buffer
	fReader := bufio.NewReader(imgFile)
	_, err = fReader.Read(buf)
	if err != nil {
		return "", err
	}

	imgBase64Str := base64.StdEncoding.EncodeToString(buf)
	return imgBase64Str, nil
}
