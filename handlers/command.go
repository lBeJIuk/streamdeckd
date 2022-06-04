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

func (c *CommandOptions) GetIcon() string {
	return c.DefaultOptionsStruct.GetIcon()
}

func (handler *Command) GetType() string {
	return handler.Type
}
func (handler *Command) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
	var options *CommandOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	} else {
		options = key.Options.(*CommandOptions)
	}
	setKeyImage(dev, key, keyIndex, page, options)
}

func (handler *Command) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	var options *CommandOptions
	options = key.Options.(*CommandOptions)
	runCommand(options.Command)
}
