package handlers

import (
	"context"
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

var sem = semaphore.NewWeighted(int64(1))

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
	SetIcon(icon string)
	GetBackgroundColor() string
	GetText() string
	GetTextColor() string
	GetTextSize() int
	GetTextAlignment() string
}

func (defaultOptions *DefaultOptionsStruct) GetIcon() string {
	return defaultOptions.Icon
}
func (defaultOptions *DefaultOptionsStruct) SetIcon(icon string) {
	defaultOptions.Icon = icon
}
func (defaultOptions *DefaultOptionsStruct) GetText() string {
	return defaultOptions.Text
}
func (defaultOptions *DefaultOptionsStruct) GetTextSize() int {
	return defaultOptions.TextSize
}
func (defaultOptions *DefaultOptionsStruct) GetTextAlignment() string {
	return defaultOptions.TextAlignment
}
func (defaultOptions *DefaultOptionsStruct) GetBackgroundColor() string {
	return defaultOptions.BackgroundColor
}
func (defaultOptions *DefaultOptionsStruct) GetTextColor() string {
	return defaultOptions.TextColor
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

func prepareImages(dev *utils.VirtualDev, key *api.KeyConfig, i int, page int, options DefaultOptions) {
	if key.CachedImage == nil {
		img, err := createImg(options, int(dev.Deck.Pixels))
		if err != nil {
			log.Println(err)
			return
		}
		key.CachedImage = img
	}
	if key.CachedPressedImage == nil {
		img, err := createImg(options, int(dev.Deck.Pixels))
		if err != nil {
			log.Println(err)
			return
		}
		border := 7
		resizedImg := api.ResizeImage(img, int(dev.Deck.Pixels)-border)
		if tmpImg, ok := img.(*image.RGBA); ok {
			for x := 0; x < int(dev.Deck.Pixels); x++ {
				for y := 0; y < int(dev.Deck.Pixels); y++ {
					if (x < border || x > int(dev.Deck.Pixels)-border) || (y < border || y > int(dev.Deck.Pixels)-border) {
						// Draw borders
						tmpImg.Set(x, y, image.Black)
					} else {
						// Draw downscaled image
						tmpImg.Set(x, y, resizedImg.At(x-border, y-border))
					}
				}
			}
		}
		key.CachedPressedImage = img
	}
}

func createImg(options DefaultOptions, size int) (image.Image, error) {
	icon := options.GetIcon()
	var img image.Image
	if icon != "" {
		var err error
		img, err = utils.ParseIcon(options.GetIcon())
		if err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		simpleImg := image.NewRGBA(image.Rect(0, 0, size, size))
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
	return img, nil
}
func setKeyImage(dev *utils.VirtualDev, key *api.KeyConfig, i int, page int, options DefaultOptions) {
	if key.CachedImage != nil {
		setImage(dev, key.CachedImage, i, page)
	}
}
func setPressedKeyImage(dev *utils.VirtualDev, key *api.KeyConfig, i int, page int, options DefaultOptions) {
	if key.CachedPressedImage != nil {
		setImage(dev, key.CachedPressedImage, i, page)
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
