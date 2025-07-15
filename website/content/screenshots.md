---
title: "Screenshots"
description: "See Nestor in action with these screenshots of popular NES games running on the emulator."
---

# Nestor Screenshots

Experience the visual quality and accuracy of Nestor emulation through these game screenshots. All images show games running in real-time on the emulator.

## Action & Adventure Games

<div class="screenshot-gallery">
  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/8b283d1f-9eca-49da-849f-d4c9c91f98cd" alt="Castlevania">
    <div class="screenshot-info">
      <h3>Castlevania</h3>
      <p>Classic action-adventure platformer showcasing Nestor's accurate sprite rendering and smooth scrolling.</p>
      <span class="mapper-info">Mapper: 2 (UxROM)</span>
    </div>
  </div>

  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/a59fbc21-4938-441d-81d7-1dabda65c929" alt="Contra">
    <div class="screenshot-info">
      <h3>Contra</h3>
      <p>Fast-paced run-and-gun action demonstrating Nestor's performance with intense gameplay.</p>
      <span class="mapper-info">Mapper: 2 (UxROM)</span>
    </div>
  </div>

  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/014025c9-6c7e-4f68-b351-3557c345a12e" alt="Adventures of Rad Gravity">
    <div class="screenshot-info">
      <h3>Adventures of Rad Gravity</h3>
      <p>Sci-fi platformer showcasing complex level design and detailed pixel art.</p>
      <span class="mapper-info">Mapper: 1 (MMC1)</span>
    </div>
  </div>

  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/cdb49c3e-4ac4-4dd9-94fe-ac4d91af4aff" alt="Prince of Persia">
    <div class="screenshot-info">
      <h3>Prince of Persia</h3>
      <p>Cinematic platformer with fluid character animation and atmospheric graphics.</p>
      <span class="mapper-info">Mapper: 1 (MMC1)</span>
    </div>
  </div>
</div>

## Fighting & Sports

<div class="screenshot-gallery">
  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/d7a03db0-fcf7-4e8f-a8f7-23ec0d01fae7" alt="Battletoads">
    <div class="screenshot-info">
      <h3>Battletoads</h3>
      <p>Challenging beat-em-up known for its difficulty and varied gameplay mechanics.</p>
      <span class="mapper-info">Mapper: 7 (AxROM)</span>
    </div>
  </div>

  <div class="screenshot-large">
    <img src="https://github.com/user-attachments/assets/534e5d32-7bf0-48a1-9b3e-bb580f651585" alt="Tsuppari Oozumou">
    <div class="screenshot-info">
      <h3>Tsuppari Oozumou</h3>
      <p>Sumo wrestling game demonstrating Nestor's compatibility with Japanese exclusives.</p>
      <span class="mapper-info">Mapper: 0 (NROM)</span>
    </div>
  </div>
</div>

## User Interface Screenshots

### Main Window
<div class="ui-screenshot">
  <img src="https://github.com/user-attachments/assets/2515bce2-a926-40f0-9213-2505d87f102b" alt="Main Window">
  <div class="screenshot-info">
    <h3>ROM Selection Interface</h3>
    <p>Clean, intuitive interface for browsing and selecting ROM files. Features recent games list and easy navigation.</p>
  </div>
</div>

### Emulator Window
<div class="ui-screenshot">
  <img src="https://github.com/user-attachments/assets/5b4b7e7a-b8af-4f81-83c1-2df4f1814591" alt="Emulator Window">
  <div class="screenshot-info">
    <h3>In-Game Controls</h3>
    <p>Emulator window with accompanying control panel for real-time adjustments during gameplay.</p>
  </div>
</div>

### Input Configuration
<div class="ui-screenshot">
  <img src="https://github.com/user-attachments/assets/4add9e06-1eff-4bb0-82f0-c4e2f6583e59" alt="Input Configuration">
  <div class="screenshot-info">
    <h3>Controller Setup</h3>
    <p>Comprehensive input configuration interface supporting both keyboard and gamepad controls.</p>
  </div>
</div>

## Visual Quality Features

### Authentic NES Experience
All screenshots demonstrate Nestor's commitment to accuracy:
- **Pixel-perfect rendering** with no smoothing artifacts
- **Authentic color palette** matching original hardware
- **Proper aspect ratio** preservation
- **Scanline effects** available for CRT simulation

### Performance Indicators
- **60 FPS gameplay** maintained across all tested games
- **No frame drops** during intense action sequences
- **Smooth scrolling** in all directions
- **Responsive controls** with minimal input lag

## Game Compatibility Gallery

These screenshots represent games across different mapper types, demonstrating Nestor's broad compatibility:

| Game | Mapper | Genre | Status |
|------|--------|-------|--------|
| Castlevania | UxROM (2) | Action | ✅ Perfect |
| Contra | UxROM (2) | Shooter | ✅ Perfect |
| Adventures of Rad Gravity | MMC1 (1) | Platformer | ✅ Perfect |
| Prince of Persia | MMC1 (1) | Adventure | ✅ Perfect |
| Battletoads | AxROM (7) | Beat-em-up | ✅ Perfect |
| Tsuppari Oozumou | NROM (0) | Sports | ✅ Perfect |

## Testing Methodology

Screenshots were captured using:
- **Native resolution** output
- **Standard NTSC timing**
- **Default video settings**
- **No post-processing filters**

This ensures what you see represents the actual emulation quality without enhancement or modification.

## Submit Your Screenshots

Have great screenshots of games running in Nestor? We'd love to feature them! Please:
1. Use high-quality settings
2. Capture interesting gameplay moments
3. Include game name and mapper information
4. Submit via our [GitHub repository]({{ .Site.Params.github }})

<style>
.screenshot-gallery {
  display: grid;
  gap: 2rem;
  margin: 2rem 0;
}

.screenshot-large {
  background: var(--surface);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
}

.screenshot-large img {
  width: 100%;
  height: auto;
  display: block;
}

.screenshot-info {
  padding: 1.5rem;
}

.screenshot-info h3 {
  margin: 0 0 0.5rem 0;
  color: var(--primary-color);
}

.screenshot-info p {
  margin: 0 0 1rem 0;
  color: var(--text-secondary);
}

.mapper-info {
  background: var(--primary-color);
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 600;
}

.ui-screenshot {
  background: var(--surface);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  margin: 2rem 0;
}

.ui-screenshot img {
  width: 100%;
  height: auto;
  display: block;
}

.ui-screenshot .screenshot-info {
  padding: 1.5rem;
}

@media (min-width: 768px) {
  .screenshot-gallery {
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  }
}
</style>