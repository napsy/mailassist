package main

import (
	"fmt"
	"os/exec"
)

type desktop struct {
	notifications bool
}

func newDesktop() *desktop {
	return &desktop{notifications: true}
}

func (d *desktop) notify(msg string) error {
	if !d.notifications {
		return nil
	}
	// execute notify
	cmd := exec.Command("notify-send", "Incoming emails", fmt.Sprintf("%q", msg))
	cmd.Run()
	return nil
}
