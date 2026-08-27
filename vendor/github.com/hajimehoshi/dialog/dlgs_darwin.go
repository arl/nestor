package dialog

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

const (
	nsAlertFirstButtonReturn                = 1000
	nsApplicationActivationPolicyProhibited = 2
	nsApplicationActivationPolicyAccessory  = 1
)

const (
	alertYesNo = iota
	alertError
	alertInfo
)

const (
	loadDlg = iota
	saveDlg
	dirDlg
)

func checkActivationPolicy() {
	nsApp := objc.ID(class_NSApplication).Send(sel_sharedApplication)
	policy := objc.Send[int](nsApp, sel_activationPolicy)
	if policy == nsApplicationActivationPolicyProhibited {
		nsApp.Send(sel_setActivationPolicy, nsApplicationActivationPolicyAccessory)
	}
}

func alertDlg(msg, title string, style int) bool {
	var result bool
	runOnMainThread(func() {
		alert := objc.ID(class_NSAlert).Send(sel_alloc).Send(sel_init)
		if title != "" {
			alert.Send(sel_window).Send(sel_setTitle, toNSString(title))
		}
		alert.Send(sel_setMessageText, toNSString(msg))
		switch style {
		case alertYesNo:
			alert.Send(sel_addButtonWithTitle, toNSString("Yes"))
			alert.Send(sel_addButtonWithTitle, toNSString("No"))
		case alertError:
			alert.Send(sel_setIcon, objc.ID(class_NSImage).Send(sel_imageNamed, nsImageNameCaution))
			alert.Send(sel_addButtonWithTitle, toNSString("OK"))
		case alertInfo:
			alert.Send(sel_setIcon, objc.ID(class_NSImage).Send(sel_imageNamed, nsImageNameInfo))
			alert.Send(sel_addButtonWithTitle, toNSString("OK"))
		}
		checkActivationPolicy()
		result = objc.Send[int](alert, sel_runModal) == nsAlertFirstButtonReturn
	})
	return result
}

func fileDlg(mode int, title string, exts []string, relaxExt bool, startDir, filename string) (string, error) {
	var resultPath string
	var resultErr error

	runOnMainThread(func() {
		var panel objc.ID
		if mode == saveDlg {
			panel = objc.ID(class_NSSavePanel).Send(sel_savePanel)
		} else {
			panel = objc.ID(class_NSOpenPanel).Send(sel_openPanel)
			if mode == dirDlg {
				panel.Send(sel_setCanChooseDirectories, true)
				panel.Send(sel_setCanChooseFiles, false)
			}
		}

		panel.Send(sel_setFloatingPanel, true)

		if title != "" {
			panel.Send(sel_setTitle, toNSString(title))
		}
		if len(exts) > 0 {
			nsExts := make([]objc.ID, len(exts))
			for i, ext := range exts {
				nsExts[i] = toNSString(ext)
			}
			array := objc.ID(class_NSArray).Send(sel_arrayWithObjects_count,
				unsafe.Pointer(&nsExts[0]), uint(len(nsExts)))
			runtime.KeepAlive(nsExts)
			panel.Send(sel_setAllowedFileTypes, array)
		}
		if relaxExt {
			panel.Send(sel_setAllowsOtherFileTypes, true)
		}
		if startDir != "" {
			url := objc.ID(class_NSURL).Send(sel_fileURLWithPath, toNSString(startDir))
			panel.Send(sel_setDirectoryURL, url)
		}
		if filename != "" {
			panel.Send(sel_setNameFieldStringValue, toNSString(filename))
		}

		checkActivationPolicy()

		if objc.Send[int](panel, sel_runModal) == 0 {
			return
		}

		var url objc.ID
		if mode == saveDlg {
			url = panel.Send(sel_URL)
		} else {
			url = panel.Send(sel_URLs).Send(sel_objectAtIndex, uint(0))
		}
		pathNS := url.Send(sel_path)
		if pathNS == 0 {
			resultErr = errors.New("failed to get file-system representation for selected URL")
			return
		}
		resultPath = toGoString(pathNS)
	})

	return resultPath, resultErr
}

func (b *MsgBuilder) yesNo() bool {
	return alertDlg(b.Msg, b.Dlg.Title, alertYesNo)
}

func (b *MsgBuilder) info() {
	alertDlg(b.Msg, b.Dlg.Title, alertInfo)
}

func (b *MsgBuilder) error() {
	alertDlg(b.Msg, b.Dlg.Title, alertError)
}

func (b *FileBuilder) load() (string, error) {
	return b.run(false)
}

func (b *FileBuilder) save() (string, error) {
	return b.run(true)
}

func (b *FileBuilder) run(save bool) (string, error) {
	star := false
	var exts []string
	for _, filt := range b.Filters {
		for _, ext := range filt.Extensions {
			if ext == "*" {
				star = true
			} else {
				exts = append(exts, ext)
			}
		}
	}
	if star && save {
		/* OSX doesn't allow the user to switch visible file types/extensions. Also
		** NSSavePanel's allowsOtherFileTypes property has no effect for an open
		** dialog, so if "*" is a possible extension we must always show all files. */
		exts = nil
	}
	mode := loadDlg
	if save {
		mode = saveDlg
	}
	f, err := fileDlg(mode, b.Dlg.Title, exts, star, b.StartDir, b.StartFile)
	if f == "" && err == nil {
		return "", ErrCancelled
	}
	return f, err
}

func (b *DirectoryBuilder) browse() (string, error) {
	f, err := fileDlg(dirDlg, b.Dlg.Title, nil, false, b.StartDir, "")
	if f == "" && err == nil {
		return "", ErrCancelled
	}
	return f, err
}
