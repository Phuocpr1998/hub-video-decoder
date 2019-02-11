package decoder

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/jasonlvhit/gocron"
	"hub-video-decoder/api"
	"hub-video-decoder/config"
	"hub-video-decoder/models"
	"sync"
	"time"
)

const (
	ControlReload = 1
)

type StreamManager struct {
	Remuxers map[string]*Remuxer
	LastUpdate time.Time
	control chan int
	quit chan int
}

var singleton *StreamManager
var once sync.Once

func getInstance() *StreamManager {
	once.Do(func() {
		singleton = &StreamManager{}
	})
	return singleton
}

func GetRemuxers() (map[string]*Remuxer, time.Time) {
	s := getInstance()
	return s.Remuxers, s.LastUpdate
}

func Init() error {
	s := getInstance()

	s.Remuxers = make(map[string]*Remuxer)
	s.control = make(chan int, 10)
	s.quit = make(chan int)

	//message.RegisterHandler(message.BaseChange, s.handleBaseChanged)

	go func() {
		work()
	}()

	controlReload()
	gocron.Every(config.ForceReloadConfigInterval).Seconds().Do(controlReload)

	return nil
}

func control(control int) {
	s := getInstance()
	select {
	case s.control <- control:
		glog.Infof("Control ok %d", control)
	default:
		glog.Warningf("Stream manager worker busy")
	}
}

func controlReload() {
	control(ControlReload)
}

func work() {
	s := getInstance()
	for {
		select {
		case control := <-s.control:
			switch control {
			case ControlReload:
				reload()
			}
		case <-s.quit:
			glog.Info("Exit worker")
			return
			//default:
			//	time.Sleep(50 * time.Millisecond)
		}
	}
}


//Reload all config
func reload() error {
	glog.Info("Do reload config")
	s := getInstance()

	if cfg, err := s.getAllConfig(); err == nil {
		if s.checkReloadStreams(cfg) == true {
			glog.Info("Config has changed")
			s.LastUpdate = time.Now()
		}
	} else {
		glog.Errorf("Error when reload full config %v", err)
	}

	return nil
}


func (s *StreamManager) existsStream(stream models.BaseConfigInfo) bool {
	return s.Remuxers[stream.Name] != nil
}


func (s *StreamManager) addStream(stream models.BaseConfigInfo) {
	glog.Infof("Add new stream %s, input %s", stream.Name, stream.Input)
	// Init stream status
	remuxer := &Remuxer{
		Running: false,
		CamUuid: stream.CamUuid,
		Ctx: StreamContext{
			InputFileNameNew:   stream.Input,
		},
		RequestStop: false,
		Name:        stream.Name,
	}

	remuxer.CfgHash = spew.Sdump(stream)

	// Add stream
	s.Remuxers[stream.Name] = remuxer
	// Run stream
	remuxer.Run()
}

func (s *StreamManager) removeStream(streamName string) {
	glog.Info("Remove stream %s", streamName)
	// Stop stream
	remuxer := s.Remuxers[streamName]
	if remuxer != nil {
		remuxer.Stop()
		// Remove stream
		delete(s.Remuxers, streamName)
	}
}

func (s *StreamManager) updateStream(stream models.BaseConfigInfo) {
	glog.Infof("Update stream %s, input %s", stream.Name, stream.Input)
	// Update stream status
	remuxer := s.Remuxers[stream.Name]
	if remuxer != nil {
		remuxer.CfgHash = spew.Sdump(stream)
		remuxer.Update(stream)
	}
}

func Stop() {
	s := getInstance()
	s.stop()
}

func (s *StreamManager) stop() error {
	if s.quit != nil {
		glog.Info("Stop All remuxer")
		for _, stream := range s.Remuxers {
			stream.Stop()
		}
		s.quit <- 0
	}

	return nil
}

func (s *StreamManager) checkReloadStreams(Streams []models.BaseConfigInfo) bool {
	// Check new stream
	changed := false
	for _, stream := range Streams {
		if s.existsStream(stream) {
			// Exists current streams
			if spew.Sdump(stream) != s.Remuxers[stream.Name].CfgHash {
				glog.Infof("Stream %s has change config => Reload it", stream.Name)
				s.updateStream(stream)
				changed = true
			}
		} else {
			// No exists current streams
			s.addStream(stream)
			changed = true
		}
	}

	var removeStreams []string
	for _, stream := range s.Remuxers {
		found := false
		for _, newStream := range Streams {
			if stream.Name == newStream.Name {
				found = true
				break
			}
		}
		if !found {
			removeStreams = append(removeStreams, stream.Name)
			changed = true
		}
	}

	for _ , stream := range removeStreams {
		s.removeStream(stream)
	}

	return changed
}

func (p *StreamManager) getAllConfig() ([]models.BaseConfigInfo, error) {
	var baseConfig *models.BaseConfigResponse
	//Get base config
	var err error
	if baseConfig, err = api.GetBaseConfig(); err != nil {
		glog.Errorf("Get base config error %v\n", err)
		return nil, err
	}

	//Merge config
	var baseInfo []models.BaseConfigInfo

	for _, base := range baseConfig.BaseConfig {
		baseInfo = append(baseInfo, base)
	}
	return baseInfo, nil
}


func DumpStreamInfo() {
	s := getInstance()
	for _, remux := range s.Remuxers {
		remux.dumpInfo()
	}
}

func (s *StreamManager) handleBaseChanged(msg string) error {
	controlReload()
	return nil
}


