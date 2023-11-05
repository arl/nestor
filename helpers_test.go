package main

import (
	"fmt"
	"testing"
)

/* general testing helpers */

func tcheck(tb testing.TB, err error) {
	if err == nil {
		return
	}

	tb.Helper()
	tb.Fatalf("fatal error:\n\n%s\n", err)
}

func tcheckf(tb testing.TB, err error, format string, args ...any) {
	if err == nil {
		return
	}

	tb.Helper()
	tb.Fatalf("fatal error:\n\n%s: %s\n", fmt.Sprintf(format, args...), err)
}
