package debugger

import "testing"

func TestCallStack(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var cstack callStack
		cstack.push(0xC7C2, 0xC7E7, 0xC7C5, sffNone)
		cstack.push(0xC801, 0xCBAE, 0xC804, sffNone)

		fi := cstack.build(0xF099)

		for _, f := range fi {
			t.Logf("%s", f)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var cstack callStack
		fi := cstack.build(0xF099)
		for _, f := range fi {
			t.Logf("%s", f)
		}
	})
}
