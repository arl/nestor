package dialog

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func (b *MsgBuilder) yesNo() bool {
	r, _ := _MessageBox(0, b.Msg, firstOf(b.Dlg.Title, "Confirm?"), _MB_YESNO)
	return r == _IDYES
}

func (b *MsgBuilder) info() {
	_, _ = _MessageBox(0, b.Msg, firstOf(b.Dlg.Title, "Information"), _MB_OK|_MB_ICONINFORMATION)
}

func (b *MsgBuilder) error() {
	_, _ = _MessageBox(0, b.Msg, firstOf(b.Dlg.Title, "Error"), _MB_OK|_MB_ICONERROR)
}

func (b *FileBuilder) load() (string, error) {
	return showFileDialog(fileDialogConfig{
		clsid:     &_CLSID_FileOpenDialog,
		options:   _FOS_FILEMUSTEXIST | _FOS_FORCEFILESYSTEM,
		title:     b.Dlg.Title,
		startDir:  b.StartDir,
		startFile: b.StartFile,
		filters:   b.Filters,
	})
}

func (b *FileBuilder) save() (string, error) {
	return showFileDialog(fileDialogConfig{
		clsid:     &_CLSID_FileSaveDialog,
		options:   _FOS_OVERWRITEPROMPT | _FOS_FORCEFILESYSTEM,
		title:     b.Dlg.Title,
		startDir:  b.StartDir,
		startFile: b.StartFile,
		filters:   b.Filters,
	})
}

func (b *DirectoryBuilder) browse() (string, error) {
	return showFileDialog(fileDialogConfig{
		clsid:    &_CLSID_FileOpenDialog,
		options:  _FOS_PICKFOLDERS | _FOS_FORCEFILESYSTEM,
		title:    b.Dlg.Title,
		startDir: b.StartDir,
	})
}

type fileDialogConfig struct {
	clsid     *windows.GUID
	options   uint32
	title     string
	startDir  string
	startFile string
	filters   []FileFilter
}

func showFileDialog(cfg fileDialogConfig) (string, error) {
	// COM apartments are per-thread, so the goroutine must stay on the same OS
	// thread between CoInitializeEx and CoUninitialize.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	switch hr := _CoInitializeEx(_COINIT_APARTMENTTHREADED | _COINIT_DISABLEOLE1DDE); hr {
	case _S_OK, _S_FALSE:
		defer _CoUninitialize()
	case _RPC_E_CHANGED_MODE:
		// COM is already initialized on this thread with a different concurrency
		// model. The dialog still works under the existing apartment, so proceed
		// without taking ownership of the initialization.
	default:
		return "", fmt.Errorf("dialog: CoInitializeEx failed: %w", hr)
	}

	dlg, err := _CoCreateFileDialog(cfg.clsid)
	if err != nil {
		return "", err
	}
	defer dlg.Release()

	opts, err := dlg.GetOptions()
	if err != nil {
		return "", err
	}
	if err := dlg.SetOptions(opts | cfg.options); err != nil {
		return "", err
	}

	if cfg.title != "" {
		if err := dlg.SetTitle(utf16Ptr(cfg.title)); err != nil {
			return "", err
		}
	}

	if cfg.startDir != "" {
		// A start directory that cannot be resolved should not prevent the dialog
		// from opening, so any failure here is ignored and the default is used.
		if item, err := _SHCreateItemFromParsingName(cfg.startDir); err == nil {
			_ = dlg.SetFolder(item)
			item.Release()
		}
	}

	if cfg.startFile != "" {
		if err := dlg.SetFileName(utf16Ptr(cfg.startFile)); err != nil {
			return "", err
		}
	}

	if err := dlg.SetFileTypes(fileFilterSpecs(cfg.filters)); err != nil {
		return "", err
	}

	// Use GetForegroundWindow to get the current window.
	// GetActiveWindow returns the window that belongs to the current thread.
	// TODO: Use GetActiveWindow for predictable behavior.
	if err := dlg.Show(_GetForegroundWindow()); err != nil {
		if errors.Is(err, _ERROR_CANCELLED) {
			return "", ErrCancelled
		}
		return "", err
	}

	item, err := dlg.GetResult()
	if err != nil {
		return "", err
	}
	defer item.Release()

	return item.GetDisplayName(_SIGDN_FILESYSPATH)
}

func fileFilterSpecs(filters []FileFilter) []_COMDLG_FILTERSPEC {
	var specs []_COMDLG_FILTERSPEC
	for _, f := range filters {
		patterns := make([]string, len(f.Extensions))
		for i, ext := range f.Extensions {
			patterns[i] = "*." + ext
		}
		specs = append(specs, _COMDLG_FILTERSPEC{
			Name: utf16Ptr(f.Desc),
			Spec: utf16Ptr(strings.Join(patterns, ";")),
		})
	}
	return specs
}

func utf16Ptr(s string) *uint16 {
	p, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(fmt.Sprintf("dialog: UTF16PtrFromString failed: %v", err))
	}
	return p
}
