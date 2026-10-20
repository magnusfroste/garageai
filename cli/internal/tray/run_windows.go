//go:build windows

package tray

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The Windows notification-area icon: Shell_NotifyIconW, a hidden window for its messages and
// TrackPopupMenu for the menu, through user32/shell32/gdi32 directly (no cgo).

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")
	gdi32   = windows.NewLazySystemDLL("gdi32.dll")
	kernel  = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassExW       = user32.NewProc("RegisterClassExW")
	pCreateWindowExW        = user32.NewProc("CreateWindowExW")
	pDefWindowProcW         = user32.NewProc("DefWindowProcW")
	pGetMessageW            = user32.NewProc("GetMessageW")
	pTranslateMessage       = user32.NewProc("TranslateMessage")
	pDispatchMessageW       = user32.NewProc("DispatchMessageW")
	pPostQuitMessage        = user32.NewProc("PostQuitMessage")
	pPostMessageW           = user32.NewProc("PostMessageW")
	pCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	pAppendMenuW            = user32.NewProc("AppendMenuW")
	pTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	pDestroyMenu            = user32.NewProc("DestroyMenu")
	pSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	pGetCursorPos           = user32.NewProc("GetCursorPos")
	pSetTimer               = user32.NewProc("SetTimer")
	pCreateIconIndirect     = user32.NewProc("CreateIconIndirect")
	pDestroyIcon            = user32.NewProc("DestroyIcon")
	pRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	pShellNotifyIconW       = shell32.NewProc("Shell_NotifyIconW")
	pCreateBitmap           = gdi32.NewProc("CreateBitmap")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pGetModuleHandleW       = kernel.NewProc("GetModuleHandleW")
	pGetConsoleProcessList  = kernel.NewProc("GetConsoleProcessList")
	pFreeConsole            = kernel.NewProc("FreeConsole")
)

const (
	wmDestroy     = 0x0002
	wmCommand     = 0x0111
	wmTimer       = 0x0113
	wmNull        = 0x0000
	wmLButtonUp   = 0x0202
	wmRButtonUp   = 0x0205
	wmApp         = 0x8000
	wmTrayIcon    = wmApp + 1
	nimAdd        = 0
	nimModify     = 1
	nimDelete     = 2
	nifMessage    = 0x1
	nifIcon       = 0x2
	nifTip        = 0x4
	mfString      = 0x0
	mfGrayed      = 0x1
	mfChecked     = 0x8
	mfSeparator   = 0x800
	tpmReturnCmd  = 0x0100
	tpmRightBtn   = 0x0002
	tpmBottomAlgn = 0x0020
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type notifyIconData struct {
	Size            uint32
	Wnd             windows.Handle
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GUIDItem        windows.GUID
	BalloonIcon     windows.Handle
}

type msg struct {
	Wnd     windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
	private uint32
}

type iconInfo struct {
	Icon     int32
	XHotspot uint32
	YHotspot uint32
	Mask     windows.Handle
	Color    windows.Handle
}

type winUI struct {
	version        string
	wnd            windows.Handle
	nid            notifyIconData
	icons          map[Level]windows.Handle
	items          []Item
	taskbarCreated uint32
	added          bool
}

var wui *winUI

// Run shows the icon and returns when the user hides it.
func Run(version string) error {
	runtime.LockOSThread()
	unlock, err := lock()
	if err != nil {
		return err
	}
	defer unlock()
	detachConsole()
	u := &winUI{version: version, icons: map[Level]windows.Handle{}}
	wui = u
	inst, _, _ := pGetModuleHandleW.Call(0)
	name, _ := windows.UTF16PtrFromString("GarageAITray")
	wc := wndClassEx{WndProc: windows.NewCallback(wndProc), Instance: windows.Handle(inst), ClassName: name}
	wc.Size = uint32(unsafe.Sizeof(wc))
	if r, _, e := pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return fmt.Errorf("RegisterClassEx: %v", e)
	}
	title, _ := windows.UTF16PtrFromString("GarageAI")
	hwnd, _, e := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(title)), 0, 0, 0, 0, 0, 0, 0, inst, 0)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx: %v", e)
	}
	u.wnd = windows.Handle(hwnd)
	tc, _ := windows.UTF16PtrFromString("TaskbarCreated")
	r, _, _ := pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(tc)))
	u.taskbarCreated = uint32(r)
	for _, l := range []Level{NotConnected, Live, Serving, Paused, Problem, Offline} {
		u.icons[l] = makeIcon(levelColor(l))
	}
	if err := u.add(); err != nil {
		return err
	}
	fmt.Println("GarageAI is in the notification area.")
	pSetTimer.Call(hwnd, 1, 5000, 0)
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	u.nid.Flags = 0
	pShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&u.nid)))
	return nil
}

