# GOR2VM  
A rewrite of my R216 emulator now in Golang because why not.  
  
## Features  
- A simple GUI made with (go binding of) GTK+ 3 and Glade  
- A "framebuffer" which uses (GO)SFML 2.4  
  
## How to run  
`wget https://github.com/LBPHacker/R216/raw/master/r2asm.lua -O ./Tools/r2asm.lua`  
`lua ./Tools/r2asm.lua headless_model=R216DVM headless_out=a.bin <assembly>`  
`go run main.go`  
Load the a.bin file in the GUI  
Press load  
Press run  
  
Recommended demo to run is ./Tools/examples/bfcompiler.asm  
By default, the program it compiles is a hello world  

## GUI in detail
### Main window
- ON/OFF: Toggles the running state of the VM manager (Might not actually toggle the execution)
- STEP: If VM execution is active, stops it. Steps the processor 1 instruction. Useful for debugging
- RESET: Resets the registers and reloads the memory from the ROM
- LOAD: Loads the selected file to the ROM and to memory
- File Picker: Picks a file to load from
- Multiplier value spinner: Sets the multiplier at which the processor runs at (execution speed is ((1000/17)*Multiplier)Hz)

[Manual](https://lbphacker.pw/powdertoy/R216/manual.md)  
  
![works on my machine](https://johan.driessen.se/images/johan_driessen_se/WindowsLiveWriter/PersistanceinWF4beta2_E4AD/works-on-my-machine-starburst_2.png)  

### TODO  
 - Implement the """framebuffer"""
 - Implement the console input
