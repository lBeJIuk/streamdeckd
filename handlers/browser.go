package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type Browser struct {
	Dummy
	Type string `json:"type"`
}

var BrowserHandler = Browser{
	Type: "browser",
}

type BrowserOptions struct {
	DefaultOptionsStruct
	Url string `json:"Url,omitempty"`
}

func (handler *Browser) GetType() string {
	return handler.Type
}

func (handler *Browser) PrepareKey(dev *utils.VirtualDev, key *api.KeyConfig) {
	var options *BrowserOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
}
func (handler *Browser) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	runCommand("xdg-open " + key.Options.(*BrowserOptions).Url)
}
