package utils

import (
	"errors"
	"github.com/unix-streamdeck/api"
	streamdeck "github.com/unix-streamdeck/driver"
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
	HandleInput(dev *VirtualDev, key *api.KeyConfig, page int)
}
type DummyHandler struct{}

func (handler *DummyHandler) GetType() string {
	return ""
}
func (handler *DummyHandler) RenderHandlerKey(dev *VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
}
func (handler *DummyHandler) HandleInput(dev *VirtualDev, key *api.KeyConfig, page int) {}

func (vDev *VirtualDev) GetHandler(key *api.KeyConfig) (Handler, error) {
	for _, handler := range vDev.Handlers {
		if handler.GetType() == key.Type {
			return handler, nil
		}
	}
	return &DummyHandler{}, errors.New("No handler found.")
}
