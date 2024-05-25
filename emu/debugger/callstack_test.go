package debugger

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCallStack(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		var cstack callStack
		cstack.push(0xC7C2, 0xC7E7, 0xC7C5, sffNone)
		cstack.push(0xC801, 0xCBAE, 0xC804, sffNone)

		fi := cstack.build(0xF099)
		want := []frameInfo{
			{"CBAE", "$F099"},
			{"C7E7", "$C801"},
			{"[bottom of stack]", "$C7C2"},
		}
		if diff := cmp.Diff(fi, want); diff != "" {
			t.Fatalf("callstack differs (-want +got):\n%s", diff)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var cstack callStack
		fi := cstack.build(0xF099)
		want := []frameInfo{
			{"[bottom of stack]", "$F099"},
		}
		if diff := cmp.Diff(fi, want); diff != "" {
			t.Fatalf("callstack differs (-want +got):\n%s", diff)
		}
	})
}
