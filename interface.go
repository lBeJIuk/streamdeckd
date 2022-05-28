package main

import (
	"github.com/lBeJIuk/streamdeckd/handlers"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	_ "github.com/unix-streamdeck/driver"
)

func RenderPage(dev *utils.VirtualDev, page int) {
	if page != dev.Page {
		unmountPageHandlers(dev.Config[dev.Page])
	}
	dev.Page = page
	currentPage := dev.Config[page]
	for i := 0; i < len(currentPage); i++ {
		currentKey := &currentPage[i]
		go renderKey(dev, currentKey, i, page)
	}
	EmitPage(dev, page)
}

func renderKey(dev *utils.VirtualDev, currentKey *api.KeyConfig, keyIndex int, page int) {
	handler, err := handlers.GetHandler(currentKey)
	if err != nil {
		return
	}
	handler.RenderHandlerKey(dev, currentKey, keyIndex, page)
}
