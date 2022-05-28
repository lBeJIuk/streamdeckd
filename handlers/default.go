package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/lBeJIuk/streamdeckd/utils"
	"github.com/unix-streamdeck/api"
	"golang.org/x/sync/semaphore"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

type DummyHandler struct{}

func (handler *DummyHandler) GetType() string {
	return ""
}
func (handler *DummyHandler) RenderHandlerKey(dev *utils.VirtualDev, key *api.KeyConfig, keyIndex int, page int) {
}
func (handler *DummyHandler) HandleInput(dev *utils.VirtualDev, key *api.KeyConfig, page int) {}

var sem = semaphore.NewWeighted(int64(1))

func loadImage(dev *utils.VirtualDev, path string) (image.Image, error) {
	path = strings.Split(path, ",")[1]
	imgBytes, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(imgBytes)
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return api.ResizeImage(img, int(dev.Deck.Pixels)), nil
}

func setImage(dev *utils.VirtualDev, img image.Image, i int, page int) {
	ctx := context.Background()
	err := sem.Acquire(ctx, 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer sem.Release(1)
	if dev.Page == page && dev.IsOpen {
		err := dev.Deck.SetImage(uint8(i), img)
		if err != nil {
			if strings.Contains(err.Error(), "hidapi") {
				//disconnect(dev)
			} else if strings.Contains(err.Error(), "dimensions") {
				log.Println(err)
			} else {
				log.Println(err)
			}
		}
	}
}

type DefaultOptionsStruct struct {
	Icon            string `json:"icon,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	PressedIcon     string `json:"pressedIcon,omitempty"`
	Text            string `json:"text,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	TextSize        int    `json:"textSize,omitempty"`
	TextAlignment   string `json:"textAlignment,omitempty"`
}

type DefaultOptions interface {
	GetIcon() string
	GetBackgroundColor() string
	GetText() string
	GetTextColor() string
	GetTextSize() int
	GetTextAlignment() string
}

func setKeyImage(dev *utils.VirtualDev, key *api.KeyConfig, i int, page int, options DefaultOptions) {
	if key.CachedImage == nil {
		icon := options.GetIcon()
		var img image.Image
		if icon != "" {
			var err error
			img, err = loadImage(dev, options.GetIcon())
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			simpleImg := image.NewRGBA(image.Rect(0, 0, int(dev.Deck.Pixels), int(dev.Deck.Pixels)))
			backgroundColor, err := parseHexColor(options.GetBackgroundColor())
			if err != nil {
				backgroundColor = color.RGBA{0, 0, 0, 0xff}
			}
			draw.Draw(simpleImg, simpleImg.Bounds(), image.NewUniform(backgroundColor), image.Point{}, draw.Src)
			img = simpleImg
		}
		if options.GetText() != "" {
			fontColor, err := parseHexColor(options.GetTextColor())
			if err != nil {
				fontColor = color.RGBA{255, 255, 255, 0xff}
			}
			imgWithText, err := api.DrawText(img, options.GetText(), options.GetTextSize(), options.GetTextAlignment(), fontColor)
			if err != nil {
				log.Println(err)
			} else {
				img = imgWithText
			}
		}
		key.CachedImage = img
	}
	if key.CachedImage != nil {
		setImage(dev, key.CachedImage, i, page)
	}
}
func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}

func runCommand(command string) {
	go func() {
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/nohup "+command)

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pgid:      0,
			Pdeathsig: syscall.SIGHUP,
		}
		if err := cmd.Start(); err != nil {
			fmt.Println("There was a problem running ", command, ":", err)
		} else {
			pid := cmd.Process.Pid
			cmd.Process.Release()
			fmt.Println(command, " has been started with pid", pid)
		}
	}()
}
