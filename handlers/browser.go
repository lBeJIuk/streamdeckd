package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type Browser struct {
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
func (handler *Browser) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options BrowserOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
	options = key.Options.(BrowserOptions)
	setKeyImage(dev, key, keyIndex, page, &options)
}
func (handler *Browser) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	runCommand("xdg-open " + key.Options.(BrowserOptions).Url)
}

func (browserOptions *BrowserOptions) GetIcon() string {
	return browserOptions.Icon
}
func (browserOptions *BrowserOptions) GetText() string {
	return browserOptions.Text
}
func (browserOptions *BrowserOptions) GetTextSize() int {
	return browserOptions.TextSize
}
func (browserOptions *BrowserOptions) GetTextAlignment() string {
	return browserOptions.TextAlignment
}
func (browserOptions *BrowserOptions) GetBackgroundColor() string {
	return browserOptions.BackgroundColor
}
func (browserOptions *BrowserOptions) GetTextColor() string {
	return browserOptions.TextColor
}
