package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"remote/chrome"
	"remote/config"
	"remote/infra"
	"syscall"
	"time"
)

func main() {
	config := config.LoadConfig()
	fmt.Println(config)
	base := chrome.RunChrome(&config)
	c := infra.ControllerExt{ChromeController: &base}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	go c.RunWebsocket()
	<-stopChan

	log.Println("Received shutdown signal, shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := c.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Closing browser...")
	c.ChromeController.Cancel()
	time.Sleep(1 * time.Second)
	log.Println("Application shutted down")
}
