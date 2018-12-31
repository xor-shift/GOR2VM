package framebuffer

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	sfGraphics "gopkg.in/teh-cmc/go-sfml.v24/graphics"
	sfWindow "gopkg.in/teh-cmc/go-sfml.v24/window"

	peripheralmanager "GOR2VM/PeripheralManager"
)

//FrameBuffer is the framebuffer peripheral struct
type FrameBuffer struct {
	OnTick    func()
	OnRX      func(rx uint32)
	GetTX     func() uint32
	flushCall bool

	_videoMode  sfWindow.SfVideoMode
	_csettings  sfWindow.SfContextSettings
	_window     sfGraphics.SfWindow
	_mainbuffer sfGraphics.Struct_SS_sfImage
	_tempbuffer sfGraphics.Struct_SS_sfImage
	_fontSheet  sfGraphics.Struct_SS_sfImage

	window           *sdl.Window
	surface          *sdl.Surface
	renderer         *sdl.Renderer
	fontSheet        *sdl.Surface
	fontSheetTexture *sdl.Texture

	width, height uint

	registers []uint16 // LS(layer select)/MS(mode select), RX, RY, GPR0, GPR1, GPR2, GPR3, GPR4
	//if LS is 0 then mode is text, LS 1 through 16 selects layers to access
}

//NewFrameBuffer creates a new framebuffer object (This is not the peripheral)
func NewFrameBuffer(width, height uint, scale float64, title string) (fb *FrameBuffer) {
	fb = new(FrameBuffer)

	fb.width, fb.height = width, height

	fb.registers = make([]uint16, 8)
	fb.OnTick = fb.onTick

	/// SDL stuff

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(width), int32(height), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	fb.window = window

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	fb.surface = surface

	renderer, err := sdl.CreateRenderer(fb.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	fb.renderer = renderer

	fontSheet, err := img.Load("./data/font.png")
	if err != nil {
		panic(err)
	}
	fb.fontSheet = fontSheet

	fontSheetTexture, err := fb.renderer.CreateTextureFromSurface(fb.fontSheet)
	if err != nil {
		panic(err)
	}
	fb.fontSheetTexture = fontSheetTexture

	///

	fb.OnRX = func(rx uint32) {
		if rx&0x20000 == 0x20000 { //we are receiving data
			if fb.registers[0] == 0 {
				fb.processCharacter(uint16(rx & 0xFFFF))
			}
		} else {
			fb.flushCall = true
		}
	}

	fb.GetTX = func() uint32 {
		return 0
	}
	return
}

//NewPeripheral creates a framebuffer peripheral from a struct
func (fb *FrameBuffer) NewPeripheral() (p *peripheralmanager.Peripheral) {
	p = peripheralmanager.NewPeripheral(fb.OnTick, fb.OnRX, fb.GetTX)
	return
}

//DeInit deinitialises the framebuffer to free up memory
func (fb *FrameBuffer) DeInit() {
	fb.fontSheetTexture.Destroy()
	fb.fontSheet.Free()
	fb.renderer.Destroy()
	fb.window.Destroy()
	sdl.Quit()
}

func (fb *FrameBuffer) draw() {

}

func (fb *FrameBuffer) onTick() {
	width, height := fb.window.GetSize()
	fontwidth, fontheight := fb.fontSheet.W/16, fb.fontSheet.H/16
	maxColumns := width / (fontwidth + 1)
	maxRows := height / (fontheight + 1)

	fb.renderer.Clear()
	fb.renderer.SetDrawColor(0, 0, 0, 255)
	fb.renderer.FillRect(&sdl.Rect{0, 0, width, height})

	if fb.flushCall {
		fb.flushCall = false
	} else {
		return
	}

	if fb.registers[0] == 0 {
		for row := int32(0); row <= maxRows; row++ {
			for col := int32(0); col <= maxColumns; col++ {

			}
		}
	}

	fb.renderer.Present()
	fb.window.UpdateSurface()
}

func (fb *FrameBuffer) processCharacter(chr uint16) {
	width, height := fb.window.GetSize()
	fontwidth, fontheight := fb.fontSheet.W/16, fb.fontSheet.H/16
	maxColumns := width / (fontwidth + 1)
	maxRows := height / (fontheight + 1)

	switch chr & 0xF000 {
	case 0x0000:
		if int32(fb.registers[1]) > maxColumns {
			fb.registers[1] = 0
			fb.registers[2]++
		}

		if int32(fb.registers[2]) > maxRows {
			fb.registers[2]--
			for row := int32(1); row <= maxRows; row++ {
				for col := int32(0); col <= maxColumns; col++ {
					sfGraphics.SfImage_setPixel(fb._tempbuffer, uint(col), uint(row-1), sfGraphics.SfImage_getPixel(fb._tempbuffer, uint(col), uint(row)))
				}
			}
		}

		if chr&0xFF >= 0x20 || chr&0xFF <= 0x7E {
			charColorHolder := sfGraphics.NewSfColor()
			charColorHolder.SetR(uint8(chr & 0xFF))
			sfGraphics.SfImage_setPixel(fb._tempbuffer, uint(fb.registers[1]), uint(fb.registers[2]), charColorHolder)
			fb.registers[1]++
		} else {
			if chr == 0x0004 { //EOT character
				fb.registers[0] = 1
				return
			}
		}
	case 0x1000:
	case 0x2000:
	case 0x3000:
		fb.registers[1] = 0
		fb.registers[2]++
	}
}
