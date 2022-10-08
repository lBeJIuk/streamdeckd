package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type KeyLight struct {
	Dummy
	Type string `json:"type"`
}

var KeyLightHandler = KeyLight{
	Type: "keyLight",
}

type KeyLightOptions struct {
	DefaultOptionsStruct
	KeyLightAction  string `json:"keyLightAction,omitempty"`
	KeyLightAddress string `json:"keyLightAddress,omitempty"`
	KeyLightPort    string `json:"keyLightPort,omitempty"`
}
type keyLightResponse struct {
	NumberOfLights int `json:"numberOfLights"`
	Lights         []struct {
		On          int `json:"on"`
		Brightness  int `json:"brightness"`
		Temperature int `json:"temperature"`
	}
}

func makeKeyLightRequest(currentOption KeyLightOptions, method string, body io.Reader) (*keyLightResponse, error) {
	if len(currentOption.KeyLightAddress) == 0 {
		return nil, errors.New("Address ist not set")
	}
	if len(currentOption.KeyLightPort) == 0 {
		currentOption.KeyLightPort = defaultPort
	}
	url := fmt.Sprintf("http://%s:%s/elgato/lights", currentOption.KeyLightAddress, currentOption.KeyLightPort)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//We Read the response body on the line below.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Convert the body to type string
	keyLight := new(keyLightResponse)
	err = json.Unmarshal(respBody, keyLight)
	if err != nil {
		log.Printf("Error on Unmarshal: %v", err)
		return nil, err
	}
	return keyLight, nil
}

const (
	defaultPort            = "9123"
	defaultRefreshInterval = 5 * time.Second
	brightnessStep         = 10
	temperatureStep        = 30
	maxBrightness          = 100
	minBrightness          = 3
	maxTemperature         = 344
	minTemperature         = 143
)

func (handler *KeyLight) GetType() string {
	return handler.Type
}

func (handler *KeyLight) PrepareKey(dev *utils.VirtualDev, key *api.KeyConfig) {
	var options *KeyLightOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
		key.TmpString = options.Text
	}
}

func (handler *KeyLight) MountHandler(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	currentOption := key.Options.(*KeyLightOptions)
	switch currentOption.KeyLightAction {
	case "temperature+", "temperature-", "brightness+", "brightness-":
		ticker := time.NewTicker(defaultRefreshInterval)
		key.TickerQuit = make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:

					fmt.Println("hoho " + currentOption.KeyLightAction)
					keyLight, err := makeKeyLightRequest(*currentOption, http.MethodGet, nil)
					if err == nil {
						updateState(key, *keyLight, dev, keyIndex, page)
					}
				case <-key.TickerQuit:
					ticker.Stop()
					return
				}
			}
		}()
	}
}

func updateState(key *api.KeyConfig, keyLight keyLightResponse, dev *utils.VirtualDev, keyIndex int, page int) {
	currentOption := key.Options.(*KeyLightOptions)
	switch currentOption.KeyLightAction {
	case "temperature+", "temperature-":
		currentOption.Text = key.TmpString + "\n" + strconv.Itoa(keyLight.Lights[0].Temperature)
		key.CachedImage = nil
		key.CachedPressedImage = nil
		break
	case "brightness+", "brightness-":
		key.CachedImage = nil
		key.CachedPressedImage = nil
		currentOption.Text = key.TmpString + "\n" + strconv.Itoa(keyLight.Lights[0].Brightness)
		break
	}
	prepareImages(dev, key, keyIndex, page, key.Options.(DefaultOptions))
	setKeyImage(dev, key, keyIndex, page, key.Options.(DefaultOptions))
}

func (handler *KeyLight) UnmountHandler(key *api.KeyConfig) {
	if key.TickerQuit != nil {
		close(key.TickerQuit)
	}
}

func (handler *KeyLight) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	currentOption := key.Options.(*KeyLightOptions)
	keyLight, err := makeKeyLightRequest(*currentOption, http.MethodGet, nil)
	if err == nil {
		switch currentOption.KeyLightAction {
		case "toggle":
			if keyLight.Lights[0].On == 1 {
				keyLight.Lights[0].On = 0
			} else {
				keyLight.Lights[0].On = 1
			}
			break
		case "brightness+":
			keyLight.Lights[0].Brightness += brightnessStep
			if keyLight.Lights[0].Brightness > maxBrightness {
				keyLight.Lights[0].Brightness = maxBrightness
			}
			break
		case "brightness-":
			keyLight.Lights[0].Brightness -= brightnessStep
			if keyLight.Lights[0].Brightness < minBrightness {
				keyLight.Lights[0].Brightness = minBrightness
			}
			break
		case "temperature+":
			keyLight.Lights[0].Temperature += temperatureStep
			if keyLight.Lights[0].Temperature > maxTemperature {
				keyLight.Lights[0].Temperature = maxTemperature
			}
			break
		case "temperature-":
			keyLight.Lights[0].Temperature -= temperatureStep
			if keyLight.Lights[0].Temperature < minTemperature {
				keyLight.Lights[0].Temperature = minTemperature
			}
			break
		}
		requestBody, _ := json.Marshal(keyLight)
		keyLight, err = makeKeyLightRequest(*currentOption, http.MethodPut, bytes.NewReader(requestBody))
		if err == nil {
			updateState(key, *keyLight, dev, keyIndex, page)
		}
	}
}
