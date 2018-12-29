package peripheralmanager

import (
	"fmt"
)

//PManager stores 256 peripherals and an AR value
type PManager struct {
	peripherals []*Peripheral
	arequest    uint16
}

//Peripheral stores functions to a peripheral
type Peripheral struct {
	onTick func()
	onRX   func(rx uint32)
	getTX  func() uint32
	active bool
}

//NewPManager creates a new peripheral manager
//A peripheral manager stores 256 peripherals which can be accessed with a uint8
func NewPManager() (pm *PManager) {
	pm = new(PManager)
	pm.peripherals = make([]*Peripheral, 256)

	return
}

//NewPeripheral creates a new peripheral.
//onTick is executed on every CPU cycle.
//onRX is executed when a data is being sent to the peripheral.
//getTx is supposed to return whatever the peripheral wants to transmit when called.
//-> 0x00000 can be returned for no data.
//-> 0x2**** can be returned for a 16 bit data transmission.
//-> 0x10000 can be returned for an attention reply.
func NewPeripheral(onTick func(), onRX func(rx uint32), getTX func() uint32) (p *Peripheral) {
	p = new(Peripheral)
	p.onTick = onTick
	p.onRX = onRX
	p.getTX = getTX
	p.active = true
	return
}

//ExamplePeripheral returns an example peripheral.
//The example peripheral prints out what it receives (useful for peripheral interface debugging).
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

//MovePeripheral moves the peripheral at src to dst.
func (pm *PManager) MovePeripheral(dst uint8, src uint8) {
	pm.peripherals[dst] = pm.peripherals[src]
	pm.peripherals[dst].active = false
}

//SwapPeripherals swaps 2 peripherals.
func (pm *PManager) SwapPeripherals(p1 uint8, p2 uint8) {
	temp := pm.peripherals[p1]
	pm.peripherals[p1] = pm.peripherals[p2]
	pm.peripherals[p2] = temp
}

//RegisterPeripheral registers a new peripheral at the given port id.
func (pm *PManager) RegisterPeripheral(id uint8, p *Peripheral) {
	pm.peripherals[id] = p
}

//ActivatePeripheral activates a peripheral.
func (pm *PManager) ActivatePeripheral(id uint8) {
	pm.peripherals[id].active = true
}

//DeactivatePeripheral deactivates a peripheral.
func (pm *PManager) DeactivatePeripheral(id uint8) {
	pm.peripherals[id].active = false
}

//TickPeripherals ticks all active and non nil peripherals.
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

//TXToPort transmits a piece of 32 bit data to the port given.
//0x2xxxx is regular data transmission
//0x10000 is an attention request
func (pm *PManager) TXToPort(id uint8, data uint32) { //it's the peripheral's job to parse the attention/data bit
	pm.peripherals[id].onRX(data)
}

//GetTXDataOfPort gets transmission from a given port.
//If no data bit is set at the port, a value of 0 and a false is returned.
//If the data bit is set at the port, the data and a true is returned.
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

//GetARequest returns the port id of the peripheral with the highest port number which has sent an AR on the last tick
//If no peripheral has sent an AR, 0xFFFF is returned
func (pm *PManager) GetARequest() uint16 {
	return pm.arequest
}
