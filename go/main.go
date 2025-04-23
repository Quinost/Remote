package main

import (
	"context"
	"fmt"
	"log"
	"net"
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

	writeInterfaces()

	c.Server = &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

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

func writeInterfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Failure: %v", err)
	}

	interfaces := getInterfaces(ifaces)

	fmt.Println("------------------------------------")
	fmt.Println("Server interfaces:")
	for _, inter := range interfaces {
		fmt.Println(inter)
	}
	fmt.Println("------------------------------------")
}

func getInterfaces(ifaces []net.Interface) []string {
	var interfaces []string

	interfaces = append(interfaces, "  ➜  http://localhost:8080")

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				interfaces = append(interfaces, fmt.Sprintf("  ➜  http://%s:%d/", ip.String(), 8080))
			}
		}
	}

	return interfaces
}
