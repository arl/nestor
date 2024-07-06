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

func decompressTestRoms(zipFile, dest string) error {
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

func downloadTestRoms(dest string) error {
	const url = `https://github.com/christopherpow/nes-test-roms/archive/refs/heads/master.zip`
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmpf, err := os.CreateTemp("", "nes-test-roms-*-.zip")
	if err != nil {
		return err
	}
	defer tmpf.Close()

	if _, err := io.Copy(tmpf, resp.Body); err != nil {
		return err
	}

	log.Println("nes test roms downloaded")

	if err := decompressTestRoms(tmpf.Name(), dest); err != nil {
		return fmt.Errorf("failed to decompress test roms: %s", err)
	}

	log.Println("nes test roms decompressed into", dest)
	return nil
}

// RomsPath returns the path to the 'nes-test-roms' directory. If this path is
// not found, test roms are downloaded and moved to the expected path.
func RomsPath(tb testing.TB) string {
	return sync.OnceValue(func() string {
		_, b, _, _ := runtime.Caller(0)
		testsDir := filepath.Dir(b)
		romsDir := filepath.Join(testsDir, "nes-test-roms")

		if _, err := os.Stat(romsDir); errors.Is(err, fs.ErrNotExist) {
			tb.Log("nes-test-roms directory not found, downloading it...")
			if err := downloadTestRoms(testsDir); err != nil {
				tb.Fatalf("failed to download test roms: %s", err)
			}
			tb.Log("Test roms downloaded in", romsDir)
		}
		return romsDir
	})()
}

func downloadTomHarteProcTests(dest string) error {
	const urlfmt = `https://raw.githubusercontent.com/SingleStepTests/65x02/main/nes6502/v1/%s.json`

	tempdir, err := os.MkdirTemp("", "tom.harte.processor.tests.*")
	if err != nil {
		return err
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

			log.Println("downloaded", url, "to", f.Name())
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to download all files: %s", err)
	}

	if err := os.Rename(tempdir, dest); err != nil {
		return err
	}

	log.Println("renaming", tempdir, "to", dest)
	return nil
}

// TomHarteProcTestsPath returns the path that is expected to contain all 256
// Tom Harte 6502 processor tests. If this path is not found, the expected
// content is downloaded and moved to the expected path.
func TomHarteProcTestsPath(tb testing.TB) string {
	return sync.OnceValue(func() string {
		_, b, _, _ := runtime.Caller(0)
		testsDir := filepath.Join(filepath.Dir(b), "tomharte.processor.tests")

		if _, err := os.Stat(testsDir); errors.Is(err, fs.ErrNotExist) {
			tb.Log("tomharte.processor.tests directory not found, downloading it...")
			if err := downloadTomHarteProcTests(testsDir); err != nil {
				tb.Fatalf("failed to download tom harte proc tests: %s", err)
			}
			tb.Log("Tom Harte Processor Tests downloaded in", testsDir)
		}
		return testsDir
	})()
}
