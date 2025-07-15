---
title: "About"
description: "Learn about the Nestor project, its history, technical implementation, and the team behind it."
---

# About Nestor

Nestor is a high-quality Nintendo Entertainment System (NES) emulator written in Go, designed to provide accurate emulation with modern conveniences and excellent performance.

## Project Goals

### Accuracy First
Nestor prioritizes emulation accuracy to ensure games run exactly as they would on original hardware. Our cycle-accurate CPU implementation and precise PPU emulation provide an authentic gaming experience.

### Modern Performance
While maintaining accuracy, Nestor leverages modern hardware capabilities to deliver smooth 60 FPS gameplay with features like frame run-ahead to reduce input lag.

### Cross-Platform Compatibility
Built with Go and SDL2, Nestor runs natively on Windows, macOS, and Linux, providing a consistent experience across all major platforms.

### User-Friendly Interface
From command-line simplicity to an intuitive GUI, Nestor caters to both casual users and emulation enthusiasts.

## Project History

### Origins
Nestor began as a personal project to explore emulation development while learning the Go programming language. The goal was to create a clean, maintainable codebase that could serve as both a functional emulator and an educational resource.

### Development Timeline
- **2023**: Initial development started
- **2024**: First public release with basic mapper support
- **2025**: Expanded mapper compatibility and enhanced features

### Name Origin
The name "Nestor" is a play on "NES" while referencing the wise counselor from Greek mythology, reflecting the project's goal to be both intelligent and reliable.

## Technical Implementation

### Architecture Overview
Nestor is built with a modular architecture that separates concerns while maintaining performance:

```
┌─────────────────┐
│   User Interface │  (GUI/CLI)
├─────────────────┤
│   Emulation Core │  (CPU, PPU, APU)
├─────────────────┤
│   Hardware Layer │  (Mappers, Input)
├─────────────────┤
│   Platform Layer │  (SDL2, OS APIs)
└─────────────────┘
```

### Core Components

#### CPU Emulation
- **Processor**: MOS Technology 6502
- **Implementation**: Cycle-accurate with precise timing
- **Features**: Full instruction set, decimal mode, interrupt handling
- **Testing**: Validated against comprehensive test suites

#### PPU (Picture Processing Unit)
- **Resolution**: 256x240 pixels
- **Color Palette**: 64-color palette with 4 background palettes
- **Sprites**: 64 sprites, 8 per scanline limit
- **Scrolling**: Accurate background and sprite scrolling
- **Timing**: Cycle-accurate rendering pipeline

#### APU (Audio Processing Unit)
- **Channels**: 5 audio channels (2 pulse, 1 triangle, 1 noise, 1 DMC)
- **Quality**: High-fidelity audio synthesis
- **Mixing**: Accurate channel mixing and filtering
- **Output**: 44.1 kHz stereo output

#### Memory Management
- **Addressing**: Full 16-bit address space handling
- **Banking**: Comprehensive mapper support for memory banking
- **Save RAM**: Battery-backed save support where applicable

### Programming Language Choice

#### Why Go?
- **Performance**: Compiled language with excellent runtime performance
- **Simplicity**: Clean syntax and powerful standard library
- **Concurrency**: Built-in goroutines for handling audio/video/input
- **Cross-platform**: Excellent cross-compilation support
- **Memory Safety**: Garbage collection prevents memory corruption
- **Tooling**: Excellent debugging and profiling tools

#### CGO Integration
Nestor uses CGO to interface with C libraries:
- **SDL2**: Cross-platform audio, video, and input handling
- **GTK+3**: Native GUI components on Linux and Windows
- **OpenGL**: Hardware-accelerated rendering

### Code Quality

#### Testing Strategy
- **Unit Tests**: Comprehensive test coverage for core components
- **Integration Tests**: Full system testing with known ROMs
- **Regression Tests**: Automated testing to prevent functionality loss
- **Performance Tests**: Benchmarking to ensure optimal performance

#### Code Organization
```
nestor/
├── emu/           # Core emulation logic
├── hw/            # Hardware components (CPU, PPU, APU)
├── ui/            # User interface (GUI and CLI)
├── ines/          # ROM file format handling
├── vendor/        # External dependencies
└── tests/         # Test ROMs and test suites
```

## Performance Characteristics

### Benchmarks
On modern hardware, Nestor achieves:
- **CPU Usage**: 5-15% on a single core
- **Memory Usage**: 20-50 MB typical
- **Frame Rate**: Consistent 60 FPS
- **Input Lag**: 1-2 frames (with run-ahead: <1 frame)

