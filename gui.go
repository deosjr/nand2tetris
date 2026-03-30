//go:build gui

package main

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func (s *Screen512x256) RunScreen(win *pixelgl.Window) {
	white := color.RGBA{255, 255, 255, 0}
	black := color.RGBA{0, 0, 0, 0}
	pd := pixel.MakePictureData(win.Bounds())
	for row := 0; row < 256; row++ {
		for w := 0; w < 32; w++ {
			addr := row*32 + w
			word := s.ram.mem[addr]
			for c := 0; c < 16; c++ {
				b := nthBit(word, uint16(15-c))
				invrow := (255 - row)
				if b {
					pd.Pix[invrow*512+16*w+c] = white
				} else {
					pd.Pix[invrow*512+16*w+c] = black
				}
			}
		}
	}
	sprite := pixel.NewSprite(pd, pd.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
}

// TODO: separate textmode screen that only writes characters?
func (s *Screen256x512) RunScreen(win *pixelgl.Window) {
	white := color.RGBA{255, 255, 255, 0}
	pd := pixel.MakePictureData(win.Bounds())
	for row := 0; row < 512; row++ {
		for r := 0; r < 16; r++ {
			addr := row*16 + r
			word := s.ram.mem[addr]
			for c := 0; c < 16; c++ {
				b := nthBit(word, uint16(15-c))
				if b {
					invrow := (511 - row)
					pd.Pix[invrow*256+16*r+c] = white
				}
			}
		}
	}
	sprite := pixel.NewSprite(pd, pd.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
}

func upperIfShift(upper uint16, shift bool) uint16 {
	if shift {
		return upper
	}
	return upper + 32
}

func (k *SimpleKeyboard) RunKeyboard(win *pixelgl.Window) {
	shift := win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyRightShift)
	switch {
	case win.Pressed(pixelgl.KeyComma):
		if shift {
			k.reg.SendIn(0x3C)
		} else {
			k.reg.SendIn(0x2C)
		}
	case win.Pressed(pixelgl.KeyPeriod):
		if shift {
			k.reg.SendIn(0x3E)
		} else {
			k.reg.SendIn(0x2E)
		}
	case win.Pressed(pixelgl.Key0):
		if shift {
			k.reg.SendIn(0x29)
		} else {
			k.reg.SendIn(0x30)
		}
	case win.Pressed(pixelgl.Key1):
		if shift {
			k.reg.SendIn(0x21)
		} else {
			k.reg.SendIn(0x31)
		}
	case win.Pressed(pixelgl.Key2):
		if shift {
			k.reg.SendIn(0x40)
		} else {
			k.reg.SendIn(0x32)
		}
	case win.Pressed(pixelgl.Key3):
		k.reg.SendIn(0x33)
	case win.Pressed(pixelgl.Key4):
		k.reg.SendIn(0x34)
	case win.Pressed(pixelgl.Key5):
		k.reg.SendIn(0x35)
	case win.Pressed(pixelgl.Key6):
		k.reg.SendIn(0x36)
	case win.Pressed(pixelgl.Key7):
		k.reg.SendIn(0x37)
	case win.Pressed(pixelgl.Key8):
		if shift {
			k.reg.SendIn(0x2A)
		} else {
			k.reg.SendIn(0x38)
		}
	case win.Pressed(pixelgl.Key9):
		if shift {
			k.reg.SendIn(0x28)
		} else {
			k.reg.SendIn(0x39)
		}
	case win.Pressed(pixelgl.KeyMinus):
		k.reg.SendIn(0x2D)
	case win.Pressed(pixelgl.KeySemicolon):
		k.reg.SendIn(0x3B)
	case win.Pressed(pixelgl.KeyEqual):
		if shift {
			k.reg.SendIn(0x2B)
		} else {
			k.reg.SendIn(0x3D)
		}
	case win.Pressed(pixelgl.KeyA):
		k.reg.SendIn(upperIfShift(0x41, shift))
	case win.Pressed(pixelgl.KeyB):
		k.reg.SendIn(upperIfShift(0x42, shift))
	case win.Pressed(pixelgl.KeyC):
		k.reg.SendIn(upperIfShift(0x43, shift))
	case win.Pressed(pixelgl.KeyD):
		k.reg.SendIn(upperIfShift(0x44, shift))
	case win.Pressed(pixelgl.KeyE):
		k.reg.SendIn(upperIfShift(0x45, shift))
	case win.Pressed(pixelgl.KeyF):
		k.reg.SendIn(upperIfShift(0x46, shift))
	case win.Pressed(pixelgl.KeyG):
		k.reg.SendIn(upperIfShift(0x47, shift))
	case win.Pressed(pixelgl.KeyH):
		k.reg.SendIn(upperIfShift(0x48, shift))
	case win.Pressed(pixelgl.KeyI):
		k.reg.SendIn(upperIfShift(0x49, shift))
	case win.Pressed(pixelgl.KeyJ):
		k.reg.SendIn(upperIfShift(0x4A, shift))
	case win.Pressed(pixelgl.KeyK):
		k.reg.SendIn(upperIfShift(0x4B, shift))
	case win.Pressed(pixelgl.KeyL):
		k.reg.SendIn(upperIfShift(0x4C, shift))
	case win.Pressed(pixelgl.KeyM):
		k.reg.SendIn(upperIfShift(0x4D, shift))
	case win.Pressed(pixelgl.KeyN):
		k.reg.SendIn(upperIfShift(0x4E, shift))
	case win.Pressed(pixelgl.KeyO):
		k.reg.SendIn(upperIfShift(0x4F, shift))
	case win.Pressed(pixelgl.KeyP):
		k.reg.SendIn(upperIfShift(0x50, shift))
	case win.Pressed(pixelgl.KeyQ):
		k.reg.SendIn(upperIfShift(0x51, shift))
	case win.Pressed(pixelgl.KeyR):
		k.reg.SendIn(upperIfShift(0x52, shift))
	case win.Pressed(pixelgl.KeyS):
		k.reg.SendIn(upperIfShift(0x53, shift))
	case win.Pressed(pixelgl.KeyT):
		k.reg.SendIn(upperIfShift(0x54, shift))
	case win.Pressed(pixelgl.KeyU):
		k.reg.SendIn(upperIfShift(0x55, shift))
	case win.Pressed(pixelgl.KeyV):
		k.reg.SendIn(upperIfShift(0x56, shift))
	case win.Pressed(pixelgl.KeyW):
		k.reg.SendIn(upperIfShift(0x57, shift))
	case win.Pressed(pixelgl.KeyX):
		k.reg.SendIn(upperIfShift(0x58, shift))
	case win.Pressed(pixelgl.KeyY):
		k.reg.SendIn(upperIfShift(0x59, shift))
	case win.Pressed(pixelgl.KeyZ):
		k.reg.SendIn(upperIfShift(0x5A, shift))
	case win.Pressed(pixelgl.KeySpace):
		k.reg.SendIn(0x20)
	case win.Pressed(pixelgl.KeyEnter):
		k.reg.SendIn(0x80)
	case win.Pressed(pixelgl.KeyBackspace):
		k.reg.SendIn(0x08)
	default:
		k.reg.SendIn(0)
	}
}

func startComputer(computer *Computer, debugger Debugger) {
	go run(computer, debugger)
	pixelgl.Run(runPeripherals(computer))
}

type guiScreen interface {
	RunScreen(win *pixelgl.Window)
}

type guiKeyboard interface {
	RunKeyboard(win *pixelgl.Window)
}

func runPeripherals(computer *Computer) func() {
	return func() {
		cfg := pixelgl.WindowConfig{
			Title:  "nand2tetris",
			Bounds: pixel.R(0, 0, 512, 256),
			VSync:  true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			panic(err)
		}

		screen := computer.data_mem.screen.(guiScreen)
		keyboard := computer.data_mem.keyboard.(guiKeyboard)
		for !win.Closed() {
			keyboard.RunKeyboard(win)
			screen.RunScreen(win)
			win.Update()
		}
	}
}
