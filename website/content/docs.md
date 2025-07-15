---
title: "Documentation"
description: "Complete installation, configuration, and usage guide for Nestor NES emulator."
---

# Nestor Documentation

This comprehensive guide covers everything you need to know about installing, configuring, and using Nestor.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Configuration](#configuration)
- [Controls and Input](#controls-and-input)
- [Troubleshooting](#troubleshooting)
- [Building from Source](#building-from-source)
- [System Requirements](#system-requirements)

## Installation

### macOS - Homebrew (Recommended)

The easiest way to install Nestor on macOS is through Homebrew:

```bash
# Add the tap
brew tap arl/arl

# Install Nestor
brew install nestor
```

### macOS - Build from Source

If you prefer to build from source on macOS:

```bash
# Install dependencies
brew install go gtk+3 sdl2 sdl2_ttf

# Clone and build
git clone https://github.com/arl/nestor.git
cd nestor
go build
```

### Linux - Build from Source

#### Ubuntu/Debian Dependencies

```bash
sudo apt-get update && sudo apt-get install \
  gcc \
  pkg-config \
  libsdl2-dev \
  libgtk-3-dev \
  libglib2.0-dev \
  libgdk-pixbuf-2.0-dev \
  libsdl2-image-dev \
  libsdl2-mixer-dev \
  libsdl2-ttf-dev \
  libsdl2-gfx-dev
```

#### Build Process

```bash
# Install Go 1.24+
# Then clone and build
git clone https://github.com/arl/nestor.git
cd nestor
go build
```

Or install directly:
```bash
go install github.com/arl/nestor@main
```

### Windows

Windows builds are available from [GitHub Releases](https://github.com/arl/nestor/releases). Download the latest Windows executable.

For building from source on Windows, you'll need:
- Go 1.24+
- CGO compiler (like TDM-GCC)
- SDL2 development libraries

## Basic Usage

### Command Line Interface

#### Running a ROM Directly
```bash
nestor run /path/to/your/rom.nes
```

#### GUI Mode (Default)
```bash
nestor
```

#### ROM Information
```bash
nestor rominfo /path/to/your/rom.nes
```

#### Version Information
```bash
nestor --version
```

### Command Line Options

```bash
# View all available options
nestor --help
```

## Configuration

Nestor has a built-in configuration GUI and stores configuration in platform-specific locations:

- **macOS**: `~/Library/Application Support/nestor/config.toml`
- **Linux**: `~/.config/nestor/config.toml`
- **Windows**: `%APPDATA%/nestor/config.toml`

Configuration is managed through the emulator's interface and automatically saved to the TOML file.

## Controls and Input

### Default Keyboard Controls

#### Player 1
| NES Button | Keyboard Key |
|------------|--------------|
| D-Pad Up | ↑ Arrow |
| D-Pad Down | ↓ Arrow |
| D-Pad Left | ← Arrow |
| D-Pad Right | → Arrow |
| A Button | X |
| B Button | Z |
| Start | Enter |
| Select | Right Shift |

#### Player 2
| NES Button | Keyboard Key |
|------------|--------------|
| D-Pad Up | W |
| D-Pad Down | S |
| D-Pad Left | A |
| D-Pad Right | D |
| A Button | G |
| B Button | F |
| Start | T |
| Select | R |

### Gamepad Support

Nestor supports standard gamepads including:
- Xbox controllers
- PlayStation controllers
- Generic USB gamepads
- Bluetooth controllers

#### Gamepad Mapping
- **D-Pad/Left Stick**: NES D-Pad
- **A/Cross Button**: NES A
- **B/Circle Button**: NES B
- **Start**: NES Start
- **Select/Back**: NES Select

### Custom Input Configuration

1. Launch Nestor in GUI mode
2. Go to **Preferences > Input**
3. Select the control to change
4. Press the desired key or button
5. Click **Apply** to save

## File Format Support

### ROM Formats
- **.nes** files (iNES format)
- **Headered ROMs** with iNES headers
- **Headerless ROMs** (auto-detection)

### Unsupported Formats
- .zip files (extract first)
- .7z files (extract first)
- FDS disk images (planned)
- NSF audio files (planned)

## Troubleshooting

### Common Issues

#### Game Won't Load
1. **Check file format**: Ensure the ROM is in .nes format
2. **Verify ROM integrity**: Try a different ROM file
3. **Check mapper support**: See our [mappers page](/mappers/) for compatibility

#### Audio Problems
1. **Check audio settings**: Ensure audio is enabled in config
2. **Try different sample rate**: 44100 Hz usually works best
3. **Increase buffer size**: Helps with crackling audio

#### Performance Issues
1. **Disable VSync**: Can improve performance on some systems
2. **Use frame skip**: Skip 1-2 frames if needed
3. **Close other applications**: Free up system resources

#### Graphics Issues
1. **Try different renderer**: Switch between OpenGL and software
2. **Update graphics drivers**: Ensure latest drivers are installed
3. **Check scale setting**: Very high scales may cause issues

### Error Messages

#### "Unsupported mapper X"
The ROM uses a mapper not yet implemented in Nestor. Check our [mappers page](/mappers/) for current support status.

#### "Failed to initialize audio"
Audio system couldn't be initialized. Check that:
- Audio device isn't being used by another application
- Audio drivers are properly installed
- Try different audio settings

#### "SDL initialization failed"
System libraries couldn't be initialized. On Linux, ensure all SDL2 dependencies are installed.

### Getting Help

If you encounter issues not covered here:

1. Check [GitHub Issues](https://github.com/arl/nestor/issues) for known problems
2. Search for similar issues or create a new one
3. Include your system information, Nestor version, and error messages
4. Provide the ROM information if it's a game-specific issue

## Building from Source

### Prerequisites

- **Go 1.24+** 
- **C compiler** (gcc, clang, or Visual Studio)
- **SDL2 development libraries**
- **GTK+3 development libraries** (Linux/Windows only)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/arl/nestor.git
cd nestor

# Build the project
go build

# Run tests (optional)
go test ./...

# Install (optional)
go install
```

### Cross-Compilation

```bash
# For Windows from Linux/macOS
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build

# For macOS from Linux (requires osxcross)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build
```

## System Requirements

### Minimum Requirements
- **CPU**: 1 GHz processor
- **RAM**: 512 MB
- **Graphics**: Any graphics card with OpenGL 2.1+ support
- **Storage**: 50 MB for installation
- **OS**: Windows 7+, macOS 10.12+, Linux (any modern distribution)

### Recommended Requirements
- **CPU**: 2+ GHz multi-core processor
- **RAM**: 2+ GB
- **Graphics**: Dedicated graphics card
- **Storage**: 100+ MB for ROMs and save files
- **OS**: Windows 10+, macOS 10.15+, Ubuntu 20.04+ or equivalent

### Performance Notes
- **Single-threaded**: Nestor primarily uses one CPU core for emulation
- **GPU acceleration**: OpenGL renderer provides better performance
- **Memory usage**: Typically 20-50 MB per running instance
- **Input lag**: Frame run-ahead feature minimizes input delay

---

## Advanced Usage

### Save Data Management

Nestor automatically handles save data for games that support it:

- **Battery-backed saves**: Stored as `.sav` files next to ROMs
- **Auto-save**: Save data is written when game saves
- **Manual backup**: Copy `.sav` files to backup save progress

### Performance Tuning

For optimal performance:

```bash
# Recommended settings for modern systems
nestor run game.nes \
  --scale 3 \
  --renderer opengl \
  --vsync \
  --audio-sample-rate 44100
```

For slower systems:
```bash
# Performance-focused settings
nestor run game.nes \
  --scale 1 \
  --renderer software \
  --no-vsync \
  --frame-skip 1
```

### Development and Testing

Nestor includes features useful for developers:

```bash
# ROM analysis
nestor rominfo game.nes

# Debug mode (if compiled with debug flags)
nestor run game.nes --debug

# Screenshot capture
nestor capture game.nes --frame 1000 --output test.png
```