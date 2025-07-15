---
title: "Mappers"
description: "Complete guide to NES mappers supported by Nestor, including implementation status and compatible games."
---

# NES Mappers in Nestor

## What are Mappers?

NES cartridges contain various circuits and hardware configurations that extend the system's capabilities beyond its basic limitations. These configurations are commonly called "mappers" because they control how the NES CPU and PPU map memory addresses to the cartridge's ROM and RAM.

Mappers were designed to:
- Extend available memory beyond the NES's 32KB address space
- Add extra RAM to cartridges
- Provide additional sound channels
- Enable complex graphics effects
- Support larger games and more sophisticated features

Each mapper has a unique number assigned by the iNES ROM format specification.

## Supported Mappers

### Mapper 0: NROM ✅
**Status**: Fully Implemented  
**Complexity**: Simple  

The most basic mapper with no bank switching. Games are limited to 32KB of program ROM and 8KB of character ROM.

**Popular Games**:
- Super Mario Bros.
- Donkey Kong
- Popeye
- Ice Climber
- Balloon Fight

**Technical Details**:
- No bank switching
- Fixed memory layout
- 16KB or 32KB PRG ROM
- 8KB CHR ROM
- Optional 2KB work RAM

---

### Mapper 1: MMC1 ✅
**Status**: Fully Implemented  
**Complexity**: Medium  

Nintendo's first Memory Management Controller, supporting larger games with bank switching capabilities.

**Popular Games**:
- The Legend of Zelda
- Metroid
- Kid Icarus
- Zelda II: The Adventure of Link
- Final Fantasy

**Technical Details**:
- 5-bit serial register interface
- Up to 512KB PRG ROM support
- Up to 128KB CHR ROM support
- Multiple PRG/CHR banking modes
- Optional 8KB PRG RAM with battery backup

---

### Mapper 2: UxROM ✅
**Status**: Fully Implemented  
**Complexity**: Simple  

Simple PRG ROM banking mapper commonly used by Konami and others.

**Popular Games**:
- Mega Man
- Contra
- Castlevania
- Duck Tales
- Gradius

**Technical Details**:
- 16KB PRG ROM banking
- Fixed 16KB PRG ROM bank at $C000-$FFFF
- No CHR ROM banking (uses CHR RAM)
- Up to 256KB PRG ROM support

---

### Mapper 3: CNROM ✅
**Status**: Fully Implemented  
**Complexity**: Simple  

Basic CHR ROM banking mapper for games needing more graphics data.

**Popular Games**:
- Solomon's Key
- Arkanoid
- Cybernoid
- Pipe Dream

**Technical Details**:
- Simple CHR ROM banking
- No PRG ROM banking
- Up to 32KB CHR ROM support
- 8KB CHR banks

---

### Mapper 7: AxROM ✅
**Status**: Fully Implemented  
**Complexity**: Simple  

Single-screen mirroring mapper with 32KB PRG banking.

**Popular Games**:
- Battletoads
- Wizards & Warriors
- Marble Madness
- R.C. Pro-Am

**Technical Details**:
- 32KB PRG ROM banking
- Single-screen mirroring control
- Up to 256KB PRG ROM support
- Uses CHR RAM

---

### Mapper 66: GxROM ✅
**Status**: Fully Implemented  
**Complexity**: Simple  

Combined PRG and CHR ROM banking used by Nintendo for certain cartridges.

**Popular Games**:
- Super Mario Bros. + Duck Hunt
- SMB + Duck Hunt + World Class Track Meet

**Technical Details**:
- 32KB PRG ROM banking
- 8KB CHR ROM banking
- Simple register interface
- Used in multicart releases

---

## Planned Mappers

### Mapper 4: MMC3 🔄
**Status**: Planned  
**Complexity**: High  

Nintendo's most sophisticated first-party mapper with advanced features.

**Target Games**:
- Super Mario Bros. 3
- Mega Man 2-6
- Kirby's Adventure
- Super Mario Bros. 2

**Planned Features**:
- 8KB and 16KB PRG banking
- 1KB and 2KB CHR banking
- Scanline counter IRQ
- Multiple mirroring modes

---

### Mapper 10: MMC5 🔄
**Status**: Planned  
**Complexity**: Very High  

Nintendo's most advanced mapper with numerous enhancements.

**Target Games**:
- Castlevania III: Dracula's Curse
- Just Breed
- Koei games

**Planned Features**:
- Advanced PRG/CHR banking
- Extra sound channels
- Extended attributes
- IRQ generation
- Large ROM support

---

## Implementation Priority

Our mapper implementation follows this priority order:

1. **High Priority**: Mappers 0, 1, 2, 3 (✅ Complete)
2. **Medium Priority**: Mappers 7, 66 (✅ Complete)
3. **Future Priority**: Mappers 4, 5, 9, 10, 11, 71

## Compatibility Statistics

| Mapper Type | Games Supported | Completion Rate |
|-------------|----------------|-----------------|
| Basic (0, 2, 3, 7, 66) | ~200+ games | 100% |
| MMC1 (1) | ~150+ games | 100% |
| MMC3 (4) | ~300+ games | Planned |
| Others | ~100+ games | Future |

**Total Current Compatibility**: Approximately 350+ games fully supported

## Testing and Verification

Each mapper implementation is thoroughly tested with:
- Popular commercial games
- Homebrew test ROMs
- Edge case scenarios
- Timing-sensitive operations
- Memory access patterns

## Technical Implementation Notes

Nestor's mapper implementations prioritize:
- **Accuracy**: Cycle-precise timing where required
- **Compatibility**: Support for all known ROM variants
- **Performance**: Efficient bank switching operations
- **Maintainability**: Clean, well-documented code

## Contributing Mapper Support

We welcome contributions for additional mapper support! Please see our [development guide]({{ .Site.Params.github }}) for:
- Mapper implementation guidelines
- Testing procedures
- Documentation requirements
- Code review process