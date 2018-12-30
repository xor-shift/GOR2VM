package framebuffer

import (
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

	videoMode  sfWindow.SfVideoMode
	csettings  sfWindow.SfContextSettings
	window     sfGraphics.SfWindow
	mainbuffer sfGraphics.Struct_SS_sfImage
	tempbuffer sfGraphics.Struct_SS_sfImage

	registers []uint16 // LS(layer select)/MS(mode select), XR, YR, GPR0, GPR1, GPR2, GPR3, GPR4
	//if LS is 0 then mode is text, LS 1 through 16 selects layers to access
}

//NewFrameBuffer creates a new framebuffer object (This is not the peripheral)
func NewFrameBuffer(width, height uint, scale float64, title string) (fb *FrameBuffer) {
	fb = new(FrameBuffer)

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

	fb.OnRX = func(rx uint32) {
		if rx&0x20000 == 0x20000 { //we are receiving data
			if fb.registers[0] == 0 {
				if rx&0xFFFF == 0x0004 { //EOT character
					fb.registers[0] = 1
					return
				}
				//display text:

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
	sfGraphics.SfRenderWindow_clear(fb.window, sfGraphics.GetSfBlack())

	if fb.flushCall {
		fb.mainbuffer = fb.tempbuffer
		fb.flushCall = false
	} else {
		return
	}

	sfGraphics.SfRenderWindow_display(fb.window)
}
