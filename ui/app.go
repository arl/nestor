package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"gioui.org/app"
)

type Application struct {
	mu   sync.Mutex
	wins map[string]*Window

	stack []func()

	ctx    context.Context
	cancel context.CancelFunc
}

func NewApplication(title string, main View, opts ...app.Option) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	a := &Application{
		ctx:    ctx,
		cancel: cancel,
		wins:   make(map[string]*Window),
	}

	mainWindow, err := a.newWindow(true, title, main)
	if err != nil {
		panic(err)
	}
	_ = mainWindow
	return a
}

var ErrWindowTitleExists = errors.New("window title already exists")

func (a *Application) NewWindow(title string, v View, opts ...app.Option) (*Window, error) {
	return a.newWindow(false, title, v, opts...)
}

func (a *Application) newWindow(isMain bool, title string, v View, opts ...app.Option) (*Window, error) {
	if a.WindowExists(title) {
		fmt.Println("WindowExists", title)
		return nil, ErrWindowTitleExists
	}

	chw := make(chan *Window)

	go func() {
		w := Window{App: a}
		opts = append(opts, app.Title(title))
		w.Window.Option(opts...)

		a.mu.Lock()
		a.wins[title] = &w
		a.mu.Unlock()

		chw <- &w

		v.Run(a.ctx, &w)

		a.mu.Lock()
		delete(a.wins, title)
		a.mu.Unlock()

		if isMain {
			a.cancel()
		}
	}()

	return <-chw, nil
}

func (a *Application) Defer(f func()) {
	a.stack = append(a.stack, f)
}

func (a *Application) Wait() {
	go func() {
		<-a.ctx.Done()
		// Run defer stack before exiting
		for i := len(a.stack) - 1; i >= 0; i-- {
			a.stack[i]()
		}
		os.Exit(0)
	}()

	app.Main()
}

func (a *Application) Shutdown() {
	a.cancel()
}

func (a *Application) WindowExists(title string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.wins[title]
	return ok
}

// Window holds window state.
type Window struct {
	App *Application
	app.Window
}

// A View handles the event loop for a Window.
type View interface {
	// Run handles the window event loop. When the context is done, the view
	// should exit from the main loop.
	Run(ctx context.Context, w *Window) error
}