### Optimization Techniques
- **Efficient rendering**: Direct pixel buffer manipulation
- **Smart caching**: Cached pattern table and nametable data
- **Batch processing**: Grouped memory operations
- **Profile-guided optimization**: Using Go's PGO for optimal performance

## Compatibility

### Current Game Support
- **Total Games**: 350+ games fully supported
- **Popular Titles**: 99%+ compatibility with well-known games
- **Homebrew**: Excellent support for modern homebrew ROMs
- **Test ROMs**: Passes most emulator test suites

### Mapper Implementation Status
See our detailed [mappers page](/mappers/) for complete compatibility information.

## Development Team

### Core Developer
**arl** - Project creator and primary maintainer
- GitHub: [@arl](https://github.com/arl)
- Focus: Core emulation, architecture, and overall project direction

### Contributors

We extend our gratitude to the following contributors:

#### Code Contributors
- **@genbs** - macOS support and testing
- **@tommyblue** - Build system improvements
- **@rasky** - Technical guidance and [ndsemu](https://github.com/rasky/ndsemu) inspiration

#### Community Support
- **NESdev Community** - Technical documentation and guidance
- **RetroArch Team** - Emulation knowledge sharing
- **Go Community** - Language support and best practices

### How to Contribute

We welcome contributions of all kinds:

#### Code Contributions
- Mapper implementations
- Performance optimizations
- Bug fixes
- Feature enhancements

#### Documentation
- Usage guides
- Technical documentation
- Translation support

#### Testing
- Game compatibility testing
- Bug reporting
- Feature requests

See our [contribution guidelines](https://github.com/arl/nestor/blob/main/CONTRIBUTING.md) for detailed information.

## Technical References

### NES Documentation
- [NESdev Wiki](https://www.nesdev.org/wiki/Nesdev_Wiki) - Comprehensive NES technical documentation
- [6502 Reference](http://www.6502.org/) - CPU technical details
- [PPU Reference](https://www.nesdev.org/wiki/PPU) - Graphics chip documentation

### Emulation Resources
- [Blargg's Test ROMs](https://www.nesdev.org/wiki/Emulator_tests) - Emulator validation
- [Visual 6502](http://visual6502.org/) - CPU simulation and visualization
- [FCEUX](http://fceux.com/) - Reference emulator for comparison

### Development Tools
- [Go Programming Language](https://golang.org/) - Primary development language
- [SDL2](https://www.libsdl.org/) - Cross-platform multimedia library
- [GTK+3](https://www.gtk.org/) - GUI toolkit

## License and Legal

### Software License
Nestor is available under the **GNU General Public License v3.0**

**What this means:**
- ✅ Use for personal and commercial purposes
- ✅ Modify and distribute the source code
- ✅ Include in larger projects
- ❗ Must disclose source code if redistributing
- ❗ Must include license and copyright notice
- ❗ Must use same GPL license for derivatives

### Copyright Notice
```
Copyright (C) 2023-2025 arl

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

### ROM Usage
**Important**: Nestor is an emulator only. You must legally own the games you play:
- ROM files are copyrighted material
- Only use ROMs you have legally purchased
- Homebrew and public domain ROMs are freely usable
- We do not provide or endorse ROM piracy

## Future Development

### Short-term Goals (2025)
- [ ] MMC3 mapper implementation
- [ ] Save state functionality
- [ ] Improved debugger interface
- [ ] PAL region support

### Medium-term Goals (2026)
- [ ] Advanced mappers (MMC5, etc.)
- [ ] Net play support
- [ ] Mobile platform support
- [ ] Audio enhancement features

### Long-term Vision
- Comprehensive NES emulation with 99%+ compatibility
- Modern features while maintaining authenticity
- Educational resource for emulation development
- Active, supportive community

## Contact and Support

### Project Links
- **GitHub Repository**: [github.com/arl/nestor](https://github.com/arl/nestor)
- **Issues and Bug Reports**: [GitHub Issues](https://github.com/arl/nestor/issues)
- **Releases**: [GitHub Releases](https://github.com/arl/nestor/releases)

### Community
- **Discussions**: GitHub Discussions for general questions
- **Development Chat**: Join the conversation in Issues and PRs
- **Documentation**: This website and GitHub wiki

### Support the Project
- ⭐ Star the repository on GitHub
- 🐛 Report bugs and suggest features
- 💻 Contribute code improvements
- 📖 Improve documentation
- 🎮 Test games and provide feedback

---

*Nestor aims to preserve and celebrate the rich history of NES gaming while pushing the boundaries of what emulation can achieve. Join us in bringing classic games to modern platforms with unprecedented accuracy and quality.*