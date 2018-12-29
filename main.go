package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gotk3/gotk3/cairo"
	_ "github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"GOR2VM/Core"
	_ "GOR2VM/PeripheralManager"
	"GOR2VM/PeripheralManager/Peripherals/DumbConsole"
	framebuffer "GOR2VM/PeripheralManager/Peripherals/FrameBuffer"
)

var (
	core        *Core.Core
	runningLoop bool
	loadpath    string
	loaded      bool
	memEditAddr int

	builder         *gtk.Builder
	builder_signals map[string]interface{}

	mainWindow           *gtk.Window
	mainWindow_titleBase string
	core_toggleButton    *gtk.ToggleButton

	inspectorWindow             *gtk.Window
	inspectorWindow_registers   []*gtk.Label
	inspector_memEditor_value   *gtk.Entry
	inspector_memEditor_address *gtk.Entry

	termWindow             *gtk.Window
	termWindow_consoleView *gtk.TextView
	consoleView_port       uint8
	termWindow_show        bool

	framebufferWindow             *gtk.Window
	framebufferWindow_framebuffer *gtk.DrawingArea
	onAreaDraw                    func(widget *gtk.Widget, cr *cairo.Context) bool

	//for dynamic UI control:
)

func init() {
	runtime.LockOSThread() //for sfml
	builder_signals = make(map[string]interface{})
	//  builder_signals[""] =
	builder_signals["onToggle"] = onToggle
	builder_signals["onStep"] = onStep
	builder_signals["onDestroy"] = onDestroy
	builder_signals["onReset"] = onReset
	builder_signals["onMMRead"] = onMMRead
	builder_signals["onMMWrite"] = onMMWrite
	builder_signals["onSetMult"] = onSetMult
	builder_signals["onLoad"] = onLoad
	builder_signals["onFileChoose"] = onFileChoose

	loadpath = "a.bin"
	mainWindow_titleBase = "R216K64A DVM"
	runningLoop = false
	loaded = false

	core = Core.NewCore()
	core.MUL = 1
	consoleView_port = 0

	gtk.Init(&os.Args)

	builder, bErr := gtk.BuilderNew()
	if bErr != nil {
		log.Fatal(bErr)
		os.Exit(1)
	}

	bfErr := builder.AddFromFile("GUI.glade")
	if bfErr != nil {
		log.Fatal(bfErr)
		os.Exit(1)
	}

	var ok bool

	windowTemp, _ := builder.GetObject("window_main")
	mainWindow, ok = windowTemp.(*gtk.Window)
	if !ok {
		panic("couldn't type assert object")
	}

	termWindowTemp, _ := builder.GetObject("window_terminal")
	termWindow, ok = termWindowTemp.(*gtk.Window)
	if !ok {
		panic("couldn't type assert object")
	}

	inspectorWindowTemp, _ := builder.GetObject("window_inspector")
	inspectorWindow, ok = inspectorWindowTemp.(*gtk.Window)
	if !ok {
		panic("couldn't type assert object")
	}

	core_toggleButtonTemp, _ := builder.GetObject("ctrl_toggle")
	core_toggleButton, ok = core_toggleButtonTemp.(*gtk.ToggleButton)
	if !ok {
		panic("couldn't type assert object")
	}

	termWindow_consoleViewTemp, _ := builder.GetObject("consolewindow")
	termWindow_consoleView, ok = termWindow_consoleViewTemp.(*gtk.TextView)
	if !ok {
		panic("couldn't type assert object")
	}

	inspector_memEditor_valueTemp, _ := builder.GetObject("medit_value")
	inspector_memEditor_value, ok = inspector_memEditor_valueTemp.(*gtk.Entry)
	if !ok {
		panic("couldn't type assert object")
	}

	inspector_memEditor_addressTemp, _ := builder.GetObject("medit_address")
	inspector_memEditor_address, ok = inspector_memEditor_addressTemp.(*gtk.Entry)
	if !ok {
		panic("couldn't type assert object")
	}

	ok2 := true

	inspectorWindow_registers = make([]*gtk.Label, 17)
	regmaptemp_0, _ := builder.GetObject("regs_0")
	inspectorWindow_registers[0], ok = regmaptemp_0.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_1, _ := builder.GetObject("regs_1")
	inspectorWindow_registers[1], ok = regmaptemp_1.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_2, _ := builder.GetObject("regs_2")
	inspectorWindow_registers[2], ok = regmaptemp_2.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_3, _ := builder.GetObject("regs_3")
	inspectorWindow_registers[3], ok = regmaptemp_3.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_4, _ := builder.GetObject("regs_4")
	inspectorWindow_registers[4], ok = regmaptemp_4.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_5, _ := builder.GetObject("regs_5")
	inspectorWindow_registers[5], ok = regmaptemp_5.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_6, _ := builder.GetObject("regs_6")
	inspectorWindow_registers[6], ok = regmaptemp_6.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_7, _ := builder.GetObject("regs_7")
	inspectorWindow_registers[7], ok = regmaptemp_7.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_8, _ := builder.GetObject("regs_8")
	inspectorWindow_registers[8], ok = regmaptemp_8.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_9, _ := builder.GetObject("regs_9")
	inspectorWindow_registers[9], ok = regmaptemp_9.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_10, _ := builder.GetObject("regs_10")
	inspectorWindow_registers[10], ok = regmaptemp_10.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_11, _ := builder.GetObject("regs_11")
	inspectorWindow_registers[11], ok = regmaptemp_11.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_12, _ := builder.GetObject("regs_12")
	inspectorWindow_registers[12], ok = regmaptemp_12.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_13, _ := builder.GetObject("regs_13")
	inspectorWindow_registers[13], ok = regmaptemp_13.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_14, _ := builder.GetObject("regs_14")
	inspectorWindow_registers[14], ok = regmaptemp_14.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_15, _ := builder.GetObject("regs_15")
	inspectorWindow_registers[15], ok = regmaptemp_15.(*gtk.Label)
	ok2 = ok2 && ok
	regmaptemp_wm, _ := builder.GetObject("regs_wm1")
	inspectorWindow_registers[16], ok = regmaptemp_wm.(*gtk.Label)
	ok2 = ok2 && ok

	if !ok2 {
		panic("reg")
	}

	builder.ConnectSignals(builder_signals)
	updateLastMessage("welcome")

	///Peripherals:

	peripherals_dumbConsole := DumbConsole.NewDumbConsole(termWindow_consoleView).NewPeripheral()
	core.PMngr.RegisterPeripheral(0, peripherals_dumbConsole)

	peripherals_frameBuffer_obj := framebuffer.NewFrameBuffer(320, 240, 2, "Ass")
	peripherals_frameBuffer := peripherals_frameBuffer_obj.NewPeripheral()
	core.PMngr.RegisterPeripheral(1, peripherals_frameBuffer)

	///

	glib.TimeoutAdd(17, func() bool { //~60Hz (58)
		if !core.State.Running || !runningLoop {
			core_toggleButton.SetActive(false)
			return true
		}

		for i := core.MUL; i > 0; i-- {
			b := onStepCond()
			if !b {
				break
			}
		}
		return true
	}, nil)
	updateLastMessage("please load a file")
}

