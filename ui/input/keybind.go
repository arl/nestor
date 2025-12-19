package input

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// KeyCombo represents a key with optional modifiers
type KeyCombo struct {
	Key   ebiten.Key
	Ctrl  bool
	Shift bool
	Alt   bool
}

var keyNames = map[string]ebiten.Key{
	"a": ebiten.KeyA, "b": ebiten.KeyB, "c": ebiten.KeyC, "d": ebiten.KeyD,
	"e": ebiten.KeyE, "f": ebiten.KeyF, "g": ebiten.KeyG, "h": ebiten.KeyH,
	"i": ebiten.KeyI, "j": ebiten.KeyJ, "k": ebiten.KeyK, "l": ebiten.KeyL,
	"m": ebiten.KeyM, "n": ebiten.KeyN, "o": ebiten.KeyO, "p": ebiten.KeyP,
	"q": ebiten.KeyQ, "r": ebiten.KeyR, "s": ebiten.KeyS, "t": ebiten.KeyT,
	"u": ebiten.KeyU, "v": ebiten.KeyV, "w": ebiten.KeyW, "x": ebiten.KeyX,
	"y": ebiten.KeyY, "z": ebiten.KeyZ,
	"0": ebiten.Key0, "1": ebiten.Key1, "2": ebiten.Key2, "3": ebiten.Key3,
	"4": ebiten.Key4, "5": ebiten.Key5, "6": ebiten.Key6, "7": ebiten.Key7,
	"8": ebiten.Key8, "9": ebiten.Key9,
	"f1": ebiten.KeyF1, "f2": ebiten.KeyF2, "f3": ebiten.KeyF3, "f4": ebiten.KeyF4,
	"f5": ebiten.KeyF5, "f6": ebiten.KeyF6, "f7": ebiten.KeyF7, "f8": ebiten.KeyF8,
	"f9": ebiten.KeyF9, "f10": ebiten.KeyF10, "f11": ebiten.KeyF11, "f12": ebiten.KeyF12,
	"escape": ebiten.KeyEscape, "esc": ebiten.KeyEscape,
	"enter": ebiten.KeyEnter, "return": ebiten.KeyEnter,
	"space": ebiten.KeySpace,
	"tab": ebiten.KeyTab,
	"backspace": ebiten.KeyBackspace,
	"delete": ebiten.KeyDelete,
	"insert": ebiten.KeyInsert,
	"home": ebiten.KeyHome,
	"end": ebiten.KeyEnd,
	"pageup": ebiten.KeyPageUp,
	"pagedown": ebiten.KeyPageDown,
	"up": ebiten.KeyUp, "down": ebiten.KeyDown, "left": ebiten.KeyLeft, "right": ebiten.KeyRight,
	"minus": ebiten.KeyMinus, "-": ebiten.KeyMinus,
	"equal": ebiten.KeyEqual, "=": ebiten.KeyEqual,
	"comma": ebiten.KeyComma, ",": ebiten.KeyComma,
	"period": ebiten.KeyPeriod, ".": ebiten.KeyPeriod,
	"slash": ebiten.KeySlash, "/": ebiten.KeySlash,
	"backslash": ebiten.KeyBackslash, "\\": ebiten.KeyBackslash,
	"semicolon": ebiten.KeySemicolon, ";": ebiten.KeySemicolon,
	"quote": ebiten.KeyQuote, "'": ebiten.KeyQuote,
	"backquote": ebiten.KeyBackquote, "`": ebiten.KeyBackquote,
	"leftbracket": ebiten.KeyLeftBracket, "[": ebiten.KeyLeftBracket,
	"rightbracket": ebiten.KeyRightBracket, "]": ebiten.KeyRightBracket,
}

// Parse parses strings like "ctrl+o", "shift+f1", "f11"
func Parse(s string) (KeyCombo, error) {
	var combo KeyCombo
	parts := strings.Split(strings.ToLower(s), "+")

	for i, part := range parts {
		part = strings.TrimSpace(part)
		switch part {
		case "ctrl", "control":
			combo.Ctrl = true
		case "shift":
			combo.Shift = true
		case "alt":
			combo.Alt = true
		default:
			if i != len(parts)-1 {
				return KeyCombo{}, fmt.Errorf("invalid modifier %q in key combination %q", part, s)
			}
			key, ok := keyNames[part]
			if !ok {
				return KeyCombo{}, fmt.Errorf("unknown key %q in key combination %q", part, s)
			}
			combo.Key = key
		}
	}

	return combo, nil
}

func (k KeyCombo) String() string {
	var parts []string
	if k.Ctrl {
		parts = append(parts, "ctrl")
	}
	if k.Shift {
		parts = append(parts, "shift")
	}
	if k.Alt {
		parts = append(parts, "alt")
	}

	keyStr := ""
	for name, key := range keyNames {
		if key == k.Key && len(name) > len(keyStr) {
			keyStr = name
		}
	}
	if keyStr == "" {
		keyStr = k.Key.String()
	}
	parts = append(parts, keyStr)

	return strings.Join(parts, "+")
}

// IsJustPressed returns true if this combo was just pressed
func (k KeyCombo) IsJustPressed() bool {
	if !inpututil.IsKeyJustPressed(k.Key) {
		return false
	}

	ctrlPressed := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	shiftPressed := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	altPressed := ebiten.IsKeyPressed(ebiten.KeyAlt) || ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)

	return ctrlPressed == k.Ctrl && shiftPressed == k.Shift && altPressed == k.Alt
}

