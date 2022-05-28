package utils

import (
	"github.com/unix-streamdeck/api"
	streamdeck "github.com/unix-streamdeck/driver"
)

type VirtualDev struct {
	Deck    streamdeck.Device
	Page    int
	Profile string
	IsOpen  bool
	Config  []api.Page
}
