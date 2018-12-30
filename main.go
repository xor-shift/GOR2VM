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
	termWindow_show        bool

	framebufferWindow             *gtk.Window
	framebufferWindow_framebuffer *gtk.DrawingArea
	onAreaDraw                    func(widget *gtk.Widget, cr *cairo.Context) bool
)

func init() {
	runtime.LockOSThread() //for sfml
	builder_signals = make(map[string]interface{})
	//  builder_signals[""] =
	builder_signals["onToggle"] = func() { core.State.Running = !core.State.Running }
	builder_signals["onStep"] = onStep
	builder_signals["onDestroy"] = onDestroy
	builder_signals["onReset"] = onReset
	builder_signals["onMMRead"] = onMMRead
	builder_signals["onMMWrite"] = onMMWrite
	builder_signals["onSetMult"] = func(s *gtk.SpinButton) { core.MUL = (uint32)(s.GetValueAsInt()) }
	builder_signals["onLoad"] = onLoad
	builder_signals["onFileChoose"] = func(f *gtk.FileChooserButton) { loadpath = f.FileChooser.GetFilename() }

	loadpath = "a.bin"
	loaded = false

	core = Core.NewCore()
	core.MUL = 1

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

	regmaptemp := make([]glib.IObject, 17)
	inspectorWindow_registers = make([]*gtk.Label, 17)
	var regmaperror error
	for i := 0; i < 17; i++ {
		regmaptemp[i], regmaperror = builder.GetObject(fmt.Sprintf("regs_%v", i))
		if regmaperror != nil {
			panic(regmaperror.Error())
		}
		inspectorWindow_registers[i], ok = regmaptemp[i].(*gtk.Label)
		if !ok {
			panic(fmt.Sprintf("Error while type asserting register %v", i))
		}
	}

	builder.ConnectSignals(builder_signals)

	///Peripherals:

	peripherals_dumbConsole := DumbConsole.NewDumbConsole(termWindow_consoleView).NewPeripheral()
	core.PMngr.RegisterPeripheral(0, peripherals_dumbConsole)

	peripherals_frameBuffer_obj := framebuffer.NewFrameBuffer(320, 240, 2, "Ass")
	peripherals_frameBuffer := peripherals_frameBuffer_obj.NewPeripheral()
	core.PMngr.RegisterPeripheral(1, peripherals_frameBuffer)

	///

	glib.TimeoutAdd(17, func() bool { //~60Hz (58)
		if !core.State.Running {
			core_toggleButton.SetActive(core.State.Running)
			return true
		}

		for i := core.MUL; i > 0; i-- {
			b := onStepCond()
			if !b {
				break
			}
			if !core.State.Running {
				core_toggleButton.SetActive(core.State.Running)
				return true
			}
		}
		return true
	}, nil)
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

func onStep() {
	onStepRaw()

	core.State.Running = false
	fmt.Println(core.State.Running)
	core_toggleButton.SetActive(core.State.Running)
	fmt.Println(core.State.Running) //true? how?
	core.State.Running = false
	fmt.Println(core.State.Running)
}

func onStepRaw() {
	if !loaded {
		return
	}
	core.Tick()
	updateRegisterView()
}

func onStepCond() bool {
	if !core.State.Running {
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

func updateRegisterView() {
	for k, v := range core.State.Regfile {
		inspectorWindow_registers[k].SetLabel(fmt.Sprintf("%#0004x", v))
	}
	inspectorWindow_registers[16].SetLabel(fmt.Sprintf("%#00004x", core.State.WMask))
}

func onLoad() {
	loaded = true

	binaryErr := core.LoadFromBinary(loadpath)

	if binaryErr != nil {
		fmt.Println(binaryErr)
		loaded = false
	}

	onReset()
}
