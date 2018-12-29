package DumbConsole

import (
	peripheralmanager "GOR2VM/PeripheralManager"
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

type DumbConsole struct {
	OnTick  func()
	OnRX    func(rx uint32)
	GetTX   func() uint32
	outView *gtk.TextView
}

func NewDumbConsole(view *gtk.TextView) (dc *DumbConsole) {
	dc = new(DumbConsole)

	dc.outView = view
	dc.OnTick = func() {
		//
	}
	dc.OnRX = func(rx uint32) {
		if rx&0x20000 == 0x20000 { //we are receiving data
			buf, _ := dc.outView.GetBuffer()
			buf.InsertAtCursor(fmt.Sprintf("%c", rx&0xFF))
		} else {
			//fmt.Println("Received bump")
		}
	}
	dc.GetTX = func() uint32 {
		return 0
	}
	return
}

func (dc *DumbConsole) NewPeripheral() (p *peripheralmanager.Peripheral) {
	p = peripheralmanager.NewPeripheral(dc.OnTick, dc.OnRX, dc.GetTX)
	return
}
