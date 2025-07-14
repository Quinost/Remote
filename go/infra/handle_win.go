//go:build windows

package infra

import (
	"log"
	"os/exec"
	"remote/chrome"
	"time"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	keybdEvent              = user32.NewProc("keybd_event")
	setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	ES_CONTINUOUS       = 0x80000000
	ES_SYSTEM_REQUIRED  = 0x00000001
	ES_DISPLAY_REQUIRED = 0x00000002
)

const (
	VK_VOLUME_UP    = 0xAF
	VK_VOLUME_DOWN  = 0xAE
	KEYEVENTF_KEYUP = 0x0002
)

func handleButtons(button string, c *ControllerExt) {
	pressKey := func(code uintptr) {
		keybdEvent.Call(
			code,
			uintptr(0),
			uintptr(0),
			uintptr(0),
		)
		time.Sleep(20 * time.Millisecond)

		keybdEvent.Call(
			code,
			uintptr(0),
			uintptr(KEYEVENTF_KEYUP),
			uintptr(0),
		)
	}

	switch button {
	case "volume_up":
		pressKey(VK_VOLUME_UP)
	case "volume_down":
		pressKey(VK_VOLUME_DOWN)
	case "exit_fullscreen":
		go chrome.Exit_Fullscreen(c.ChromeController)
	case "shutdown":
		exec.Command("shutdown", "/s", "/t", "0").Run()
	default:
		log.Printf("Unknown button: %s", button)
	}
}

func PreventSleep() {
	ret, _, _ := setThreadExecutionState.Call(
		uintptr(ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED),
	)

	if ret != 0 {
		log.Printf("Prevent sleep successfully set")
	}
}

func AllowSleep() {
	ret, _, _ := setThreadExecutionState.Call(
		uintptr(ES_CONTINUOUS),
	)

	if ret != 0 {
		log.Printf("Allow sleep successfully set")
	}
}
