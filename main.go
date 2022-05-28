package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/lBeJIuk/streamdeckd/handlers"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/unix-streamdeck/api"
	"github.com/unix-streamdeck/driver"
	"golang.org/x/sync/semaphore"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var devs map[string]*utils.VirtualDev
var config *api.Config
var migrateConfig = 0
var configPath string
var disconnectSem = semaphore.NewWeighted(1)
var connectSem = semaphore.NewWeighted(1)
var basicConfig = api.Config{
	Modules: []string{},
	Decks: []api.Deck{
		{},
	},
}
var isRunning = true

func main() {
	checkOtherRunningInstances()
	configPtr := flag.String("config", configPath, "Path to config file")
	flag.Parse()
	if *configPtr != "" {
		configPath = *configPtr
	} else {
		basePath := os.Getenv("HOME") + string(os.PathSeparator) + ".config"
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			basePath = os.Getenv("XDG_CONFIG_HOME")
		}
		configPath = basePath + string(os.PathSeparator) + ".streamdeck-config.json"
	}
	cleanupHook()
	go InitDBUS()
	go InitWS()
	handlers.InitHandlers()
	loadConfig()
	devs = make(map[string]*utils.VirtualDev)
	attemptConnection()
}

func checkOtherRunningInstances() {
	processes, err := process.Processes()
	if err != nil {
		log.Println("Could not check for other instances of streamdeckd, assuming no others running")
	}
	for _, proc := range processes {
		name, err := proc.Name()
		if err == nil && name == "streamdeckd" && int(proc.Pid) != os.Getpid() {
			log.Fatalln("Another instance of streamdeckd is already running, exiting...")
		}
	}
}

func attemptConnection() {
	for isRunning {
		dev := &utils.VirtualDev{}
		dev, _ = openDevice()
		if dev.IsOpen {
			RenderPage(dev, dev.Page)
			found := false
			for i := range sDInfo {
				if sDInfo[i].Serial == dev.Deck.Serial {
					found = true
				}
			}
			if !found {
				sDInfo = append(sDInfo, api.StreamDeckInfo{
					Cols:     int(dev.Deck.Columns),
					Rows:     int(dev.Deck.Rows),
					IconSize: int(dev.Deck.Pixels),
					Page:     0,
					Serial:   dev.Deck.Serial,
				})
			}
			go listen(dev)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func listen(dev *utils.VirtualDev) {
	kch, err := dev.Deck.ReadKeys()
	if err != nil {
		log.Println(err)
	}
	for dev.IsOpen {
		select {
		case k, ok := <-kch:
			if !ok {
				disconnect(dev)
				return
			}
			if k.Pressed == true {
				if len(dev.Config)-1 >= dev.Page && len(dev.Config[dev.Page])-1 >= int(k.Index) {
					HandleInput(dev, &dev.Config[dev.Page][k.Index], dev.Page)
				}
			}
		}
	}
}

func disconnect(dev *utils.VirtualDev) {
	ctx := context.Background()
	err := disconnectSem.Acquire(ctx, 1)
	if err != nil {
		return
	}
	defer disconnectSem.Release(1)
	if !dev.IsOpen {
		return
	}
	log.Println("Device (" + dev.Deck.Serial + ") disconnected")
	_ = dev.Deck.Close()
	dev.IsOpen = false
	unmountDevHandlers(dev)
}

func openDevice() (*utils.VirtualDev, error) {
	ctx := context.Background()
	err := connectSem.Acquire(ctx, 1)
	if err != nil {
		return &utils.VirtualDev{}, err
	}
	defer connectSem.Release(1)
	d, err := streamdeck.Devices()
	if err != nil {
		return &utils.VirtualDev{}, err
	}
	if len(d) == 0 {
		return &utils.VirtualDev{}, errors.New("No streamdeck devices found")
	}
	device := streamdeck.Device{Serial: ""}
	for i := range d {
		found := false
		for s := range devs {
			if d[i].ID == devs[s].Deck.ID && devs[s].IsOpen {
				found = true
				break
			} else if d[i].Serial == s && !devs[s].IsOpen {
				err = d[i].Open()
				if err != nil {
					return &utils.VirtualDev{}, err
				}
				devs[s].Deck = d[i]
				devs[s].IsOpen = true
				return devs[s], nil
			}
		}
		if !found {
			device = d[i]
		}
	}
	if len(device.Serial) != 12 {
		return &utils.VirtualDev{}, errors.New("No streamdeck devices found")
	}
	err = device.Open()
	if err != nil {
		return &utils.VirtualDev{}, err
	}
	devNo := -1
	if migrateConfig != 0 {
		switch migrateConfig {
		case 1:
			config.Decks[0].Serial = device.Serial
			break
		}
		_ = SaveConfig()
		migrateConfig = 0
	}
	for i := range config.Decks {
		if config.Decks[i].Serial == device.Serial {
			devNo = i
		}
	}
	if devNo == -1 {
		var pages []api.Page
		page := api.Page{}
		for i := 0; i < int(device.Rows)*int(device.Columns); i++ {
			page = append(page, api.KeyConfig{})
		}
		pages = append(pages, page)
		config.Decks = append(config.Decks, api.Deck{
			Serial: device.Serial,
			Profiles: []api.Profile{{
				Name:  "default Profile",
				Pages: pages,
			}},
		})
		devNo = len(config.Decks) - 1
	}
	dev := &utils.VirtualDev{
		Deck:    device,
		Page:    0,
		Profile: config.Decks[devNo].Profiles[0].Name,
		IsOpen:  true,
		Config:  config.Decks[devNo].Profiles[0].Pages,
	}
	devs[device.Serial] = dev
	log.Println("Device (" + device.Serial + ") connected")
	return dev, nil
}

func loadConfig() {
	var err error
	config, err = readConfig()
	if err != nil && !os.IsNotExist(err) {
		log.Println(err)
	} else if os.IsNotExist(err) {
		file, err := os.Create(configPath)
		if err != nil {
			log.Println(err)
		}
		err = file.Close()
		if err != nil {
			log.Println(err)
		}
		config = &basicConfig
		err = SaveConfig()
		if err != nil {
			log.Println(err)
		}
	}
	//if len(config.Modules) > 0 {
	//	for _, module := range config.Modules {
	//		handlers.LoadModule(module)
	//	}
	//}
}

func readConfig() (*api.Config, error) {
	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &api.Config{}, err
	}
	var config api.Config
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		return &api.Config{}, err
	}
	config, err = checkConfig(rawConfig, config)
	return &config, nil
}

func checkConfig(rawConfig []byte, config api.Config) (api.Config, error) {
	return config, nil
}

func cleanupHook() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT)
	go func() {
		<-sigs
		log.Println("Cleaning up")
		isRunning = false
		unmountHandlers()
		var err error
		for s := range devs {
			if devs[s].IsOpen {
				err = devs[s].Deck.Reset()
				if err != nil {
					log.Println(err)
				}
				err = devs[s].Deck.Close()
				if err != nil {
					log.Println(err)
				}
			}
		}
		os.Exit(0)
	}()
}

