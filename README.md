![Nestor](logo.png)

# Nestor - NES emulator

Nestor is a work in progress NES/Famicom emulator written in Go.

## Build from source

### Gio

Uses [Gio](gioui.org), a crossplatform GUI for Go, so you first need to install its dependencies.
Please refer to [gioui.org/doc/install](https://gioui.org/doc/install).

### Nesto

Requires at least go1.22.

Then you can directly download, build and install `nestor` in your $PATH with:

```
go install github.com/arl/nestor@latest
```

## Usage

```
nestor [options] rom

Options:
  -cpuprofile string
        write cpu profile to file
  -log string
        enable logging for specified modules
  -nolog
        disable all logging
  -reset int
        overwrite CPU reset vector with (default: rom-defined) (default -1)
  -rominfos
        print infos about the iNes rom and exit
  -trace value
        write cpu trace log to [file|stdout|stderr] (warning: quickly gets very big)
```


## TODO

 - [x] CPU
 - [x] PPU (background)
 - [ ] PPU (sprites)
 - [ ] APU (sound)
 - [ ] debugger
