package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	"log"
)

type Brightness struct {
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
func (handler *Brightness) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options *BrightnessOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	} else {
		options = key.Options.(*BrightnessOptions)
	}
	setKeyImage(dev, key, keyIndex, page, options)
}
func (handler *Brightness) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	var options *BrightnessOptions
	options = key.Options.(*BrightnessOptions)
	err := dev.Deck.SetBrightness(uint8(options.Brightness))
	if err != nil {
		log.Println(err)
	}
}
