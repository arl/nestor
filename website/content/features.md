---
title: "Features"
description: "What Nestor can do - features and capabilities of this NES emulator."
---

# Features

Here's what Nestor currently supports and what's planned for the future.

## Core Emulation

### ✅ CPU (6502)
The emulator tries to accurately simulate the NES's 6502 processor, including cycle timing. This helps games run more like they would on real hardware.

### ✅ PPU (Graphics)
Handles most NES graphics features:
- Sprite rendering  
- Background tiles
- Scrolling
- Color palettes

### ✅ NTSC Video
Supports the standard NTSC timing used by American and Japanese NES systems.

### ⏳ PAL Support
PAL (European) timing isn't implemented yet, but it's on the todo list.

## Audio

### ✅ APU (Sound)
Has the NES audio processing unit with:
- All 5 sound channels (2 pulse, 1 triangle, 1 noise, 1 DMC)
- Decent sound quality

## Input and Controls

### ✅ Controllers
- Keyboard input
- Joystick/gamepad support  
- Basic input mapping

## Performance Features

### ✅ Frame Run-Ahead
Can reduce input lag by predicting frames ahead of time. Useful for games where timing matters.

## Planned Features

### 🔄 Debugger
Would be nice to have debugging tools for looking at CPU state and memory.

### 🔄 Save States
Would be useful to have save/load functionality for quick saves.

## Mapper Support

Nestor supports several mapper types. See the [mappers page](/mappers/) for more details:

| Mapper | Name | Status | Examples |
|--------|------|--------|----------|
| 0 | NROM | ✅ Works | Super Mario Bros, Donkey Kong |
| 1 | MMC1 | ✅ Works | The Legend of Zelda, Metroid |
| 2 | UxROM | ✅ Works | Mega Man, Contra |
| 3 | CNROM | ✅ Works | Solomon's Key, Arkanoid |
| 7 | AxROM | ✅ Works | Battletoads, Wizards & Warriors |
| 66 | GxROM | ✅ Works | SMB + Duck Hunt |
| 4 | MMC3 | 🔄 TODO | Super Mario Bros. 3, Mega Man 2 |
| 10 | MMC5 | 🔄 TODO | Castlevania III |

## Technical Stuff

### What it's built with
- Written in Go 
- Uses SDL2 for graphics and input
- Cross-platform (Windows, macOS, Linux)
- Reads standard NES ROM files

### Interface
- GUI version available
- Command line version too
- Basic configuration options