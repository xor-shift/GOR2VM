package PeripheralManager

import (
	"fmt"
)

type PManager struct {
	peripherals []*Peripheral
	arequest    uint16
}

type Peripheral struct {
	onTick func()
	onRX   func(rx uint32)
	getTX  func() uint32
	active bool
}

func NewPManager() (pm *PManager) {
	pm = new(PManager)
	pm.peripherals = make([]*Peripheral, 256)

	return
}

func NewPeripheral(onTick func(), onRX func(rx uint32), getTX func() uint32) (p *Peripheral) {
	p = new(Peripheral)
	p.onTick = onTick
	p.onRX = onRX
	p.getTX = getTX
	p.active = true
	return
}

func ExamplePeripheral() *Peripheral {
	p := new(Peripheral)
	p.onTick = func() {
		fmt.Println("Tick!")
	}
	p.onRX = func(rx uint32) {
		if rx&0x20000 == 0x20000 { //we are receiving data
			fmt.Println("Received data:", (rx & 0xFFFF))
		} else {
			fmt.Println("Received bump")
		}
	}
	p.getTX = func() uint32 {
		return 1
	}
	p.active = true
	return p
}

func (pm *PManager) MovePeripheral(dst uint8, src uint8) {
	pm.peripherals[dst] = pm.peripherals[src]
	pm.peripherals[dst].active = false
}

func (pm *PManager) SwapPeripherals(p1 uint8, p2 uint8) {
	temp := pm.peripherals[p1]
	pm.peripherals[p1] = pm.peripherals[p2]
	pm.peripherals[p2] = temp
}

func (pm *PManager) RegisterPeripheral(id uint8, p *Peripheral) {
	pm.peripherals[id] = p
}

func (pm *PManager) ActivatePeripheral(id uint8) {
	pm.peripherals[id].active = true
}

func (pm *PManager) DeactivatePeripheral(id uint8) {
	pm.peripherals[id].active = false
}

func (pm *PManager) TickPeripherals() {
	var arequest uint16
	arequest = 0xFFFF
	for i := 0; i < len(pm.peripherals); i++ {
		if pm.peripherals[i] == nil {
			continue
		}
		if pm.peripherals[i].active {
			pm.peripherals[i].onTick()
			if pm.peripherals[i].getTX()&0x10000 == 0x10000 {
				arequest = (uint16)(i)
			}
		}
	}
	pm.arequest = arequest
}

func (pm *PManager) TXToPort(id uint8, data uint32) { //it's the peripheral's job to parse the attention/data bit
	pm.peripherals[id].onRX(data)
}

func (pm *PManager) GetTXDataOfPort(id uint8) (tx uint16, b bool) { //not giving attention request as the instruction needs the highest port number
	tx = 0
	b = false

	if !pm.peripherals[id].active {
		return
	}

	temp := pm.peripherals[id].getTX()
	if temp&0x20000 == 0x20000 {
		tx = (uint16)(temp & 0xFFFF)
		b = true
	} else if temp&0x10000 == 0x10000 {
		return
	} else {
		return
	}
	return
}

func (pm *PManager) GetARequest() uint16 {
	return pm.arequest
}
