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
func (handler *Dummy) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	img := image.NewRGBA(image.Rect(0, 0, int(dev.Deck.Pixels), int(dev.Deck.Pixels)))
	draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
	setImage(dev, img, keyIndex, page)
}

func (handler *Dummy) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {}
