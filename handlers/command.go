package handlers

import (
	"encoding/json"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
)

type Command struct {
	Dummy
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

func (handler *Command) PrepareKey(dev *utils.VirtualDev, key *api.KeyConfig) {
	var options *CommandOptions
	if key.Options == nil {
		err := json.Unmarshal(key.RawOptions, &options)
		if err != nil {
			return
		}
		key.Options = options
	}
}

func (handler *Command) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {
	var options *CommandOptions
	options = key.Options.(*CommandOptions)
	runCommand(options.Command)
}
