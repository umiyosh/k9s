package views

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	maxBuff1     int64 = 200
	refreshRate        = 200 * time.Millisecond
	maxCleanse         = 100
	logBuffSize        = 100
	flushTimeout       = 200 * time.Millisecond

	logCoFmt = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
	logFmt   = " Logs([fg:bg:]%s) "
)

type (
	masterView interface {
		backFn() actionHandler
		appView() *appView
	}

	logsView struct {
		*tview.Pages

		app          *appView
		title        string
		parent       loggable
		container    string
		actions      keyActions
		cancelFunc   context.CancelFunc
		showPrevious bool
	}
)

func newLogsView(title string, app *appView, parent loggable) *logsView {
	v := logsView{
		app:    app,
		Pages:  tview.NewPages(),
		parent: parent,
		title:  title,
	}

	return &v
}

// Protocol...

func (v *logsView) reload(co string, parent loggable, title string, prevLogs bool) {
	v.parent, v.title, v.showPrevious = parent, title, prevLogs
	v.deletePage()
	v.AddPage(co, newLogView(co, v.app, v.backCmd), true, true)
	v.container = co
	v.load()
}

// SetActions to handle keyboard events.
func (v *logsView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints show action hints
func (v *logsView) hints() hints {
	l := v.CurrentPage().Item.(*logView)
	return l.actions.toHints()
}

func (v *logsView) backFn() actionHandler {
	return v.backCmd
}

func (v *logsView) deletePage() {
	v.RemovePage(v.container)
	v.container = ""
}

func (v *logsView) stop() {
	if v.cancelFunc == nil {
		return
	}
	v.cancelFunc()
	log.Debug().Msgf("Canceling logs...")
	v.cancelFunc = nil
}

func (v *logsView) load() {
	if err := v.doLoad(v.parent.getSelection(), v.container); err != nil {
		v.app.flash().err(err)
		l := v.CurrentPage().Item.(*logView)
		l.logLine("😂 Doh! No logs are available at this time. Check again later on...")
		return
	}
	v.app.SetFocus(v)
}

func (v *logsView) doLoad(path, co string) error {
	v.stop()

	maxBuff := int64(v.app.config.K9s.LogRequestSize)
	l := v.CurrentPage().Item.(*logView)
	l.logs.Clear()
	l.path = path

	var fmat string
	if co == "" {
		fmat = skinTitle(fmt.Sprintf(logFmt, path), v.app.styles.Style)
	} else {
		fmat = skinTitle(fmt.Sprintf(logCoFmt, path, co), v.app.styles.Style)
	}
	l.SetTitle(fmat)

	c := make(chan string, 10)
	go func(l *logView) {
		buff, index := make([]string, logBuffSize), 0
		for {
			select {
			case line, ok := <-c:
				if !ok {
					l.flush(index, buff)
					index = 0
					return
				}
				if index < logBuffSize {
					buff[index] = line
					index++
					continue
				}
				l.flush(index, buff)
				index = 0
				buff[index] = line
			case <-time.After(flushTimeout):
				l.flush(index, buff)
				index = 0
			}
		}
	}(l)

	ns, po := namespaced(path)
	res, ok := v.parent.getList().Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.parent.getList().Resource())
	}
	var ctx context.Context
	ctx = context.WithValue(context.Background(), resource.IKey("informer"), v.app.informer)
	ctx, v.cancelFunc = context.WithCancel(ctx)
	opts := resource.LogOptions{
		Namespace: ns,
		Name:      po,
		Container: co,
		Lines:     maxBuff,
		Previous:  v.showPrevious,
	}
	if err := res.Logs(ctx, c, opts); err != nil {
		v.cancelFunc()
		return err
	}

	return nil
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logsView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.stop()
	v.parent.switchPage(v.title)

	return evt
}
