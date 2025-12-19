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

// IsModifier returns true if the key is a modifier key (Ctrl, Shift, Alt, Meta)
func IsModifier(key ebiten.Key) bool {
	switch key {
	case ebiten.KeyControl, ebiten.KeyControlLeft, ebiten.KeyControlRight,
		ebiten.KeyShift, ebiten.KeyShiftLeft, ebiten.KeyShiftRight,
		ebiten.KeyAlt, ebiten.KeyAltLeft, ebiten.KeyAltRight,
		ebiten.KeyMeta, ebiten.KeyMetaLeft, ebiten.KeyMetaRight:
		return true
	}
	return false
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
			var key ebiten.Key
			if err := key.UnmarshalText([]byte(part)); err != nil {
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

	keyStr := strings.ToLower(k.Key.String())
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

// CaptureKeyCombo captures a key combination from just-pressed keys.
// It filters out modifier keys and returns a KeyCombo with the main key and modifier states.
// Returns ok=false if no non-modifier key was just pressed.
func CaptureKeyCombo() (combo KeyCombo, ok bool) {
	keys := inpututil.AppendJustPressedKeys(nil)
	if len(keys) == 0 {
		return KeyCombo{}, false
	}

	// Find the first non-modifier key
	var mainKey ebiten.Key
	found := false
	for _, key := range keys {
		if !IsModifier(key) {
			mainKey = key
			found = true
			break
		}
	}

	if !found {
		return KeyCombo{}, false
	}

	combo.Key = mainKey
	combo.Ctrl = ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	combo.Shift = ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	combo.Alt = ebiten.IsKeyPressed(ebiten.KeyAlt) || ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)

	return combo, true
}
