// SPDX-License-Identifier: ISC
// SPDX-FileCopyrightText: 2026 Hajime Hoshi

package dialog

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	_MB_OK              = 0x00000000
	_MB_YESNO           = 0x00000004
	_MB_ICONHAND        = 0x00000010
	_MB_ICONASTERISK    = 0x00000040
	_MB_ICONERROR       = _MB_ICONHAND
	_MB_ICONINFORMATION = _MB_ICONASTERISK

	_IDYES = 6

	_CLSCTX_INPROC_SERVER = 0x1

	_COINIT_APARTMENTTHREADED = 0x2
	_COINIT_DISABLEOLE1DDE    = 0x4

	_FOS_OVERWRITEPROMPT = 0x00000002
	_FOS_PICKFOLDERS     = 0x00000020
	_FOS_FORCEFILESYSTEM = 0x00000040
	_FOS_FILEMUSTEXIST   = 0x00001000

	_SIGDN_FILESYSPATH = 0x80058000
)

type _HRESULT uint32

const (
	_S_OK               _HRESULT = 0x00000000
	_S_FALSE            _HRESULT = 0x00000001
	_RPC_E_CHANGED_MODE _HRESULT = 0x80010106
	_ERROR_CANCELLED    _HRESULT = 0x800704C7
)

func (hr _HRESULT) failed() bool {
	return hr&0x80000000 != 0
}

func (hr _HRESULT) Error() string {
	return fmt.Sprintf("HRESULT 0x%08x", uint32(hr))
}

var (
	_CLSID_FileOpenDialog = windows.GUID{Data1: 0xDC1C5A9C, Data2: 0xE88A, Data3: 0x4DDE, Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	_CLSID_FileSaveDialog = windows.GUID{Data1: 0xC0B4E2F3, Data2: 0xBA21, Data3: 0x4773, Data4: [8]byte{0x8D, 0xBA, 0x33, 0x5E, 0xC9, 0x46, 0xEB, 0x8B}}
	_IID_IFileDialog      = windows.GUID{Data1: 0x42F85136, Data2: 0xDB7E, Data3: 0x439C, Data4: [8]byte{0x85, 0xF1, 0xE4, 0x07, 0x5D, 0x13, 0x5F, 0xC8}}
	_IID_IShellItem       = windows.GUID{Data1: 0x43826D1E, Data2: 0xE718, Data3: 0x42EE, Data4: [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE}}
)

type _COMDLG_FILTERSPEC struct {
	Name *uint16
	Spec *uint16
}

type _IFileDialogVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Show           uintptr

	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
}

type _IFileDialog struct {
	vtbl *_IFileDialogVtbl
}

type _IShellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

type _IShellItem struct {
	vtbl *_IShellItemVtbl
}

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
)

var (
	ole32 = windows.NewLazySystemDLL("ole32.dll")

	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")
)

var (
	shell32 = windows.NewLazySystemDLL("shell32.dll")

	procSHCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")
)

func _GetForegroundWindow() windows.HWND {
	r, _, _ := procGetForegroundWindow.Call()
	return windows.HWND(r)
}

func _MessageBox(hwnd windows.HWND, text, caption string, flags uint32) (int32, error) {
	textPtr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return 0, err
	}
	captionPtr, err := windows.UTF16PtrFromString(caption)
	if err != nil {
		return 0, err
	}
	r, _, err := procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(flags),
	)
	if r == 0 {
		if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
			return 0, fmt.Errorf("dialog: MessageBoxW failed: %w", err)
		}
		return 0, fmt.Errorf("dialog: MessageBoxW failed: returned 0")
	}
	return int32(r), nil
}

func _CoInitializeEx(coInit uint32) _HRESULT {
	r, _, _ := procCoInitializeEx.Call(0, uintptr(coInit))
	return _HRESULT(uint32(r))
}

func _CoUninitialize() {
	_, _, _ = procCoUninitialize.Call()
}

func _CoTaskMemFree(p unsafe.Pointer) {
	_, _, _ = procCoTaskMemFree.Call(uintptr(p))
}

