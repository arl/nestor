# ui package for Nestor

## Overview

This package contains the entry point for Nestor (NES emulator written in Go) user interface, as well as reusable widgets.
User interface components are built using the Ebiten UI library: https://github.com/ebitenui/ebitenui


## Resources

UI aims to be minimalist and clean. Widgets resources (colors, fonts, etc.) are defined in `ui/resources.go`.


## Coding style

Don't add comments anywhere in the code unless code is intricated enough to require explanation, or as separators for big blocks of code, generally when defining UI containers, widgets or layouts

Name variables should be as small as possible while remaining meaningful. For instance, use `lbl` for label widgets, `cfg` for configuration objects, `btn` for buttons, etc.

When defining containers or widgets with multiple options, use one line per option to improve readability.

When defining event handlers, prefer defining them as named functions rather than anonymous functions when they are more than a couple of lines long.