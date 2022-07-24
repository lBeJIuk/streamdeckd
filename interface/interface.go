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
		unmountPageHandlers(dev.Config[i])
	}
}

func unmountPageHandlers(page api.Page) {
	//for i2 := 0; i2 < len(page); i2++ {
	//	key := &page[i2]
	//	if key.IconHandlerStruct != nil {
	//		log.Printf("Stopping %s\n", key.IconHandler)
	//		if key.IconHandlerStruct.IsRunning() {
	//			go func() {
	//				key.IconHandlerStruct.Stop()
	//				log.Printf("Stopped %s\n", key.IconHandler)
	//			}()
	//		}
	//	}
	//}
}

func RenderPage(dev *utils.VirtualDev, page int) {
	if page != dev.Page {
		unmountPageHandlers(dev.Config[dev.Page])
	}
	dev.Page = page
	currentPage := dev.Config[page]
	for i := 0; i < len(currentPage); i++ {
		currentKey := &currentPage[i]
		RenderKey(dev, currentKey, i, page)
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

func RenderKey(dev *utils.VirtualDev, currentKey *api.KeyConfig, keyIndex int, page int) {
	handler := dev.GetHandler(currentKey)
	handler.RenderHandlerKey(dev, currentKey, keyIndex, page)
}
