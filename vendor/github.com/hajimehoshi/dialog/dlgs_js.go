// SPDX-License-Identifier: ISC
// SPDX-FileCopyrightText: 2026 Hajime Hoshi

package dialog

import (
	"errors"
	"fmt"
	"syscall/js"
)

// Browsers do not allow customizing the title or the button labels of the
// built-in dialogs: the title is determined by the page origin, and confirm
// always shows OK/Cancel.

func (b *MsgBuilder) yesNo() bool {
	return js.Global().Call("confirm", b.Msg).Bool()
}

func (b *MsgBuilder) info() {
	js.Global().Call("alert", b.Msg)
}

func (b *MsgBuilder) error() {
	js.Global().Call("alert", b.Msg)
}

func (b *FileBuilder) load() (string, error) {
	return "", fmt.Errorf("dialog: FileBuilder.Load is not supported in browsers: %w", errors.ErrUnsupported)
}

func (b *FileBuilder) save() (string, error) {
	return "", fmt.Errorf("dialog: FileBuilder.Save is not supported in browsers: %w", errors.ErrUnsupported)
}

func (b *DirectoryBuilder) browse() (string, error) {
	return "", fmt.Errorf("dialog: DirectoryBuilder.Browse is not supported in browsers: %w", errors.ErrUnsupported)
}
