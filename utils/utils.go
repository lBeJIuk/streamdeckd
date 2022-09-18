package utils

import (
	"bytes"
	"encoding/base64"
	"github.com/unix-streamdeck/api"
	streamdeck "github.com/unix-streamdeck/driver"
	"image"

	"strings"
)

type VirtualDev struct {
	Deck     streamdeck.Device
	Page     int
	Profile  string
	IsOpen   bool
	Config   []api.Page
	Handlers []Handler
}

type Handler interface {
	GetType() string
	RenderHandlerKey(dev *VirtualDev, key *api.KeyConfig, keyIndex int, page int)
	RenderPressedHandlerKey(dev *VirtualDev, key *api.KeyConfig, keyIndex int, page int)
	HandleInput(dev *VirtualDev, key *api.KeyConfig, keyIndex int, page int)
	PrepareKey(dev *VirtualDev, key *api.KeyConfig)
	MountHandler(dev *VirtualDev, key *api.KeyConfig, keyIndex int, page int)
	UnmountHandler(key *api.KeyConfig)
}

func (vDev *VirtualDev) GetHandler(key *api.KeyConfig) Handler {
	for _, handler := range vDev.Handlers {
		if handler.GetType() == key.Type {
			return handler
		}
	}
	return nil
}

func ParseIcon(base64Image string) (image.Image, error) {
	base64Image = strings.Split(base64Image, ",")[1]
	imgBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(imgBytes)
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}
