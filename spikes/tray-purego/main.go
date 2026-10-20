//go:build darwin
package main

import (
	"runtime"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func init() { runtime.LockOSThread() }

func main() {
	if _, err := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_GLOBAL); err != nil {
		panic(err)
	}
	app := objc.ID(objc.GetClass("NSApplication")).Send(objc.RegisterName("sharedApplication"))
	app.Send(objc.RegisterName("setActivationPolicy:"), 1) // accessory: no Dock icon
	bar := objc.ID(objc.GetClass("NSStatusBar")).Send(objc.RegisterName("systemStatusBar"))
	item := bar.Send(objc.RegisterName("statusItemWithLength:"), float64(-1))
	btn := item.Send(objc.RegisterName("button"))
	title := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "GarageAI")
	btn.Send(objc.RegisterName("setTitle:"), title)
	menu := objc.ID(objc.GetClass("NSMenu")).Send(objc.RegisterName("new"))
	quit := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "Quit")
	key := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "q")
	menu.Send(objc.RegisterName("addItemWithTitle:action:keyEquivalent:"), quit, objc.RegisterName("terminate:"), key)
	item.Send(objc.RegisterName("setMenu:"), menu)
	app.Send(objc.RegisterName("run"))
}
