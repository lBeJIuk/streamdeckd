package handlers

import (
	"encoding/json"
	_interface "github.com/lBeJIuk/streamdeckd/interface"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type ChangePage struct {
	Type string `json:"type"`
}

var ChangePageHandler = ChangePage{
	Type: "changePage",
}

type ChangePageOptions struct {
	DefaultOptionsStruct
	Page int `json:"page,omitempty"`
}

func (handler *ChangePage) GetType() string {
	return handler.Type
}
func (handler *ChangePage) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options ChangePageOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
	options = key.Options.(ChangePageOptions)
	setKeyImage(dev, key, keyIndex, page, &options)
}
func (handler *ChangePage) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	var options ChangePageOptions
	options = key.Options.(ChangePageOptions)
	_interface.RenderPage(dev, options.Page-1)
}
