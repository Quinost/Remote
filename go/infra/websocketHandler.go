package infra

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func (c *ControllerExt) RunWebsocket() {
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
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || stat.IsDir() {
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error checking file stat %s: %v", path, err)
		return
	}

	http.FileServer(http.Dir(distDir)).ServeHTTP(w, r)
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
			ip = ip.To4()
			if ip != nil {
				interfaces = append(interfaces, fmt.Sprintf("  ➜  http://%s:%d/", ip.String(), 8080))
			}
		}
	}

	return interfaces
}