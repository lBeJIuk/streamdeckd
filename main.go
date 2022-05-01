package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/unix-streamdeck/api"
	"github.com/unix-streamdeck/driver"
	"github.com/unix-streamdeck/streamdeckd/handlers"
	"github.com/unix-streamdeck/streamdeckd/handlers/examples"
	"golang.org/x/sync/semaphore"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type VirtualDev struct {
	Deck    streamdeck.Device
	Page    int
	Profile string
	IsOpen  bool
	Config  []api.Page
}

var devs map[string]*VirtualDev
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
	examples.RegisterBaseModules()
	loadConfig()
	devs = make(map[string]*VirtualDev)
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
		dev := &VirtualDev{}
		dev, _ = openDevice()
		if dev.IsOpen {
			SetPage(dev, dev.Page)
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
			go Listen(dev)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func disconnect(dev *VirtualDev) {
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

func openDevice() (*VirtualDev, error) {
	ctx := context.Background()
	err := connectSem.Acquire(ctx, 1)
	if err != nil {
		return &VirtualDev{}, err
	}
	defer connectSem.Release(1)
	d, err := streamdeck.Devices()
	if err != nil {
		return &VirtualDev{}, err
	}
	if len(d) == 0 {
		return &VirtualDev{}, errors.New("No streamdeck devices found")
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
					return &VirtualDev{}, err
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
		return &VirtualDev{}, errors.New("No streamdeck devices found")
	}
	err = device.Open()
	if err != nil {
		return &VirtualDev{}, err
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
			page = append(page, api.Key{})
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
	dev := &VirtualDev{
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
	if len(config.Modules) > 0 {
		for _, module := range config.Modules {
			handlers.LoadModule(module)
		}
	}
}

func readConfig() (*api.Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &api.Config{}, err
	}
	var config api.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return &api.Config{}, err
	}
	config, err = checkConfig(data, config)
	return &config, nil
}

func checkConfig(data []byte, config api.Config) (api.Config, error) {
	if config.Decks == nil {
		var deprecatedConfig api.DepracatedConfig
		err := json.Unmarshal(data, &deprecatedConfig)
		if err != nil {
			return api.Config{}, err
		}
		config = api.Config{Modules: deprecatedConfig.Modules, Decks: []api.Deck{{Serial: "", Profiles: []api.Profile{{Pages: deprecatedConfig.Pages, Name: "default profile"}}}}}
		migrateConfig = 1
	} else if config.Decks[0].Profiles == nil {
		var deprecatedConfig api.DepracatedConfig_v2
		err := json.Unmarshal(data, &deprecatedConfig)
		if err != nil {
			return api.Config{}, err
		}
		config = api.Config{Modules: deprecatedConfig.Modules, Decks: []api.Deck{}}
		for _, deck := range deprecatedConfig.Decks {
			newDecks := api.Deck{Serial: deck.Serial, Profiles: []api.Profile{{Name: "default profile", Pages: deck.Pages}}}
			config.Decks = append(config.Decks, newDecks)
		}
		for _, deck := range config.Decks {
			for _, profile := range deck.Profiles {
				for _, page := range profile.Pages {
					for _, key := range page {
						if key.Icon != "" {
							img, err := ioutil.ReadFile(key.Icon)
							if err != nil {
								var base64Encoding string
								// Determine the content type of the image file
								mimeType := http.DetectContentType(img)
								// Prepend the appropriate URI scheme header depending on the MIME type
								switch mimeType {
								case "image/jpeg":
									base64Encoding += "data:image/jpeg;base64,"
								case "image/png":
									base64Encoding += "data:image/png;base64,"
								}
								// Append the base64 encoded output
								base64Encoding += base64.StdEncoding.EncodeToString(img)
								key.Icon = base64Encoding
							}
						}
					}
				}
			}
		}
		migrateConfig = 2
	}
	return config, nil
}

func runCommand(command string) {
	go func() {
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/nohup "+command)

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pgid:      0,
			Pdeathsig: syscall.SIGHUP,
		}
		if err := cmd.Start(); err != nil {
			fmt.Println("There was a problem running ", command, ":", err)
		} else {
			pid := cmd.Process.Pid
			cmd.Process.Release()
			fmt.Println(command, " has been started with pid", pid)
		}
	}()
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
		SetPage(dev, devs[s].Page)
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
		SetPage(dev, devs[s].Page)
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

func unmountDevHandlers(dev *VirtualDev) {
	for i := range dev.Config {
		unmountPageHandlers(dev.Config[i])
	}
}

func unmountPageHandlers(page api.Page) {
	for i2 := 0; i2 < len(page); i2++ {
		key := &page[i2]
		if key.IconHandlerStruct != nil {
			log.Printf("Stopping %s\n", key.IconHandler)
			if key.IconHandlerStruct.IsRunning() {
				go func() {
					key.IconHandlerStruct.Stop()
					log.Printf("Stopped %s\n", key.IconHandler)
				}()
			}
		}
	}
}
