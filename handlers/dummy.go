package handlers

import (
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	"image"
	"image/draw"
)

type Dummy struct {
	Type string `json:"type"`
}

var DummyHandler = Dummy{
	Type: "",
}

func (handler *Dummy) GetType() string {
	return ""
}
func (handler *Dummy) PrepareKey(dev *utils.VirtualDev, key *api.KeyConfig) {
	if key.Options == nil {
		key.Options = &DefaultOptionsStruct{}
		img := image.NewRGBA(image.Rect(0, 0, int(dev.Deck.Pixels), int(dev.Deck.Pixels)))
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		key.CachedImage = img
	}
}

func (handler *Dummy) RenderPressedHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	setPressedKeyImage(dev, key, keyIndex, page, key.Options.(DefaultOptions))
}

func (handler *Dummy) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	prepareImages(dev, key, keyIndex, page, key.Options.(DefaultOptions))
	setKeyImage(dev, key, keyIndex, page, key.Options.(DefaultOptions))
}

func (handler *Dummy) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {}
