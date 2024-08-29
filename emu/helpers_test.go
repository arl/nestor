package emu

import (
	"flag"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func CompareFileWithGolden(t *testing.T, gotfile, filename string, update bool) {
	got, err := os.ReadFile(gotfile)
	if err != nil {
		t.Fatal(err)
	}
	CompareWithGolden(t, string(got), filename, update)
}

func CompareWithGolden(t *testing.T, got, filename string, update bool) {
	t.Helper()
	if update {
		writeGolden(t, filename, got)
	} else {
		want := readGolden(t, filename)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s: mismatch (-want, +got):\n%s", filename, diff)
		}
	}
}

func writeGolden(t *testing.T, name string, data string) {
	if err := os.WriteFile(name, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", name)
}

func readGolden(t *testing.T, name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
