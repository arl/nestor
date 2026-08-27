// SPDX-License-Identifier: ISC
// SPDX-FileCopyrightText: 2026 Hajime Hoshi

//go:build !darwin && !js && !windows

package dialog

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const (
	_GTK_MESSAGE_INFO     = 0
	_GTK_MESSAGE_QUESTION = 2
	_GTK_MESSAGE_ERROR    = 3

	_GTK_BUTTONS_OK     = 1
	_GTK_BUTTONS_YES_NO = 4

	_GTK_RESPONSE_ACCEPT = -3
	_GTK_RESPONSE_CANCEL = -6
	_GTK_RESPONSE_YES    = -8

	_GTK_FILE_CHOOSER_ACTION_OPEN          = 0
	_GTK_FILE_CHOOSER_ACTION_SAVE          = 1
	_GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER = 2
)

// gtk_message_dialog_new and gtk_file_chooser_dialog_new are variadic C
// functions. They are bound with fixed signatures matching the exact
// arguments passed, which is compatible with the SysV calling conventions
// used on Linux.
var (
	gFree func(p *byte)

	gtkDialogRun                             func(dialog uintptr) int32
	gtkEventsPending                         func() bool
	gtkFileChooserAddFilter                  func(chooser, filter uintptr)
	gtkFileChooserDialogNew                  func(title string, parent uintptr, action int32, firstButtonText string, firstResponseID int32, secondButtonText string, secondResponseID int32, terminator uintptr) uintptr
	gtkFileChooserGetFilename                func(chooser uintptr) *byte
	gtkFileChooserSetCurrentFolder           func(chooser uintptr, filename string) bool
	gtkFileChooserSetCurrentName             func(chooser uintptr, name string)
	gtkFileChooserSetDoOverwriteConfirmation func(chooser uintptr, doOverwriteConfirmation bool)
	gtkFileFilterAddPattern                  func(filter uintptr, pattern string)
	gtkFileFilterNew                         func() uintptr
	gtkFileFilterSetName                     func(filter uintptr, name string)
	gtkInitCheck                             func(argc, argv uintptr) bool
	gtkMainIteration                         func() bool
	gtkMessageDialogNew                      func(parent uintptr, flags uint32, messageType, buttons int32, format string, arg string) uintptr
	gtkWidgetDestroy                         func(widget uintptr)
	gtkWindowSetTitle                        func(window uintptr, title string)
)

var (
	gtkOnce    sync.Once
	gtkInitErr error
	gtkCalls   = make(chan func())
)

// ensureGTK loads and initializes GTK 3 on first use. It panics if the
// libraries cannot be loaded or if GTK cannot be initialized, which
// typically means no display is available.
func ensureGTK() {
	gtkOnce.Do(func() {
		if err := loadGTK(); err != nil {
			gtkInitErr = err
			return
		}
		// GTK is not thread-safe: every GTK call must happen on the same OS
		// thread. All calls are funneled to a goroutine locked to one thread.
		go func() {
			runtime.LockOSThread()
			for f := range gtkCalls {
				f()
			}
		}()
		runOnGTKThread(func() {
			if !gtkInitCheck(0, 0) {
				gtkInitErr = errors.New("dialog: gtk_init_check failed; presumably no display is available")
			}
		})
	})
	if gtkInitErr != nil {
		panic(gtkInitErr)
	}
}

func loadGTK() error {
	glib, err := openLibrary("libglib-2.0.so.0", "libglib-2.0.so")
	if err != nil {
		return fmt.Errorf("dialog: failed to load GLib: %w", err)
	}
	gtk, err := openLibrary("libgtk-3.so.0", "libgtk-3.so")
	if err != nil {
		return fmt.Errorf("dialog: failed to load GTK 3: %w", err)
	}

	purego.RegisterLibFunc(&gFree, glib, "g_free")

	purego.RegisterLibFunc(&gtkDialogRun, gtk, "gtk_dialog_run")
	purego.RegisterLibFunc(&gtkEventsPending, gtk, "gtk_events_pending")
	purego.RegisterLibFunc(&gtkFileChooserAddFilter, gtk, "gtk_file_chooser_add_filter")
	purego.RegisterLibFunc(&gtkFileChooserDialogNew, gtk, "gtk_file_chooser_dialog_new")
	purego.RegisterLibFunc(&gtkFileChooserGetFilename, gtk, "gtk_file_chooser_get_filename")
	purego.RegisterLibFunc(&gtkFileChooserSetCurrentFolder, gtk, "gtk_file_chooser_set_current_folder")
	purego.RegisterLibFunc(&gtkFileChooserSetCurrentName, gtk, "gtk_file_chooser_set_current_name")
	purego.RegisterLibFunc(&gtkFileChooserSetDoOverwriteConfirmation, gtk, "gtk_file_chooser_set_do_overwrite_confirmation")
	purego.RegisterLibFunc(&gtkFileFilterAddPattern, gtk, "gtk_file_filter_add_pattern")
	purego.RegisterLibFunc(&gtkFileFilterNew, gtk, "gtk_file_filter_new")
	purego.RegisterLibFunc(&gtkFileFilterSetName, gtk, "gtk_file_filter_set_name")
	purego.RegisterLibFunc(&gtkInitCheck, gtk, "gtk_init_check")
	purego.RegisterLibFunc(&gtkMainIteration, gtk, "gtk_main_iteration")
	purego.RegisterLibFunc(&gtkMessageDialogNew, gtk, "gtk_message_dialog_new")
	purego.RegisterLibFunc(&gtkWidgetDestroy, gtk, "gtk_widget_destroy")
	purego.RegisterLibFunc(&gtkWindowSetTitle, gtk, "gtk_window_set_title")
	return nil
}

func openLibrary(names ...string) (uintptr, error) {
	var firstErr error
	for _, name := range names {
		lib, err := purego.Dlopen(name, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err == nil {
			return lib, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return 0, firstErr
}

// runOnGTKThread runs f on the dedicated GTK thread and waits for f to
// complete.
func runOnGTKThread(f func()) {
	done := make(chan struct{})
	gtkCalls <- func() {
		defer close(done)
		f()
	}
	<-done
}

// goString copies a NUL-terminated C string to a Go string. It does not free
// the C string.
func goString(p *byte) string {
	return unix.BytePtrToString(p)
}
