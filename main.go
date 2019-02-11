package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/jasonlvhit/gocron"
	"hub-video-decoder/decoder"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	glog.Info("Hub video decoder start !!!")
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	go func() {
		decoder.Init()
	}()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go waitTerminate(sigs, done)
	<-done

	decoder.Stop()
	glog.Info("Hub video decoder finish")
}

func waitTerminate(sigs <-chan os.Signal, done chan<- bool) {
	go func() {
		for sig := range sigs {
			switch sig {
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				glog.Info("Request interrupt")
				done <- true
				// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				gocron.Clear()
				done <- true
				glog.Info("Receive terminate")
			default:
				glog.Info("Skip default signal")
			}
			// sig is a ^C, handle it
		}
	}()
}
