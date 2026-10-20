//go:build darwin

package tray

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The macOS menu-bar item: AppKit called from Go through purego, with no C and no cgo, so Bridge
// stays cross-compiled from Linux (spikes/tray-purego). AppKit must run on the main thread, which
// is where main.main starts; LockOSThread in init keeps the main goroutine there.

func init() { runtime.LockOSThread() }

func sel(name string) objc.SEL { return objc.RegisterName(name) }

func class(name string) objc.ID { return objc.ID(objc.GetClass(name)) }

func nsString(s string) objc.ID {
	return class("NSString").Send(sel("stringWithUTF8String:"), s)
}

type darwinUI struct {
	version  string
	app      objc.ID
	item     objc.ID
	menu     objc.ID
	delegate objc.ID
	items    []Item // the menu on screen: a click's tag is an index into it
	text     string // the menu's text, to skip redrawing an unchanged menu
}

var ui *darwinUI

// symbols are SF Symbols (macOS 11+), template images that follow light and dark menu bars.
var symbols = map[Level]string{
	NotConnected: "circle.dashed",
	Live:         "bolt.horizontal.circle",
	Serving:      "bolt.horizontal.circle.fill",
	Paused:       "pause.circle",
	Problem:      "exclamationmark.triangle",
	Offline:      "bolt.slash.circle",
}

// Run shows the item and returns only on error; quitting ends the process.
func Run(version string) error {
	unlock, err := lock()
	if err != nil {
		return err
	}
	defer unlock()
	if _, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		return fmt.Errorf("could not load AppKit: %w", err)
	}
	delegateClass, err := objc.RegisterClass("GarageAITrayTarget", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
		{Cmd: sel("clicked:"), Fn: func(self objc.ID, _ objc.SEL, sender objc.ID) {
			ui.click(objc.Send[int](sender, sel("tag")))
		}},
		{Cmd: sel("tick:"), Fn: func(self objc.ID, _ objc.SEL, _ objc.ID) {
			ui.refresh()
		}},
	})
	if err != nil {
		return err
	}
	pool := class("NSAutoreleasePool").Send(sel("new"))
	app := class("NSApplication").Send(sel("sharedApplication"))
	app.Send(sel("setActivationPolicy:"), 1) // accessory: no Dock icon, no app menu
	item := class("NSStatusBar").Send(sel("systemStatusBar")).Send(sel("statusItemWithLength:"), float64(-1))
	item.Send(sel("retain"))
	menu := class("NSMenu").Send(sel("new"))
	menu.Send(sel("setAutoenablesItems:"), false)
	item.Send(sel("setMenu:"), menu)
	ui = &darwinUI{version: version, app: app, item: item, menu: menu, delegate: objc.ID(delegateClass).Send(sel("new"))}
	ui.refresh()
	class("NSTimer").Send(sel("scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:"),
		float64(5), ui.delegate, sel("tick:"), objc.ID(0), true)
	pool.Send(sel("drain"))
	app.Send(sel("run"))
	return nil
}

func (u *darwinUI) refresh() {
	m := Build(Load(time.Now(), u.version))
	button := u.item.Send(sel("button"))
	if img := class("NSImage").Send(sel("imageWithSystemSymbolName:accessibilityDescription:"),
		nsString(symbols[m.Level]), nsString("GarageAI: "+m.Level.String())); img != 0 {
		img.Send(sel("setTemplate:"), true)
		button.Send(sel("setImage:"), img)
	} else {
		button.Send(sel("setTitle:"), nsString("GarageAI")) // macOS 10.15 and older: no SF Symbols
	}
	button.Send(sel("setToolTip:"), nsString(m.Tooltip))
	text := m.Text()
	if text == u.text {
		return
	}
	u.text, u.items = text, m.Items
	u.menu.Send(sel("removeAllItems"))
	for i, it := range m.Items {
		if it.Separator {
			u.menu.Send(sel("addItem:"), class("NSMenuItem").Send(sel("separatorItem")))
			continue
		}
		var action objc.SEL
		if it.Action != "" {
			action = sel("clicked:")
		}
		mi := class("NSMenuItem").Send(sel("alloc")).Send(sel("initWithTitle:action:keyEquivalent:"), nsString(it.Label), action, nsString(""))
		if it.Action != "" {
			mi.Send(sel("setTarget:"), u.delegate)
			mi.Send(sel("setTag:"), i)
			mi.Send(sel("setEnabled:"), true)
		} else {
			mi.Send(sel("setEnabled:"), false)
		}
		if it.Checked {
			mi.Send(sel("setState:"), 1)
		}
		u.menu.Send(sel("addItem:"), mi)
		mi.Send(sel("release"))
	}
}

func (u *darwinUI) click(i int) {
	if i < 0 || i >= len(u.items) {
		return
	}
	if quit := Do(u.items[i].Action); quit {
		u.app.Send(sel("terminate:"), objc.ID(0))
	}
	u.text = "" // redraw: a toggled check mark
	u.refresh()
}
