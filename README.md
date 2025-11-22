<p align="center">
 <img src="./assets/graphics/logo.png" width="384" align="center">
</p>

# Nestor - NES emulator

Nestor is a NES/Famicom emulator written in Go.


| ![adventures of rad gravity](https://github.com/user-attachments/assets/014025c9-6c7e-4f68-b351-3557c345a12e) | ![battletoads](https://github.com/user-attachments/assets/d7a03db0-fcf7-4e8f-a8f7-23ec0d01fae7) | ![tsuppari oozumou](https://github.com/user-attachments/assets/534e5d32-7bf0-48a1-9b3e-bb580f651585) |
|----|----|----|
| ![castevania](https://github.com/user-attachments/assets/8b283d1f-9eca-49da-849f-d4c9c91f98cd) | ![prince of persia](https://github.com/user-attachments/assets/cdb49c3e-4ac4-4dd9-94fe-ac4d91af4aff) | ![contra](https://github.com/user-attachments/assets/a59fbc21-4938-441d-81d7-1dabda65c929) |


- [Nestor - NES emulator](#nestor---nes-emulator)
  - [Features](#features)
    - [Implemented mappers](#implemented-mappers)
  - [Installation](#installation)
    - [MacOS - homebrew](#macos---homebrew)
    - [Build from source](#build-from-source)
  - [Usage](#usage)
  - [UI Screenshots](#ui-screenshots)
  - [Thanks](#thanks)
  - [License](#license)



## Features

All these features are planned, but not all of them are implemented yet.

 - [x] Cycle accurate CPU
 - [X] PPU (Picture Processing Unit)
 - [x] NTSC (USA / Japan)
 - [ ] PAL (Europe)
 - [x] Joystick/Joypad support
 - [x] APU (Audio Processing Unit)
 - [x] CRT Shader effects
 - [ ] Save state
 - [x] Frame run-ahead

### Implemented mappers

A NES games cartridge is made up of various circuits and hardware, which varies from game to game. The configuraion and capabilities of such cartridges is commonly called their mapper. Mappers are designed to extend the system and bypass its limitations, such as by adding RAM to the cartridge or even extra sound channels.

| Name  | iNES mapper | Implemented |
|-------|------------:|:-----------:|
| NROM  |           0 |     [x]     |
| MMC1  |           1 |     [x]     |
| UxROM |           2 |     [x]     |
| CNROM |           3 |     [x]     |
| MMC3  |           4 |     [ ]     |
| MMC5  |          10 |     [ ]     |
| AxROM |           7 |     [x]     |
| GxROM |          66 |     [x]     |


## Installation

### MacOS - homebrew

```
brew tap arl/arl
brew install nestor
```

### Build from source

 1. Install Go, following installation instructions at https://go.dev/doc/install (>= go1.25 is required).
 2. Nestor uses the Ebitengine library, which requires some dependencies to be installed first. Follow the [Ebitengine installation instructions](https://ebitengine.org/en/documents/install.html) for your platform.
 3. Clone github.com/arl/nestor
 4. Run `go build` in the repository folder.


## Usage

You can either directly run a rom file with:

```
$ ./nestor /path/to/rom.nes
```

or start the Graphical User Interface (GUI) mode:

```
$ ./nestor
```

Run `nestor -help` for more options.



## UI Screenshots

| ![mainwindow rom selection](https://github.com/user-attachments/assets/2515bce2-a926-40f0-9213-2505d87f102b) | 
|:--:| 
| *Main window / Rom selection* |


| ![emuwindow gamepanel](https://github.com/user-attachments/assets/5b4b7e7a-b8af-4f81-83c1-2df4f1814591) | 
|:--:| 
| *Emulator window with accompanying in-game controls window* |


| ![input config ui](https://github.com/user-attachments/assets/4add9e06-1eff-4bb0-82f0-c4e2f6583e59) | 
|:--:| 
| *Input configuration page* |

## Thanks

Many thanks to:
 - @genbs for the help on macos x!
 - @tommyblue for the paddle!
 - @rasky for [ndsemu](https://github.com/rasky/ndsemu) codebase!
 - [NesDev Wifi](https://www.nesdev.org/wiki/Nesdev_Wiki) for the great documentation and community!

## License

Nestor is available under the GPL V3 license.  Full text here: <http://www.gnu.org/licenses/gpl-3.0.en.html>

Copyright (C) 2023-2025 Aurélien Rainone

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
