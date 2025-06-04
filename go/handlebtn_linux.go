//go:build linux

package main

import (
	"log"
	"os/exec"
	"remote/chrome"
)

func handleButtons(button string, c *ControllerExt) {
	switch button {
	case "volume_up":
		exec.Command("amixer", "set", "Master", "5%+").Run() 
	case "volume_down":
		exec.Command("amixer", "set", "Master", "5%-").Run()
	case "exit_fullscreen":
		go chrome.Exit_Fullscreen(c.ChromeController)
	case "shutdown":
		exec.Command("shutdown", "-h", "now").Run()
	default:
		log.Printf("Unknown button: %s", button)
	}
}