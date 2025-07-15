---
title: "About"
description: "About the Nestor project - a side project NES emulator written in Go."
---

# About Nestor

Nestor is a Nintendo Entertainment System (NES) emulator written in Go. It started as a personal side project to learn about emulation and the Go programming language.

## What is it?

Nestor tries to accurately emulate the NES hardware so you can play classic games on modern computers. It focuses on:

- **Cycle accuracy** - The CPU emulation tries to match the timing of the original 6502 processor
- **Cross-platform** - Runs on Windows, macOS, and Linux thanks to Go and SDL2
- **Simple to use** - Both command-line and GUI interfaces available

## Why "Nestor"?

The name is a play on "NES" - plus Nestor was a wise character in Greek mythology, which seemed fitting for an emulator that tries to be smart about accuracy.

## Technical Details

### What's implemented:
- **CPU**: MOS 6502 with cycle-accurate timing
- **PPU**: Picture processing for graphics rendering
- **APU**: Audio processing for sound
- **Mappers**: Support for 6 different mapper types (see [mappers page](/mappers/))
- **Input**: Joystick and keyboard support

### Written in Go because:
- Good performance for an interpreted language
- Nice cross-platform support
- Clean, readable code
- Great standard library

## Current Status

The emulator can run many popular NES games reasonably well. It's not perfect - there are still bugs and missing features, but it's functional enough to enjoy classic games.

### Game Compatibility
- Most popular games work fine
- Some edge cases and less common mappers aren't supported yet
- Homebrew games generally work well

## Contributing

This is an open source project on GitHub. If you find bugs or want to add features, feel free to:
- Report issues on GitHub
- Submit pull requests with fixes or improvements
- Test games and report compatibility

## Legal Stuff

Nestor is licensed under GPL v3. The emulator itself is legal - you just need to make sure you own the ROM files you're playing. Don't pirate games!

## Links

- **Source Code**: [github.com/arl/nestor](https://github.com/arl/nestor)
- **Bug Reports**: [GitHub Issues](https://github.com/arl/nestor/issues)
- **Releases**: [GitHub Releases](https://github.com/arl/nestor/releases)

---

*Nestor is a hobby project made for fun and learning. It's not trying to be the best or most advanced emulator - just a decent one that works for playing classic NES games.*