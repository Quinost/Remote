package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"remote/chrome"
	"strings"
	"syscall"
	"time"
)

type ControllerExt struct {
	*chrome.ChromeController
	Server *http.Server
}

func main() {
	base := chrome.RunChrome()
	c := ControllerExt{ChromeController: &base}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	go c.runWebsocket()
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

func (c *ControllerExt) runWebsocket() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.HandleWebSocket)
	mux.HandleFunc("/", c.handleIndex)

	c.Server = &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Println("Running server on port 0.0.0.0:8080...")
	if err := c.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error HTTP: %v", err)
	}
}

func (c *ControllerExt) handleIndex(w http.ResponseWriter, r *http.Request) {
	distDir := "./dist"

	path := filepath.Join(distDir, r.URL.Path)
	_, err := os.Stat(path)
	fsHandler := http.FileServer(http.Dir(distDir))

	if err != nil {
		_, err := os.Stat(path)

		if os.IsNotExist(err) || strings.HasSuffix(r.URL.Path, "/") {
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}

	}

	fsHandler.ServeHTTP(w, r)
}