func (u *winUI) add() error {
	u.nid = notifyIconData{Wnd: u.wnd, ID: 1, Flags: nifMessage | nifIcon | nifTip, CallbackMessage: wmTrayIcon}
	u.nid.Size = uint32(unsafe.Sizeof(u.nid))
	u.fill(Build(Load(time.Now(), u.version)))
	if r, _, e := pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&u.nid))); r == 0 {
		return fmt.Errorf("Shell_NotifyIcon: %v", e)
	}
	u.added = true
	return nil
}

func (u *winUI) fill(m Menu) {
	u.nid.Icon = u.icons[m.Level]
	tip, _ := windows.UTF16FromString(m.Tooltip)
	if len(tip) > len(u.nid.Tip) {
		tip = tip[:len(u.nid.Tip)-1]
		tip = append(tip, 0)
	}
	u.nid.Tip = [128]uint16{}
	copy(u.nid.Tip[:], tip)
}

func (u *winUI) refresh() {
	u.fill(Build(Load(time.Now(), u.version)))
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&u.nid)))
}

func (u *winUI) showMenu() {
	m := Build(Load(time.Now(), u.version))
	u.items = m.Items
	h, _, _ := pCreatePopupMenu.Call()
	if h == 0 {
		return
	}
	defer pDestroyMenu.Call(h)
	for i, it := range m.Items {
		if it.Separator {
			pAppendMenuW.Call(h, mfSeparator, 0, 0)
			continue
		}
		flags := uintptr(mfString)
		if it.Action == "" {
			flags |= mfGrayed
		}
		if it.Checked {
			flags |= mfChecked
		}
		label, _ := windows.UTF16PtrFromString(it.Label)
		pAppendMenuW.Call(h, flags, uintptr(i+1), uintptr(unsafe.Pointer(label)))
	}
	var pt struct{ X, Y int32 }
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(uintptr(u.wnd)) // else the menu does not close on a click elsewhere
	cmd, _, _ := pTrackPopupMenu.Call(h, tpmReturnCmd|tpmRightBtn|tpmBottomAlgn, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(u.wnd), 0)
	pPostMessageW.Call(uintptr(u.wnd), wmNull, 0, 0)
	if i := int(cmd) - 1; i >= 0 && i < len(u.items) {
		if Do(u.items[i].Action) {
			pPostQuitMessage.Call(0)
			return
		}
		u.refresh()
	}
}

func wndProc(hwnd, message, wparam, lparam uintptr) uintptr {
	u := wui
	switch uint32(message) {
	case wmTrayIcon:
		switch uint32(lparam) & 0xFFFF {
		case wmLButtonUp, wmRButtonUp:
			u.showMenu()
		}
		return 0
	case wmTimer:
		u.refresh()
		return 0
	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	if u != nil && u.taskbarCreated != 0 && uint32(message) == u.taskbarCreated {
		_ = u.add() // Explorer restarted: the icon is gone until it is added again
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, message, wparam, lparam)
	return r
}

func levelColor(l Level) [3]byte { // R, G, B
	switch l {
	case Live:
		return [3]byte{0x2e, 0x9d, 0x5a}
	case Serving:
		return [3]byte{0x2f, 0x6f, 0xdf}
	case Problem:
		return [3]byte{0xd9, 0x8a, 0x00}
	case Offline:
		return [3]byte{0xc4, 0x3c, 0x3c}
	}
	return [3]byte{0x8a, 0x8a, 0x8a}
}

// makeIcon draws a 32×32 filled circle (Windows scales it to the notification area): an AND mask
// that is transparent outside the circle, and the colour inside.
func makeIcon(c [3]byte) windows.Handle {
	const n = 32
	color := make([]byte, n*n*4)
	mask := make([]byte, n*n/8) // 1 bpp, rows of 4 bytes
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			dx, dy := float64(x)-15.5, float64(y)-15.5
			if dx*dx+dy*dy <= 14*14 {
				o := (y*n + x) * 4
				color[o], color[o+1], color[o+2] = c[2], c[1], c[0] // BGR
			} else {
				mask[y*(n/8)+x/8] |= 0x80 >> (x % 8)
			}
		}
	}
	hColor, _, _ := pCreateBitmap.Call(n, n, 1, 32, uintptr(unsafe.Pointer(&color[0])))
	hMask, _, _ := pCreateBitmap.Call(n, n, 1, 1, uintptr(unsafe.Pointer(&mask[0])))
	ii := iconInfo{Icon: 1, Mask: windows.Handle(hMask), Color: windows.Handle(hColor)}
	h, _, _ := pCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	pDeleteObject.Call(hColor)
	pDeleteObject.Call(hMask)
	return windows.Handle(h)
}

// detachConsole drops a console window that belongs to the icon alone (started by double-click or
// from the Run key without conhost --headless); a console the user typed `garageai tray` in stays.
func detachConsole() {
	var ids [4]uint32
	n, _, _ := pGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&ids[0])), 4)
	if n == 1 {
		pFreeConsole.Call()
	}
}
