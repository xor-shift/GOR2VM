package framebuffer

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	peripheralmanager "GOR2VM/PeripheralManager"
)

//FrameBuffer is the framebuffer peripheral struct
type FrameBuffer struct {
	OnTick    func()
	OnRX      func(rx uint32)
	GetTX     func() uint32
	flushCall bool

	registers []uint16 // LS(layer select)/MS(mode select), RX, RY, GPR0, GPR1, GPR2, GPR3, GPR4
	//if LS is 0 then mode is text, LS 1 through 16 selects layers to access
}

//NewFrameBuffer creates a new framebuffer object (This is not the peripheral)
func NewFrameBuffer(width, height int32, scale float64, title string) (fb *FrameBuffer) {
	fb = new(FrameBuffer)

	fb.registers = make([]uint16, 8)
	fb.OnTick = fb.onTick

	rl.InitWindow(width, height, title)

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
}

func (fb *FrameBuffer) draw() {

}

func (fb *FrameBuffer) onTick() {
	width, height := int32(0), int32(0)         //fb.window.GetSize()
	fontwidth, fontheight := int32(0), int32(0) //fb.fontSheet.W/16, fb.fontSheet.H/16
	maxColumns := width / (fontwidth + 1)
	maxRows := height / (fontheight + 1)

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
}

func (fb *FrameBuffer) processCharacter(chr uint16) {
	width, height := int32(0), int32(0)         //fb.window.GetSize()
	fontwidth, fontheight := int32(0), int32(0) //fb.fontSheet.W/16, fb.fontSheet.H/16
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
					//sfGraphics.SfImage_setPixel(fb._tempbuffer, uint(col), uint(row-1), sfGraphics.SfImage_getPixel(fb._tempbuffer, uint(col), uint(row)))
				}
			}
		}

		if chr&0xFF >= 0x20 || chr&0xFF <= 0x7E {
			//charColorHolder := sfGraphics.NewSfColor()
			//charColorHolder.SetR(uint8(chr & 0xFF))
			//sfGraphics.SfImage_setPixel(fb._tempbuffer, uint(fb.registers[1]), uint(fb.registers[2]), charColorHolder)
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
