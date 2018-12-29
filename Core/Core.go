package Core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	peripheralmanager "GOR2VM/PeripheralManager"
)

type Core struct {
	State *State
	PMngr *peripheralmanager.PManager

	MUL uint32 //Multiplier

	//to be obsoleted:
	//TX    uint32 //0x2**** -> data transfer
	//0x1**** -> attention request
	//RX    uint32
	//PS    uint8  //Port Select
}

type State struct {
	Memory    []uint32
	ROM       []uint32
	Regfile   []uint16
	WMask     uint16
	Running   bool
	Icount    uint64
	ChainedOp uint16
	ROP1      uint16 //read (as in past tense) operand 1
	ROP2      uint16 //^                       ^       2

	Sign     bool
	Overflow bool
	Zero     bool
	Carry    bool
}

func NewCore() (core *Core) {
	core = new(Core)
	core.State = new(State)
	core.PMngr = peripheralmanager.NewPManager()

	core.MUL = 60

	core.State.Memory = make([]uint32, 65536)
	core.State.ROM = make([]uint32, 65536)
	core.State.Regfile = make([]uint16, 16)
	core.State.WMask = 0
	core.State.Running = false
	core.State.Icount = 0

	return
}

func (core *Core) LoadFromBinary(path string) (err error) {
	err = nil

	file, fileErr := os.Open(path)
	if fileErr != nil {
		err = fileErr
		return
	}
	defer file.Close()

	tempMemory := make([]uint32, 65536)

	readErr := binary.Read(file, binary.LittleEndian, tempMemory)
	if readErr != nil {
		err = readErr
		return
	}
	if len(tempMemory) != 65536 {
		err = errors.New("File size invalid.") //how?
		return
	}

	copy(core.State.ROM, tempMemory)

	core.State.Running = false

	return
}

func (core *Core) Reset() {
	core.State.Regfile = make([]uint16, 16)
	core.State.Memory = make([]uint32, 65536)
	copy(core.State.Memory, core.State.ROM)
	core.State.WMask = 0
	core.State.Icount = 0
	core.State.Running = false

	core.State.Carry = false
	core.State.Sign = false
	core.State.Overflow = false
	core.State.Zero = false

	core.State.ChainedOp = 0
	core.State.ROP1 = 0
	core.State.ROP2 = 0

}

func (core *Core) SaveState(path string) (err error) {
	err = nil

	return
}

func (core *Core) LoadState(path string) (err error) {
	err = nil

	return
}

func (core *Core) Run() { //Unused
	isStepping := false
	if !isStepping {
		for core.State.Running {
			core.Tick()
		}
	} else {
		for core.State.Running {
			core.Dump()
			var temp string
			fmt.Scanln(&temp)
			core.Tick()
		}
	}
	core.Exit()
}

func (core *Core) Init() {
}

func (core *Core) Exit() {

}

func (core *Core) Dump() {
	fmt.Println("[DUMP]")
	fmt.Println("\t[Registers]")
	fmt.Printf("\t\t{")
	for _, v := range core.State.Regfile {
		fmt.Printf("0x%x, ", v)
	}
	fmt.Printf("}\n")
}