func main() {
	core.Reset()

	updateRegisterView()

	core.State.Running = true

	mainWindow.ShowAll()
	termWindow.ShowAll()
	inspectorWindow.ShowAll()
	gtk.Main()
}

func onToggle() {
	runningLoop = !runningLoop
	core.State.Running = runningLoop
}

func onStep() {
	runningLoop = false
	core_toggleButton.SetActive(false)

	onStepRaw()

	runningLoop = false
	core_toggleButton.SetActive(false) ///why the hell do we need to do this for a second time?
}

func onStepRaw() {
	if !loaded {
		return
	}
	core.Tick()
	updateRegisterView()
	/*if(core.TX != 0 && ((core.TX & 0x20000) == 0x20000)){
	    buf, _ := termWindow_consoleView.GetBuffer()
	    buf.InsertAtCursor(fmt.Sprintf("%c", core.TX & 0xFF))
	}*/
}

func onStepCond() bool {
	if !(core.State.Running && runningLoop) {
		return false
	}
	onStepRaw()
	return true
}

func onReset() {
	core.Reset()
	updateRegisterView()
}

func onDestroy() {
	gtk.MainQuit()
	core.Exit()
	os.Exit(0)
}

func onMMRead() {
	addr, err := getMEdit(true)

	updateMEditVal(addr, (err != nil))
}

func onMMWrite() {
	addr, err := getMEdit(true)
	if err != nil {
		return
	}
	val, err := getMEdit(false)
	if err != nil {
		return
	}

	core.State.Memory[addr] = (uint32)(val) | 0x20000000
}

func getMEdit(Address bool) (addr uint16, err error) {
	addr = 0
	err = errors.New("error")

	var valString string
	var serr error

	if Address {
		valString, serr = inspector_memEditor_address.GetText()
	} else {
		valString, serr = inspector_memEditor_value.GetText()
	}

	if serr != nil {
		updateMEditVal((uint16)(0), true)
		return
	}

	addrSlice, aerr := hex.DecodeString(valString)
	if aerr != nil || len(addrSlice) != 2 {
		updateMEditVal((uint16)(0), true)
		return
	}

	addr |= (uint16(addrSlice[0]) << 8)
	addr |= uint16(addrSlice[1])
	err = nil
	return
}

func updateMEditVal(address uint16, errored bool) {
	addressBytes := make([]byte, 2)
	addressBytes[0] = byte(core.State.Memory[address] >> 8)
	addressBytes[1] = byte(core.State.Memory[address] & 0xFF)

	if errored {
		inspector_memEditor_value.SetText("")
	} else {
		inspector_memEditor_value.SetText(hex.EncodeToString(addressBytes))
	}

}

func onSetMult(s *gtk.SpinButton) {
	core.MUL = (uint32)(s.GetValueAsInt())
}

func updateRegisterView() {
	for k, v := range core.State.Regfile {
		inspectorWindow_registers[k].SetLabel(fmt.Sprintf("%#0004x", v))
	}
	inspectorWindow_registers[16].SetLabel(fmt.Sprintf("%#00004x", core.State.WMask))
}

func updateLastMessage(text string) {
	mainWindow.SetTitle(mainWindow_titleBase + " - " + text)
}

func onPortChange(s *gtk.SpinButton) {
	consoleView_port = (uint8)(s.GetValueAsInt() & 0xFF)
}

func onLoad() {
	loaded = true

	binaryErr := core.LoadFromBinary(loadpath)

	if binaryErr != nil {
		fmt.Println(binaryErr)
		updateLastMessage(binaryErr.Error())
		loaded = false
	}

	core.Reset()
}

func onFileChoose(f *gtk.FileChooserButton) {
	loadpath = f.FileChooser.GetFilename()
}