func SetConfig(configString string) error {
	unmountHandlers()
	var err error
	config = nil
	err = json.Unmarshal([]byte(configString), &config)
	if err != nil {
		return err
	}
	for s := range devs {
		dev := devs[s]
		for i := range config.Decks {
			if dev.Deck.Serial == config.Decks[i].Serial {
				for _, profile := range config.Decks[i].Profiles {
					if profile.Name == dev.Profile {
						dev.Config = profile.Pages
						break
					}
				}
				break
			}
		}
		RenderPage(dev, devs[s].Page)
	}
	return nil
}

func ReloadConfig() error {
	unmountHandlers()
	loadConfig()
	for s := range devs {
		dev := devs[s]
		for i := range config.Decks {
			if dev.Deck.Serial == config.Decks[i].Serial {
				for _, profile := range config.Decks[i].Profiles {
					if profile.Name == dev.Profile {
						dev.Config = profile.Pages
						break
					}
				}
				break
			}
		}
		RenderPage(dev, devs[s].Page)
	}
	return nil
}

func SaveConfig() error {
	f, err := os.OpenFile(configPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	var configString []byte
	configString, err = json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = f.Write(configString)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}

func unmountHandlers() {
	for s := range devs {
		dev := devs[s]
		unmountDevHandlers(dev)
	}
}

func unmountDevHandlers(dev *utils.VirtualDev) {
	for i := range dev.Config {
		unmountPageHandlers(dev.Config[i])
	}
}

func unmountPageHandlers(page api.Page) {
	//for i2 := 0; i2 < len(page); i2++ {
	//	key := &page[i2]
	//	if key.IconHandlerStruct != nil {
	//		log.Printf("Stopping %s\n", key.IconHandler)
	//		if key.IconHandlerStruct.IsRunning() {
	//			go func() {
	//				key.IconHandlerStruct.Stop()
	//				log.Printf("Stopped %s\n", key.IconHandler)
	//			}()
	//		}
	//	}
	//}
}

func HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	handler, err := handlers.GetHandler(key)
	if err != nil {
		log.Println(err)
		return
	}
	handler.HandleInput(dev, key, page)
}