func (core *Core) Tick() {
	core.State.Icount++
	currentCell := core.State.Memory[core.State.Regfile[15]]
	currentInst := (currentCell & 0x1F000000) >> 24
	jumpType := (currentCell & 0xF)
	aluStoring := (currentCell & 0x8000000) != 0x8000000

	/*if (core.TX != 0){
	    core.TX = 0
	}*/

	switch currentInst {
	case 0x0: //MOV
		core.readData(3)
		core.writeData(3, core.State.ROP2, true, true, true)
	case 0x1:
		fallthrough //AND
	case 0x9: //ANDS
		core.readData(3)

		result := core.State.ROP1 & core.State.ROP2
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}
	case 0x2:
		fallthrough //OR
	case 0xA: //ORS
		core.readData(3)

		result := core.State.ROP1 | core.State.ROP2
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}
	case 0x3:
		fallthrough //XOR
	case 0xB: //XORS
		core.readData(3)

		result := core.State.ROP1 ^ core.State.ROP2
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}
	case 0x4:
		fallthrough //ADD
	case 0xC: //ADDS
		core.readData(3)
		var temp32 uint32
		temp32 = (uint32)(core.State.ROP1) + (uint32)(core.State.ROP2)

		result := (uint16)(temp32)
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}

		core.State.Carry = false
		if (temp32 & 0x10000) == 0x10000 {
			core.State.Carry = true
		}

		core.State.Overflow = false
		if !(core.State.ROP1&0x8000 == 0x8000) && !(core.State.ROP2&0x8000 == 0x8000) && (temp32&0x8000 == 0x8000) {
			core.State.Overflow = true
		} else if (core.State.ROP1&0x8000 == 0x8000) && (core.State.ROP2&0x8000 == 0x8000) && !(temp32&0x8000 == 0x8000) {
			core.State.Overflow = true
		} //quick rant: THANKS GO! for not being able to use integers for booleans...

	case 0x5:
		fallthrough //ADC
	case 0xD: //ADCS
		core.readData(3)
		var temp32 uint32
		temp32 = (uint32)(core.State.ROP1) + (uint32)(core.State.ROP2)

		if core.State.Carry {
			temp32++
		}

		result := (uint16)(temp32)
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}

		core.State.Carry = false
		if (temp32 & 0x10000) == 0x10000 {
			core.State.Carry = true
		}

		core.State.Overflow = false
		if !(core.State.ROP1&0x8000 == 0x8000) && !(core.State.ROP2&0x8000 == 0x8000) && (temp32&0x8000 == 0x8000) {
			core.State.Overflow = true
		} else if (core.State.ROP1&0x8000 == 0x8000) && (core.State.ROP2&0x8000 == 0x8000) && !(temp32&0x8000 == 0x8000) {
			core.State.Overflow = true
		}

	case 0x6:
		fallthrough //SUB
	case 0xE: //SUBS
		core.readData(3)
		var temp16 uint16
		temp16 = core.State.ROP1 - core.State.ROP2

		result := temp16
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}

		core.State.Carry = (core.State.ROP1 < core.State.ROP2)

		core.State.Overflow = false
		if !(core.State.ROP1&0x8000 == 0x8000) && (core.State.ROP2&0x8000 == 0x8000) && (temp16&0x8000 == 0x8000) {
			core.State.Overflow = true
		} else if (core.State.ROP1&0x8000 == 0x8000) && !(core.State.ROP2&0x8000 == 0x8000) && !(temp16&0x8000 == 0x8000) {
			core.State.Overflow = true
		}

	case 0x7:
		fallthrough //SBB
	case 0xF: //SBBS
		var tempOP2 uint16
		var temp16 uint16

		tempOP2 = core.State.ROP2
		if core.State.Carry {
			tempOP2++
		}

		temp16 = core.State.ROP1 - tempOP2

		result := temp16
		if aluStoring {
			core.writeData(3, result, true, false, true)
		} else {
			core.basicUpdateFlags(result)
		}

		core.State.Carry = (core.State.ROP1 < tempOP2)

		core.State.Overflow = false
		if !(core.State.ROP1&0x8000 == 0x8000) && (core.State.ROP2&0x8000 == 0x8000) && (temp16&0x8000 == 0x8000) {
			core.State.Overflow = true
		} else if (core.State.ROP1&0x8000 == 0x8000) && !(core.State.ROP2&0x8000 == 0x8000) && !(temp16&0x8000 == 0x8000) {
			core.State.Overflow = true
		}

	case 0x8: //SWM
		var temp16 uint16
		temp16 = (core.readData(2) & 0x1FFF)
		core.State.WMask = temp16
		core.basicUpdateFlags(temp16)
	case 0x10: //HLT
		core.State.Running = false
		fmt.Println("[Debug] received HLT")
	case 0x11: //J**
		temp := core.State.Regfile[15]
		core.readData(2)
		Z := core.State.Zero
		C := core.State.Carry
		O := core.State.Overflow
		S := core.State.Sign
		switch jumpType {
		case 0x0: //JMP (unconditional   ) true
			if true {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x1: //JN  (never           ) false
			if false {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x2: //JC  (carry           ) C == true
			if C == true {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x3: //JNC (not carry       ) C == false
			if C == false {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x4: //JO  (overflow        ) O == true
			if O == true {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x5: //JNO (not overflow    ) O == false
			if O == false {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x6: //JS  (sign            ) S == true
			if S == true {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x7: //JNS (not sign        ) S == false
			if S == false {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x8: //JZ  (zero            ) Z == true
			if Z == true {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0x9: //JNZ (not zero        ) Z == false
			if Z == false {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xA: //JLE (lesser or equal ) (Z == true) || (S != O)
			if (Z == true) || (S != O) {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xB: //JG  (greater         ) (Z == false) && (S == O)
			if (Z == false) && (S == O) {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xC: //JL  (lesser          ) S != O
			if S != O {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xD: //JGE (greater or equal) S == O
			if S == O {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xE: //JBE (below or equal  ) (C == true) || (Z == true)
			if (C == true) || (Z == true) {
				core.State.Regfile[15] = core.State.ROP1
			}
		case 0xF: //JA  (above           ) (C == false) && (Z == false)
			if (C == false) && (Z == false) {
				core.State.Regfile[15] = core.State.ROP1
			}
		}
		if temp != core.State.Regfile[15] {
			core.State.Regfile[15]--
		}
	case 0x12: ///ROL
		core.readData(3)
		shiftamount := core.State.ROP2 & 0xF
		if shiftamount != 0 {
			temp := (core.State.ROP1 << shiftamount) | (core.State.ROP1 >> (16 - shiftamount))
			core.writeData(3, temp, true, false, true)
		}
	case 0x13: ///ROR
		core.readData(3)
		shiftamount := core.State.ROP2 & 0xF
		if shiftamount != 0 {
			temp := (core.State.ROP1 >> shiftamount) | (core.State.ROP1 << (16 - shiftamount))
			core.writeData(3, temp, true, false, true)
		}
	case 0x14:
		fallthrough //SHL
	case 0x16: //SCL
		core.readData(3)
		var shiftamount uint16

		shiftamount = core.State.ROP2 & 0xF
		if currentInst == 0x16 { //if it is chained
			shiftamount = core.State.ChainedOp & 0xF
		}

		core.writeData(3, (core.State.ROP1 << shiftamount), true, false, true)
	case 0x15:
		fallthrough //SHR
	case 0x17: //SCR
		core.readData(3)
		var shiftamount uint16

		shiftamount = core.State.ROP2 & 0xF
		if currentInst == 0x17 { //if it is chained
			shiftamount = core.State.ChainedOp & 0xF
		}

		core.writeData(3, (core.State.ROP1 >> shiftamount), true, false, true)

	//-------------------IO INSTRUCTIONS-------------------

	case 0x18: ///BUMP
		core.readData(1)

		/*core.TX = 0x10000
		  core.PS = (uint8)(core.State.ROP1 & 0xFF)*/

		core.PMngr.TXToPort((uint8)(core.State.ROP1&0xFF), 0x10000)

	case 0x19: ///WAIT //I'm assuming there's a fault in the documentation since this is class 1 but you cannot write data to primary on class 1

		/*core.writeData(2, 0xFFFF, true, false, true) //flags set for sign
		  if(core.RX == 0x10000){
		      core.writeData(2, (uint16)(core.PS), true, false, true)
		  }*/

		core.writeData(2, core.PMngr.GetARequest(), true, false, true)

	case 0x1A: //SEND
		core.readData(3)

		/*core.PS = (uint8)(core.State.ROP1 & 0xFF)
		  core.TX = ((uint32)(core.State.ROP2) | 0x20000)*/

		core.PMngr.TXToPort((uint8)(core.State.ROP1&0xFF), ((uint32)(core.State.ROP2) | 0x20000))

	case 0x1B: //RECV
		core.readData(3)

		data, gotData := core.PMngr.GetTXDataOfPort((uint8)(core.State.ROP2 & 0xFF))
		core.writeData(2, data, false, false, true)
		core.State.Carry = gotData

		core.State.Carry = true
		/*if(((core.RX & 0x20000) == 0x20000) && (core.PS == (uint8)(core.State.ROP2 & 0xFF))){
		    core.State.Carry = false
		    core.writeData(2, (uint16)(core.RX & 0xFFFF), false, false, true)
		}*/

	//-------------------IO INSTRUCTIONS-------------------

	case 0x1C: //PUSH
		core.readData(2)
		core.State.Regfile[14]--
		core.State.Memory[core.State.Regfile[14]] = (uint32)(core.State.ROP1) | 0x20000000
	case 0x1D: //POP
		val := core.State.Memory[core.State.Regfile[14]]
		core.writeData(2, (uint16)(val&0xFFFF), true, false, true)
		core.State.Regfile[14]++
	case 0x1E: //CALL
		core.readData(2)
		core.State.Regfile[14]--
		core.State.Memory[core.State.Regfile[14]] = (uint32)(core.State.Regfile[15]) | 0x20000000
		core.State.Regfile[15] = core.State.ROP1
		core.State.Regfile[15]--
	case 0x1F: //RET
		core.State.Regfile[15] = (uint16)(core.State.Memory[core.State.Regfile[14]] & 0xFFFF)
		core.State.Regfile[14]++
	default:
	}
	core.State.Regfile[15]++
	core.State.ChainedOp = core.State.ROP1

	core.PMngr.TickPeripherals()
}

func (core *Core) basicUpdateFlags(data uint16) {
	core.State.Sign = false
	if (data & 0x8000) == 0x8000 {
		core.State.Sign = true
	}

	core.State.Zero = false
	if data == 0 {
		core.State.Zero = true
	}

	core.State.Carry = false
	core.State.Overflow = false
}

func (core *Core) writeData(class int, data uint16, updateFlags bool, useWM bool, store bool) {
	currentCell := core.State.Memory[core.State.Regfile[15]]
	mode := (currentCell & 0xF00000) >> 20
	sign := (((currentCell & 0x8000) >> 15) == 1)

	if class == 0 || class == 1 || class > 3 {
		return
	}

	/* Modes other than 0, 4, 5 ,C, D accesses are undefined for class 1
	 *
	 * [19:51] <LBPHacker> in the case of modes 8 and A, it just does mode 0
	 * [19:51] <LBPHacker> and mode 2
	 * [19:51] <LBPHacker> respectively
	 */

	switch mode {
	case 0x9:
		fallthrough // OPER REG_R1, [REG_RB +- REG_R2] -- shifts: 0, 16, 4
	case 0xB:
		fallthrough // OPER REG_R1, [REG_RB +- U11_I1] -- shifts: 0, 16, 4
	case 0x8:
		fallthrough // mode 0
	case 0x0:
		fallthrough // OPER REG_R1, REG_R2 -- shifts: 0, 4
	case 0x1:
		fallthrough // OPER REG_R1, [REG_R2] -- shifts: 0, 4
	case 0xA:
		fallthrough // mode 2
	case 0x2:
		fallthrough // OPER REG_R1, U16_I1 -- shifts: 0, 4
	case 0x3: // OPER REG_R1, [U16_I1] -- shifts: 0, 4
		core.State.Regfile[currentCell&0xF] = data

	case 0x4:
		fallthrough // OPER [REG_R1], REG_R2 -- shifts: 0, 4
	case 0x6: // OPER [REG_R1], U16_I1 -- shifts: 0, 4
		core.State.Memory[core.State.Regfile[currentCell&0xF]] = (uint32)(data) | 0x20000000

		if useWM {
			core.State.Memory[core.State.Regfile[currentCell&0xF]] |= ((uint32)(core.State.WMask) << 16)
		}

	case 0x5:
		fallthrough // OPER [U16_I1], REG_R2 -- shifts: 4, 0
	case 0x7: // OPER [U16_I1], U4_I2 -- shifts: 4, 0
		core.State.Memory[(currentCell&0xFFFF0)>>4] = (uint32)(data) | 0x20000000

		if useWM {
			core.State.Memory[(currentCell&0xFFFF0)>>4] |= ((uint32)(core.State.WMask) << 16)
		}

	case 0xE:
		fallthrough // OPER [REG_RB +- REG_R1], U11_I2 -- shifts: 16, 0, 4
	case 0xC: // OPER [REG_RB +- REG_R1], REG_R2 -- shifts: 16, 0, 4
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		REG_R1 := (uint32)(core.State.Regfile[currentCell&0xF])
		if sign {
			addr = REG_RB - REG_R1
		} else {
			addr = REG_RB + REG_R1
		}
		core.State.Memory[addr] = (uint32)(data) | 0x20000000

		if useWM {
			core.State.Memory[addr] |= ((uint32)(core.State.WMask) << 16)
		}

	case 0xF:
		fallthrough // OPER [REG_RB +- U11_I1], U4_I2 -- shifts: 16, 4, 0
	case 0xD: // OPER [REG_RB +- U11_I1], REG_R2 -- shifts: 16, 4, 0
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		U11_I1 := (currentCell & 0x7FF0) >> 4
		if sign {
			addr = REG_RB - U11_I1
		} else {
			addr = REG_RB + U11_I1
		}
		core.State.Memory[addr] = (uint32)(data) | 0x20000000

		if useWM {
			core.State.Memory[addr] |= ((uint32)(core.State.WMask) << 16)
		}
	default:

	}

	if updateFlags {
		core.basicUpdateFlags(data)
	}
}

func (core *Core) readData(class int) (quickOP uint16) {
	currentCell := core.State.Memory[core.State.Regfile[15]]
	mode := (currentCell & 0xF00000) >> 20
	sign := (((currentCell & 0x8000) >> 15) == 1)

	var TOP1, TOP2 uint16

	switch mode {
	case 0x9:
		fallthrough // OPER REG_R1, [REG_RB +- REG_R2] -- shifts: 0, 16, 4
	case 0xB:
		fallthrough // OPER REG_R1, [REG_RB +- U11_I1] -- shifts: 0, 16, 4
	case 0x8:
		fallthrough // mode 0
	case 0x0:
		fallthrough // OPER REG_R1, REG_R2 -- shifts: 0, 4
	case 0x1:
		fallthrough // OPER REG_R1, [REG_R2] -- shifts: 0, 4
	case 0xA:
		fallthrough // mode 2
	case 0x2:
		fallthrough // OPER REG_R1, U16_I1 -- shifts: 0, 4
	case 0x3: // OPER REG_R1, [U16_I1] -- shifts: 0, 4
		TOP1 = core.State.Regfile[currentCell&0xF]

	case 0x4:
		fallthrough // OPER [REG_R1], REG_R2 -- shifts: 0, 4
	case 0x6: // OPER [REG_R1], U16_I1 -- shifts: 0, 4
		TOP1 = (uint16)(core.State.Memory[core.State.Regfile[currentCell&0xF]])

	case 0x5:
		fallthrough // OPER [U16_I1], REG_R2 -- shifts: 4, 0
	case 0x7: // OPER [U16_I1], U4_I2 -- shifts: 4, 0
		TOP1 = (uint16)(core.State.Memory[(currentCell&0xFFFF0)>>4])

	case 0xE:
		fallthrough // OPER [REG_RB +- REG_R1], U11_I2 -- shifts: 16, 0, 4
	case 0xC: // OPER [REG_RB +- REG_R1], REG_R2 -- shifts: 16, 0, 4
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		REG_R1 := (uint32)(core.State.Regfile[currentCell&0xF])
		if sign {
			addr = REG_RB - REG_R1
		} else {
			addr = REG_RB + REG_R1
		}
		TOP1 = (uint16)(core.State.Memory[addr])

	case 0xF:
		fallthrough // OPER [REG_RB +- U11_I1], U4_I2 -- shifts: 16, 4, 0
	case 0xD: // OPER [REG_RB +- U11_I1], REG_R2 -- shifts: 16, 4, 0
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		U11_I1 := (currentCell & 0x7FF0) >> 4
		if sign {
			addr = REG_RB - U11_I1
		} else {
			addr = REG_RB + U11_I1
		}
		TOP1 = (uint16)(core.State.Memory[addr])
	default:

	}

	/*if(currentCell != 0x20000000){
	    fmt.Printf("%x\n", currentCell)
	}*/

	switch mode {
	case 0x8:
		fallthrough // mode 0
	case 0x4:
		fallthrough // OPER [REG_R1], REG_R2 -- shifts: 0, 4
	case 0xC:
		fallthrough // OPER [REG_RB +- REG_R1], REG_R2 -- shifts: 16, 0, 4
	case 0x0: // OPER REG_R1, REG_R2 -- shifts: 0, 4
		TOP2 = (uint16)(core.State.Regfile[(currentCell&0xF0)>>4])

	case 0x1: // OPER REG_R1, [REG_R2] -- shifts: 0, 4
		TOP2 = (uint16)(core.State.Memory[core.State.Regfile[(currentCell&0xF0)>>4]])

	case 0xA:
		fallthrough // mode 2
	case 0x6:
		fallthrough // OPER [REG_R1], U16_I1 -- shifts: 0, 4
	case 0x2: // OPER REG_R1, U16_I1 -- shifts: 0, 4
		TOP2 = (uint16)((currentCell & 0xFFFF0) >> 4)

	case 0x3: // OPER REG_R1, [U16_I1] -- shifts: 0, 4
		TOP2 = (uint16)(core.State.Memory[(currentCell&0xFFFF0)>>4])

	case 0xD:
		fallthrough // OPER [REG_RB +- U11_I1], REG_R2 -- shifts: 16, 4, 0
	case 0x5: // OPER [U16_I1], REG_R2 -- shifts: 4, 0
		TOP2 = core.State.Regfile[currentCell&0xF]

	case 0xF:
		fallthrough // OPER [REG_RB +- U11_I1], U4_I2 -- shifts: 16, 4, 0
	case 0x7: // OPER [U16_I1], U4_I2 -- shifts: 4, 0
		TOP2 = (uint16)(currentCell & 0xF)

	case 0x9: // OPER REG_R1, [REG_RB +- REG_R2] -- shifts: 0, 16, 4
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		REG_R2 := (uint32)(core.State.Regfile[(currentCell&0xF0)>>4])

		if sign {
			addr = REG_RB - REG_R2
		} else {
			addr = REG_RB + REG_R2
		}

		TOP2 = (uint16)(core.State.Memory[addr])
	case 0xB: // OPER REG_R1, [REG_RB +- U11_I1] -- shifts: 0, 16, 4
		var addr uint32
		REG_RB := (uint32)(core.State.Regfile[(currentCell&0xF0000)>>16])
		U11_I1 := (currentCell & 0x7FF0) >> 4

		if sign {
			addr = REG_RB - U11_I1
		} else {
			addr = REG_RB + U11_I1
		}

		TOP2 = (uint16)(core.State.Memory[addr])
	case 0xE: // OPER [REG_RB +- REG_R1], U11_I2 -- shifts: 16, 0, 4
		TOP2 = (uint16)((currentCell & 0x7FF0) >> 4)
	default:

	}

	switch class {
	case 1:
		core.State.ROP1 = TOP1
		quickOP = TOP1
	case 2:
		core.State.ROP1 = TOP2
		quickOP = TOP2
	case 3:
		core.State.ROP1 = TOP1
		core.State.ROP2 = TOP2
		quickOP = TOP1
	default:

	}

	return
}
