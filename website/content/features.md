---
title: "Features"
description: "Discover all the powerful features that make Nestor a top-tier NES emulator, from cycle-accurate emulation to modern enhancements."
---

# Nestor Features

Nestor provides a comprehensive NES emulation experience with both authentic accuracy and modern conveniences.

## Core Emulation

### ✅ Cycle Accurate CPU
Nestor implements precise 6502 CPU emulation, ensuring games run exactly as they would on original hardware. This level of accuracy is crucial for games that rely on precise timing.

### ✅ PPU (Picture Processing Unit)
Complete implementation of the NES Picture Processing Unit with support for:
- Sprite rendering
- Background tiles
- Scrolling effects
- Color palettes
- All standard PPU features

### ✅ NTSC Support
Full NTSC video standard support for authentic American/Japanese NES experience.

### ⏳ PAL Support
PAL (European) standard support is planned for future releases.

## Audio

### ✅ APU (Audio Processing Unit)
Complete Audio Processing Unit implementation featuring:
- All 5 audio channels (2 pulse, 1 triangle, 1 noise, 1 DMC)
- Accurate sound synthesis
- Audio mixing and filtering
- High-quality audio output

## Input and Controls

### ✅ Joystick/Joypad Support
- Full controller support for modern gamepads
- Customizable input mapping
- Support for multiple controllers
- Keyboard input support
- Configurable key bindings

## Visual Enhancements

### ✅ CRT Shader Effects
Experience games as they were meant to be seen with authentic CRT monitor simulation:
- Scanlines and phosphor effects
- Color bleeding simulation
- Curvature and distortion effects
- Adjustable intensity settings

## Performance Features

### ✅ Frame Run-Ahead
Reduce input lag for competitive gaming with run-ahead technology:
- Predictive frame rendering
- Reduced latency between input and display
- Configurable run-ahead frames
- Maintains game compatibility

## Planned Features

### 🔄 Debugger
Advanced debugging tools for developers and enthusiasts:
- CPU state inspection
- Memory viewer
- Breakpoint support
- Step-by-step execution

### 🔄 Save State
Complete save state functionality:
- Quick save/load
- Multiple save slots
- Save state management
- Cross-session compatibility

## Mapper Support

Nestor supports multiple mapper types for broad game compatibility:

| Mapper | Name | Status | Notable Games |
|--------|------|--------|---------------|
| 0 | NROM | ✅ Complete | Super Mario Bros, Donkey Kong |
| 1 | MMC1 | ✅ Complete | The Legend of Zelda, Metroid |
| 2 | UxROM | ✅ Complete | Mega Man, Contra |
| 3 | CNROM | ✅ Complete | Solomon's Key, Arkanoid |
| 7 | AxROM | ✅ Complete | Battletoads, Wizards & Warriors |
| 66 | GxROM | ✅ Complete | SMB + Duck Hunt |
| 4 | MMC3 | 🔄 Planned | Super Mario Bros. 3, Mega Man 2 |
| 10 | MMC5 | 🔄 Planned | Castlevania III |

## Technical Specifications

### Performance
- Written in Go for excellent performance and cross-platform compatibility
- Efficient memory usage
- Optimized rendering pipeline
- 60 FPS gameplay with accurate timing

### Compatibility
- Cross-platform support (Windows, macOS, Linux)
- Standard ROM file format support
- iNES header parsing
- Multiple file format support

### User Interface
- Clean, intuitive interface
- GTK-based GUI for Linux/Windows
- Native macOS interface support
- Configuration management
- Recent ROMs tracking