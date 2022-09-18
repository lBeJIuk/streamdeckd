package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	"log"
)

type Brightness struct {
	Dummy
	Type string `json:"type"`
}

var BrightnessHandler = Brightness{
	Type: "brightness",
}

type BrightnessOptions struct {
	DefaultOptionsStruct
	Brightness int `json:"brightness,omitempty"`
}

func (handler *Brightness) GetType() string {
	return handler.Type
}
func (handler *Brightness) PrepareKey(dev *utils.VirtualDev, key *api.KeyConfig) {
	var options *BrightnessOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
}
func (handler *Brightness) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options *BrightnessOptions
	options = key.Options.(*BrightnessOptions)
	err := dev.Deck.SetBrightness(uint8(options.Brightness))
	if err != nil {
		log.Println(err)
	}
}
