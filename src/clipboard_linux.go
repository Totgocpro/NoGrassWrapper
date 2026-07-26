//go:build linux

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

var (
	x11Once sync.Once
	x11Ctx  *x11Clipboard
)

type x11Clipboard struct {
	conn  *xgb.Conn
	win   xproto.Window
	atoms struct {
		clipboard xproto.Atom
		png       xproto.Atom
		targets   xproto.Atom
	}
	mu   sync.Mutex
	data []byte
}

func initX11() error {
	var initErr error
	x11Once.Do(func() {
		conn, err := xgb.NewConn()
		if err != nil {
			initErr = err
			return
		}
		setup := xproto.Setup(conn)
		screen := setup.DefaultScreen(conn)
		root := screen.Root

		clipAtom, _ := xproto.InternAtom(conn, false, 9, "CLIPBOARD").Reply()
		pngAtom, _ := xproto.InternAtom(conn, false, 8, "image/png").Reply()
		tgtAtom, _ := xproto.InternAtom(conn, false, 7, "TARGETS").Reply()
		if clipAtom == nil || pngAtom == nil || tgtAtom == nil {
			initErr = fmt.Errorf("failed to intern atoms")
			return
		}

		wid, err := xproto.NewWindowId(conn)
		if err != nil {
			initErr = fmt.Errorf("new window id: %w", err)
			return
		}
		cookie := xproto.CreateWindowChecked(conn, 0, wid, root, 0, 0, 1, 1, 0,
			xproto.WindowClassInputOnly, screen.RootVisual,
			xproto.CwEventMask, []uint32{
				xproto.EventMaskPropertyChange | xproto.EventMaskStructureNotify,
			})
		if err := cookie.Check(); err != nil {
			initErr = fmt.Errorf("create window: %w", err)
			return
		}

		ctx := &x11Clipboard{
			conn: conn,
			win:  wid,
		}
		ctx.atoms.clipboard = clipAtom.Atom
		ctx.atoms.png = pngAtom.Atom
		ctx.atoms.targets = tgtAtom.Atom
		x11Ctx = ctx

		go ctx.eventLoop()
	})
	return initErr
}

func (c *x11Clipboard) eventLoop() {
	for {
		ev, err := c.conn.WaitForEvent()
		if err != nil {
			return
		}
		switch e := ev.(type) {
		case xproto.SelectionRequestEvent:
			c.handleSelectionRequest(e)
		}
	}
}

func (c *x11Clipboard) handleSelectionRequest(e xproto.SelectionRequestEvent) {
	c.mu.Lock()
	data := c.data
	c.mu.Unlock()

	if data == nil {
		return
	}

	prop := e.Property
	if prop == 0 {
		prop = e.Target
	}

	switch e.Target {
	case c.atoms.targets:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint32(buf[0:4], uint32(c.atoms.targets))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(c.atoms.png))
		xproto.ChangeProperty(c.conn, xproto.PropModeReplace, e.Requestor,
			prop, xproto.AtomAtom, 32, 2, buf)
	case c.atoms.png:
		xproto.ChangeProperty(c.conn, xproto.PropModeReplace, e.Requestor,
			prop, c.atoms.png, 8, uint32(len(data)), data)
	default:
		return
	}

	notify := xproto.SelectionNotifyEvent{
		Time:      xproto.TimeCurrentTime,
		Requestor: e.Requestor,
		Selection: e.Selection,
		Target:    e.Target,
		Property:  prop,
	}
	xproto.SendEvent(c.conn, false, e.Requestor,
		xproto.EventMaskNoEvent, string(notify.Bytes()))
}

func copyImageToClipboard(data []byte) error {
	if isWayland() {
		return copyImageWayland(data)
	}
	return copyImageX11(data)
}

func copyImageX11(data []byte) error {
	if err := initX11(); err != nil {
		return fmt.Errorf("X11 init: %w", err)
	}
	x11Ctx.mu.Lock()
	x11Ctx.data = append([]byte{}, data...)
	x11Ctx.mu.Unlock()
	xproto.SetSelectionOwner(x11Ctx.conn, x11Ctx.win, x11Ctx.atoms.clipboard, xproto.TimeCurrentTime)
	return nil
}

func copyImageWayland(data []byte) error {
	f, err := os.CreateTemp("", "ngw-clip-*.png")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	f.Close()

	cmd := exec.Command("sh", "-c", "wl-copy -t image/png < "+f.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wl-copy: %w (output: %s)", err, string(out))
	}
	return nil
}
