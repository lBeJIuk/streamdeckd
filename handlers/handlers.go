package handlers

import (
	"errors"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type Handler interface {
	GetType() string
	RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int)
	HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int)
}

var handlers []Handler

func InitHandlers() {
	handlers = []Handler{
		&CommandHandler,
		&BrowserHandler,
	}
}

func GetHandler(key *api.KeyConfig) (Handler, error) {
	for _, handler := range handlers {
		if handler.GetType() == key.Type {
			return handler, nil
		}
	}
	return &DummyHandler{}, errors.New("No handler found.")
}
