//go:build windows

package main

import (
	"log"
	"os/exec"
	"remote/chrome"
	"syscall"
	"time"
)

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	keybdEvent = user32.NewProc("keybd_event")
)

func handleButtons(button string, c *ControllerExt) {
	pressKey := func(code byte) {
		keybdEvent.Call(
			uintptr(code),
			uintptr(0),
			uintptr(0),
			uintptr(0),
		)
		time.Sleep(20 * time.Millisecond)
		keybdEvent.Call(
			uintptr(code),
			uintptr(0),
			uintptr(0x0002),
			uintptr(0),
		)
	}
	
	switch button {
	case "volume_up":
		pressKey(0xAF)
	case "volume_down":
		pressKey(0xAE)
	case "exit_fullscreen":
		go chrome.Exit_Fullscreen(c.ChromeController)
	case "shutdown":
		exec.Command("shutdown", "/s", "/t", "0").Run()
	default:
		log.Printf("Unknown button: %s", button)
	}
}