package _interface

import (
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

func UnmountHandlers(devs map[string]*utils.VirtualDev) {
	for s := range devs {
		dev := devs[s]
		UnmountDevHandlers(dev)
	}
}

func UnmountDevHandlers(dev *utils.VirtualDev) {
	for i := range dev.Config {
		unmountPageHandlers(dev, dev.Config[i])
	}
}

func unmountPageHandlers(dev *utils.VirtualDev, page api.Page) {
	for keyindex := 0; keyindex < len(page); keyindex++ {
		key := &page[keyindex]
		handler := dev.GetHandler(key)
		go func() {
			handler.UnmountHandler(key)
		}()
	}
}

func RenderPage(dev *utils.VirtualDev, page int) {
	if page != dev.Page {
		unmountPageHandlers(dev, dev.Config[dev.Page])
	}
	dev.Page = page
	currentPage := dev.Config[page]
	for i := 0; i < len(currentPage); i++ {
		currentKey := &currentPage[i]
		handler := dev.GetHandler(currentKey)
		handler.MountHandler(dev, currentKey, i, page)
		handler.RenderHandlerKey(dev, currentKey, i, page)
	}
	//EmitPage(dev, page)
}

func PrepareConfig(dev *utils.VirtualDev, page int) {
	currentPage := dev.Config[page]
	for i := 0; i < len(currentPage); i++ {
		currentKey := &currentPage[i]
		handler := dev.GetHandler(currentKey)
		handler.PrepareKey(dev, currentKey)
	}
}
