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
	"testing"
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
		tmpf.Close()
		return err
	}
	tmpf.Close()

	if err := decompress(tmpf.Name(), dest); err != nil {
		return fmt.Errorf("failed to decompress test roms: %s", err)
	}
	return nil
}

func RomsPath(tb testing.TB) string {
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
}
