package hw

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/veandco/go-sdl2/sdl"
)

func TestInputCodeMarshalRoundTrip(t *testing.T) {
	tests := []struct {
		text string
		code *InputCode // nil for unmarsal errors
	}{
		{"", &InputCode{Type: UnsetController}},
		{"key W", &InputCode{Type: Keyboard, Scancode: sdl.SCANCODE_W}},
		{"key Up", &InputCode{Type: Keyboard, Scancode: sdl.SCANCODE_UP}},
		{"key Return", &InputCode{Type: Keyboard, Scancode: sdl.SCANCODE_RETURN}},
		{"joybtn a 030000004c050000cc0900", &InputCode{Type: ControllerButton, CtrlButton: sdl.CONTROLLER_BUTTON_A, CtrlGUID: "030000004c050000cc0900"}},
		{"joybtn x 030000004c050000cc0900", &InputCode{Type: ControllerButton, CtrlButton: sdl.CONTROLLER_BUTTON_X, CtrlGUID: "030000004c050000cc0900"}},
		{"joyaxis righttrigger+ 030000004c050000cc1212", &InputCode{Type: ControllerAxis, CtrlAxis: sdl.CONTROLLER_AXIS_TRIGGERRIGHT, CtrlAxisDir: 1, CtrlGUID: "030000004c050000cc1212"}},
		{"joyaxis lefttrigger- 123400004c050000cc1212", &InputCode{Type: ControllerAxis, CtrlAxis: sdl.CONTROLLER_AXIS_TRIGGERLEFT, CtrlAxisDir: -1, CtrlGUID: "123400004c050000cc1212"}},

		// unmarsal errors
		{"key   ", nil},
		{"joybtn foobar+ someguid", nil},
		{"foocode Return", nil},
		{"joybtn a", nil},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var code InputCode
			if err := code.UnmarshalText([]byte(tt.text)); err != nil {
				if tt.code != nil {
					t.Fatalf("UnmarshalText(%q) error: %v", tt.text, err)
				} else {
					t.Log("UnmarshalText error:", err)
					return
				}
			}

			if diff := cmp.Diff(*tt.code, code); diff != "" {
				t.Fatalf("UnmarshalText(%q) mismatch (-want +got):\n%s", tt.text, diff)
			}

			text, err := code.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.text, string(text)); diff != "" {
				t.Fatalf("UnmarshalText(%q) mismatch (-want +got):\n%s", tt.text, diff)
			}
		})
	}
}
