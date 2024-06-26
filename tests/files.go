package tests

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
)

func decompress(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fname := strings.Replace(f.Name, "nes-test-roms-master", "nes-test-roms", 1)
		fpath := filepath.Join(dest, fname)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	log.Println("decompressed", len(r.File), "files")
	return nil
}

func downloadTestRoms(tb testing.TB, dest string) {
	const url = `https://github.com/christopherpow/nes-test-roms/archive/refs/heads/master.zip`
	resp, err := http.Get(url)
	if err != nil {
		tb.Fatal(err)
	}
	defer resp.Body.Close()

	tmpf, err := os.CreateTemp("", "nes-test-roms-*-.zip")
	if err != nil {
		tb.Fatal(err)
	}
	defer tmpf.Close()

	if _, err := io.Copy(tmpf, resp.Body); err != nil {
		tb.Fatal(err)
	}

	if err := decompress(tmpf.Name(), dest); err != nil {
		tb.Fatalf("failed to decompress test roms: %s", err)
	}
}

func RomsPath(tb testing.TB) string {
	return sync.OnceValue(func() string {
		_, b, _, _ := runtime.Caller(0)
		testsDir := filepath.Dir(b)
		romsDir := filepath.Join(testsDir, "nes-test-roms")

		if _, err := os.Stat(romsDir); errors.Is(err, fs.ErrNotExist) {
			tb.Log("nes-test-roms directory not found, downloading it...")
			downloadTestRoms(tb, testsDir)
			tb.Log("Test roms downloaded in", romsDir)
		}

		return romsDir
	})()
}

// download all 256 (one per opcode) Tom harte 6502 test files into dest dir.
func downloadTomHarteProcTests(tb testing.TB, dest string) {
	const urlfmt = `https://raw.githubusercontent.com/SingleStepTests/65x02/main/nes6502/v1/%s.json`

	tempdir, err := os.MkdirTemp("", "tom.harte.processor.tests.*")
	if err != nil {
		tb.Fatal(err)
	}

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for opcode := range 256 {
		opstr := fmt.Sprintf("%02x", opcode)
		url := fmt.Sprintf(urlfmt, opstr)

		g.Go(func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			f, err := os.Create(filepath.Join(tempdir, opstr+".json"))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, resp.Body); err != nil {
				return err
			}

			tb.Log("downloaded", url, "to", f.Name())
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		tb.Fatalf("failed to download all files: %s", err)
	}

	if err := os.Rename(tempdir, dest); err != nil {
		tb.Fatal(err)
	}

	tb.Log("renaming", tempdir, "to", dest)
}

func TomHarteProcTestsPath(tb testing.TB) string {
	return sync.OnceValue(func() string {
		_, b, _, _ := runtime.Caller(0)
		testsDir := filepath.Join(filepath.Dir(b), "tomharte.processor.tests")

		if _, err := os.Stat(testsDir); errors.Is(err, fs.ErrNotExist) {
			tb.Log("tomharte.processor.tests directory not found, downloading it...")
			downloadTomHarteProcTests(tb, testsDir)
			tb.Log("Tom Harte Processor Tests downloaded in", testsDir)
		}

		return testsDir
	})()
}