func _CoCreateFileDialog(clsid *windows.GUID) (*_IFileDialog, error) {
	var dlg *_IFileDialog
	r, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(clsid)),
		0,
		_CLSCTX_INPROC_SERVER,
		uintptr(unsafe.Pointer(&_IID_IFileDialog)),
		uintptr(unsafe.Pointer(&dlg)),
	)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return nil, fmt.Errorf("dialog: CoCreateInstance failed: %w", hr)
	}
	return dlg, nil
}

func _SHCreateItemFromParsingName(path string) (*_IShellItem, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var item *_IShellItem
	r, _, _ := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		uintptr(unsafe.Pointer(&_IID_IShellItem)),
		uintptr(unsafe.Pointer(&item)),
	)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return nil, fmt.Errorf("dialog: SHCreateItemFromParsingName failed: %w", hr)
	}
	return item, nil
}

func (d *_IFileDialog) Release() {
	_, _, _ = syscall.SyscallN(d.vtbl.Release, uintptr(unsafe.Pointer(d)))
}

func (d *_IFileDialog) Show(owner windows.HWND) error {
	r, _, _ := syscall.SyscallN(d.vtbl.Show, uintptr(unsafe.Pointer(d)), uintptr(owner))
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::Show failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) GetOptions() (uint32, error) {
	var opts uint32
	r, _, _ := syscall.SyscallN(d.vtbl.GetOptions, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&opts)))
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return 0, fmt.Errorf("dialog: IFileDialog::GetOptions failed: %w", hr)
	}
	return opts, nil
}

func (d *_IFileDialog) SetOptions(opts uint32) error {
	r, _, _ := syscall.SyscallN(d.vtbl.SetOptions, uintptr(unsafe.Pointer(d)), uintptr(opts))
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::SetOptions failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) SetFolder(item *_IShellItem) error {
	r, _, _ := syscall.SyscallN(d.vtbl.SetFolder, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(item)))
	runtime.KeepAlive(item)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::SetFolder failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) SetFileName(name *uint16) error {
	r, _, _ := syscall.SyscallN(d.vtbl.SetFileName, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(name)))
	runtime.KeepAlive(name)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::SetFileName failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) SetTitle(title *uint16) error {
	r, _, _ := syscall.SyscallN(d.vtbl.SetTitle, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(title)))
	runtime.KeepAlive(title)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::SetTitle failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) SetFileTypes(specs []_COMDLG_FILTERSPEC) error {
	if len(specs) == 0 {
		return nil
	}
	r, _, _ := syscall.SyscallN(d.vtbl.SetFileTypes, uintptr(unsafe.Pointer(d)), uintptr(len(specs)), uintptr(unsafe.Pointer(&specs[0])))
	runtime.KeepAlive(specs)
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return fmt.Errorf("dialog: IFileDialog::SetFileTypes failed: %w", hr)
	}
	return nil
}

func (d *_IFileDialog) GetResult() (*_IShellItem, error) {
	var item *_IShellItem
	r, _, _ := syscall.SyscallN(d.vtbl.GetResult, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&item)))
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return nil, fmt.Errorf("dialog: IFileDialog::GetResult failed: %w", hr)
	}
	return item, nil
}

func (item *_IShellItem) Release() {
	_, _, _ = syscall.SyscallN(item.vtbl.Release, uintptr(unsafe.Pointer(item)))
}

func (item *_IShellItem) GetDisplayName(sigdn uint32) (string, error) {
	var p *uint16
	r, _, _ := syscall.SyscallN(item.vtbl.GetDisplayName, uintptr(unsafe.Pointer(item)), uintptr(sigdn), uintptr(unsafe.Pointer(&p)))
	if hr := _HRESULT(uint32(r)); hr.failed() {
		return "", fmt.Errorf("dialog: IShellItem::GetDisplayName failed: %w", hr)
	}
	defer _CoTaskMemFree(unsafe.Pointer(p))
	return windows.UTF16PtrToString(p), nil
}
