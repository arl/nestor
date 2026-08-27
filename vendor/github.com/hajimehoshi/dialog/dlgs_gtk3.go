//go:build !darwin && !js && !windows

package dialog

func closeDialog(dlg uintptr) {
	gtkWidgetDestroy(dlg)
	// Destroying the widget alone does not remove the dialog from the screen;
	// that happens once the GTK main loop processes further events. In a
	// non-GTK app no main loop is running, so the event queue is drained
	// before returning.
	for gtkEventsPending() {
		gtkMainIteration()
	}
}

func runMsgDlg(defaultTitle string, flags uint32, messageType, buttons int32, b *MsgBuilder) int32 {
	ensureGTK()
	var response int32
	runOnGTKThread(func() {
		dlg := gtkMessageDialogNew(0, flags, messageType, buttons, "%s", b.Msg)
		defer closeDialog(dlg)
		gtkWindowSetTitle(dlg, firstOf(b.Dlg.Title, defaultTitle))
		response = gtkDialogRun(dlg)
	})
	return response
}

func (b *MsgBuilder) yesNo() bool {
	return runMsgDlg("Confirm?", 0, _GTK_MESSAGE_QUESTION, _GTK_BUTTONS_YES_NO, b) == _GTK_RESPONSE_YES
}

func (b *MsgBuilder) info() {
	runMsgDlg("Information", 0, _GTK_MESSAGE_INFO, _GTK_BUTTONS_OK, b)
}

func (b *MsgBuilder) error() {
	runMsgDlg("Error", 0, _GTK_MESSAGE_ERROR, _GTK_BUTTONS_OK, b)
}

func (b *FileBuilder) load() (string, error) {
	return chooseFile("Open File", "Open", _GTK_FILE_CHOOSER_ACTION_OPEN, b)
}

func (b *FileBuilder) save() (string, error) {
	return chooseFile("Save File", "Save", _GTK_FILE_CHOOSER_ACTION_SAVE, b)
}

func chooseFile(title string, buttonText string, action int32, b *FileBuilder) (string, error) {
	ensureGTK()
	var filename string
	var accepted bool
	runOnGTKThread(func() {
		dlg := gtkFileChooserDialogNew(title, 0, action, "Cancel", _GTK_RESPONSE_CANCEL, buttonText, _GTK_RESPONSE_ACCEPT, 0)
		defer closeDialog(dlg)

		for _, filt := range b.Filters {
			filter := gtkFileFilterNew()
			gtkFileFilterSetName(filter, filt.Desc)
			for _, ext := range filt.Extensions {
				gtkFileFilterAddPattern(filter, "*."+ext)
			}
			gtkFileChooserAddFilter(dlg, filter)
		}
		if b.StartDir != "" {
			gtkFileChooserSetCurrentFolder(dlg, b.StartDir)
		}
		if b.StartFile != "" {
			gtkFileChooserSetCurrentName(dlg, b.StartFile)
		}
		gtkFileChooserSetDoOverwriteConfirmation(dlg, true)
		if gtkDialogRun(dlg) != _GTK_RESPONSE_ACCEPT {
			return
		}
		accepted = true
		if p := gtkFileChooserGetFilename(dlg); p != nil {
			filename = goString(p)
			gFree(p)
		}
	})
	if !accepted {
		return "", ErrCancelled
	}
	return filename, nil
}

func (b *DirectoryBuilder) browse() (string, error) {
	return chooseFile("Open Folder", "Open", _GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER, &FileBuilder{Dlg: b.Dlg, StartDir: b.StartDir})
}
