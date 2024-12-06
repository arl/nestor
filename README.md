<img src="./ui/logo.png" width="256">

# Nestor - NES emulator

Nestor is a work in progress NES/Famicom emulator written in Go.

## Build from source

### GTK3

Uses [GTK3](gtk.org), via github.com/gotk3/gotk3 Go bindings.
Please refer to [github.com/gotk3/gotk3](https://github.com/gotk3/gotk3).

### Nestor

Requires at least go1.22.

Then you can directly download, build and install `nestor` in your $PATH with:

```
go install github.com/arl/nestor@latest
```

## Usage

```
$ nestor --help
Usage: nestor <command> [flags]

NES emulator. github.com/arl/nestor

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  gui [flags]
    Run Nestor graphical user interface. The default if no commands are given.

  run [</path/to/rom>] [flags]
    Run ROM in emulator.

  rom-infos </path/to/rom> [flags]
    Show ROM infos.

Run "nestor <command> --help" for more information on a command.
```

## TODO

 - [x] CPU (6502)
 - [x] PPU (pixel processing unit)
 - [x] Joystick / Keyboard input
 - [X] APU (audio processing unit)
 - [ ] debugger
 - [x] mapper 0
 - [ ] mapper 1
 - [x] mapper 2
 - [ ] other mappers...
 - [ ] save states
