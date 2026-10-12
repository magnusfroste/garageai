package service

import (
	"strings"
	"time"
)

// launchdReplace registers plist under label, replacing a running instance. launchctl bootout returns
// before launchd has finished tearing the old instance down, and a bootstrap in that window fails
// with "Bootstrap failed: 5: Input/output error" (seen on two Macs on 2026-10-12, when connect was
// run again over a running tunnel). So: bootout, wait until the label is gone, then bootstrap, with
// retries for the cases where launchd is slower than that.
func launchdReplace(label, plist string) error {
	_ = run("launchctl", "bootout", "system/"+label)
	for i := 0; i < 50 && run("launchctl", "print", "system/"+label) == nil; i++ {
		time.Sleep(200 * time.Millisecond)
	}
	var err error
	for i := 0; i < 5; i++ {
		if err = run("launchctl", "bootstrap", "system", plist); err == nil || !strings.Contains(err.Error(), "Input/output error") {
			return err
		}
		time.Sleep(time.Second)
	}
	return err
}
