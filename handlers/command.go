package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type Command struct {
	Type string `json:"type"`
}

var CommandHandler = Command{
	Type: "command",
}

type CommandOptions struct {
	DefaultOptionsStruct
	Command string `json:"command,omitempty"`
}

func (handler *Command) GetType() string {
	return handler.Type
}
func (handler *Command) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options CommandOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
	options = key.Options.(CommandOptions)
	setKeyImage(dev, key, keyIndex, page, &options)
	//var deckInfo api.StreamDeckInfo
	//for i := range sDInfo {
	//	if sDInfo[i].Serial == dev.Deck.Serial {
	//		deckInfo = sDInfo[i]
	//	}
	//}
	//if currentKey.Buff == nil {
	//	if currentKey.IconHandler == "" {
	//		SetKeyImage(dev, currentKey, i, page)
	//
	//	} else if currentKey.IconHandlerStruct == nil {
	//		var handler api.IconHandler
	//		modules := handlers.AvailableModules()
	//		for _, module := range modules {
	//			if module.Name == currentKey.IconHandler {
	//				handler = module.NewIcon()
	//			}
	//		}
	//		if handler == nil {
	//			return
	//		}
	//		log.Printf("Created & Started %s\n", currentKey.IconHandler)
	//		handler.Start(*currentKey, deckInfo, func(image image.Image) {
	//			if image.Bounds().Max.X != int(dev.Deck.Pixels) || image.Bounds().Max.Y != int(dev.Deck.Pixels) {
	//				image = api.ResizeImage(image, int(dev.Deck.Pixels))
	//			}
	//			SetImage(dev, image, i, page)
	//			currentKey.Buff = image
	//		})
	//		currentKey.IconHandlerStruct = handler
	//	}
	//} else {
	//	SetImage(dev, currentKey.Buff, i, page)
	//}
	//if currentKey.IconHandlerStruct != nil && !currentKey.IconHandlerStruct.IsRunning() {
	//	log.Printf("Started %s\n", currentKey.IconHandler)
	//	currentKey.IconHandlerStruct.Start(*currentKey, deckInfo, func(image image.Image) {
	//		if image.Bounds().Max.X != int(dev.Deck.Pixels) || image.Bounds().Max.Y != int(dev.Deck.Pixels) {
	//			image = api.ResizeImage(image, int(dev.Deck.Pixels))
	//		}
	//		SetImage(dev, image, i, page)
	//		currentKey.Buff = image
	//	})
	//}

}
func (handler *Command) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	var options CommandOptions
	options = key.Options.(CommandOptions)
	runCommand(options.Command)
}

func (commandOptions *CommandOptions) GetIcon() string {
	return commandOptions.Icon
}
func (commandOptions *CommandOptions) GetText() string {
	return commandOptions.Text
}
func (commandOptions *CommandOptions) GetTextSize() int {
	return commandOptions.TextSize
}
func (commandOptions *CommandOptions) GetTextAlignment() string {
	return commandOptions.TextAlignment
}
func (commandOptions *CommandOptions) GetBackgroundColor() string {
	return commandOptions.BackgroundColor
}
func (commandOptions *CommandOptions) GetTextColor() string {
	return commandOptions.TextColor
}
