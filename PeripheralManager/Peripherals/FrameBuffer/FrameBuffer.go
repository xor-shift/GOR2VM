package framebuffer

import (
	sfGraphics "gopkg.in/teh-cmc/go-sfml.v24/graphics"
	sfSystem "gopkg.in/teh-cmc/go-sfml.v24/system"
	sfWindow "gopkg.in/teh-cmc/go-sfml.v24/window"

	peripheralmanager "GOR2VM/PeripheralManager"
)

//FrameBuffer is the framebuffer peripheral struct
type FrameBuffer struct {
	OnTick    func()
	OnRX      func(rx uint32)
	GetTX     func() uint32
	flushCall bool

	videoMode  sfWindow.SfVideoMode
	csettings  sfWindow.SfContextSettings
	window     sfGraphics.SfWindow
	mainbuffer sfGraphics.Struct_SS_sfImage
	tempbuffer sfGraphics.Struct_SS_sfImage
	fontSheet  sfGraphics.Struct_SS_sfImage

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

	fb.videoMode = sfWindow.NewSfVideoMode()
	fb.videoMode.SetWidth(width)
	fb.videoMode.SetHeight(height)
	fb.videoMode.SetBitsPerPixel(32)
	fb.csettings = sfWindow.NewSfContextSettings()
	fb.window = sfGraphics.SfRenderWindow_create(fb.videoMode, title, uint(0), fb.csettings)
	if sfWindow.SfWindow_isOpen(fb.window) <= 0 {
		fb = nil
		return
	}
	sfGraphics.SfRenderWindow_clear(fb.window, sfGraphics.GetSfBlack())
	sfGraphics.SfRenderWindow_display(fb.window)
	fb.mainbuffer = sfGraphics.SfImage_create(width, height)
	fb.tempbuffer = sfGraphics.SfImage_create(width, height)

	fb.fontSheet = sfGraphics.SfImage_createFromFile("./data/font.png")
	if fb.fontSheet == nil {
		panic("h")
	}

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
	sfWindow.DeleteSfVideoMode(fb.videoMode)
	sfWindow.DeleteSfContextSettings(fb.csettings)
	sfWindow.SfWindow_destroy(fb.window)
}

func (fb *FrameBuffer) draw() {

}

func (fb *FrameBuffer) onTick() {
	maxColumns := fb.width / (((sfSystem.SwigcptrSfVector2u)(fb.fontSheet.Swigcptr())).GetX() + 16) / 16
	maxRows := fb.height / (((sfSystem.SwigcptrSfVector2u)(fb.fontSheet.Swigcptr())).GetY() + 16) / 16

	sfGraphics.SfRenderWindow_clear(fb.window, sfGraphics.GetSfBlack())

	if fb.flushCall {
		fb.mainbuffer = fb.tempbuffer
		fb.flushCall = false
	} else {
		return
	}

	if fb.registers[0] == 0 {
		for row := uint(0); row <= maxRows; row++ {
			for col := uint(0); col <= maxColumns; col++ {
				pix := sfGraphics.SfImage_getPixel(fb.tempbuffer, uint(col), uint(row))
				val := pix.GetR()

				trow := (val & 0xF0) >> 4
				tcol := val & 0xF
				posx := tcol * 17
				posy := trow * 17

				intr := sfGraphics.NewSfIntRect()
				intr.SetLeft(int(posx))
				intr.SetTop(int(posy))
				intr.SetWidth(16)
				intr.SetWidth(16)

				tex := sfGraphics.SfTexture_createFromImage(fb.fontSheet, intr)
				spr := sfGraphics.SfSprite_create()
				sfGraphics.SfSprite_setTexture(spr, tex, 0)

				pos := sfSystem.NewSfVector2f()
				pos.SetX(float32(posx))
				pos.SetY(float32(posy))

				sfGraphics.SfSprite_setPosition(spr, pos)
			}
		}
	}
	sfGraphics.SfRenderWindow_display(fb.window)
}

func (fb *FrameBuffer) processCharacter(chr uint16) {
	maxColumns := fb.width / (((sfSystem.SwigcptrSfVector2u)(fb.fontSheet.Swigcptr())).GetX() + 16) / 16 //thank you go and go-sfml, for this abomination
	maxRows := fb.height / (((sfSystem.SwigcptrSfVector2u)(fb.fontSheet.Swigcptr())).GetY() + 16) / 16   //I spent like 10 minutes trying to figure out how to make a vector a proper vector and not some bullshit sfGraphics vector
	//fb.width and height were actually workarounds before I realised that I needed to figure this out

	switch chr & 0xF000 {
	case 0x0000:
		if uint(fb.registers[1]) > maxColumns {
			fb.registers[1] = 0
			fb.registers[2]++
		}

		if uint(fb.registers[2]) > maxRows {
			fb.registers[2]--
			for row := uint(1); row <= maxRows; row++ {
				for col := uint(0); col <= maxColumns; col++ {
					sfGraphics.SfImage_setPixel(fb.tempbuffer, uint(col), uint(row-1), sfGraphics.SfImage_getPixel(fb.tempbuffer, uint(col), uint(row)))
				}
			}
		}

		if chr&0xFF >= 0x20 || chr&0xFF <= 0x7E {
			charColorHolder := sfGraphics.NewSfColor()
			charColorHolder.SetR(uint8(chr & 0xFF))
			sfGraphics.SfImage_setPixel(fb.tempbuffer, uint(fb.registers[1]), uint(fb.registers[2]), charColorHolder)
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
