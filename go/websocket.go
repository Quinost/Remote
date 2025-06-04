package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"remote/chrome"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"
)

const (
	ScreenshotInterval = 100 * time.Millisecond
	ScreenshotTimeout  = 2 * time.Second
	BufferSize         = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  BufferSize,
	WriteBufferSize: BufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocketMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (c *ControllerExt) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	c.registerClient(conn)

	go c.handleWebSocketMessages(conn)
	go c.streamScreenshots(conn)
}

func (c *ControllerExt) registerClient(conn *websocket.Conn) {
	c.ChromeController.ClientMu.Lock()
	defer c.ChromeController.ClientMu.Unlock()
	c.Clients[conn] = true
	log.Println("Client connected:", conn.RemoteAddr())
}

func (c *ControllerExt) unregisterClient(conn *websocket.Conn) {
	c.ChromeController.ClientMu.Lock()
	defer c.ChromeController.ClientMu.Unlock()
	if _, ok := c.ChromeController.Clients[conn]; ok {
		delete(c.ChromeController.Clients, conn)
		conn.Close()
		log.Println("Client disconnected:", conn.RemoteAddr())
	}
}

func (c *ControllerExt) handleWebSocketMessages(conn *websocket.Conn) {
	defer c.unregisterClient(conn)

	for {
		messageType, p, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message (client disconnected?): %v", err)
			}
			return
		}

		if messageType == websocket.TextMessage {
			var msg WebSocketMessage
			if err := json.Unmarshal(p, &msg); err != nil {
				log.Printf("Error decoding JSON message: %v", err)
				continue
			}

			log.Printf("Received message: %s", msg.Type)

			switch msg.Type {
			case "open_url":
				chrome.Open_Url(msg.Payload, c.ChromeController)
			case "click_at":
				chrome.Click_At(msg.Payload, c.ChromeController)
			case "send_button":
				send_button(&msg, c)
			case "scroll":
				chrome.Scroll(msg.Payload, c.ChromeController)
			default:
				log.Printf("Unknown message type: %s", msg.Type)
			}
		}
	}
}

func (c *ControllerExt) streamScreenshots(conn *websocket.Conn) {
	defer c.unregisterClient(conn)

	ticker := time.NewTicker(ScreenshotInterval)
	defer ticker.Stop()

	for range ticker.C {
		screenshot, err := c.captureScreenshot()
		if err != nil {
			if err != context.DeadlineExceeded {
				log.Printf("Error capturing screenshot: %v", err)
			}
			continue
		}

		if len(screenshot) == 0 {
			continue
		}

		if err := c.sendScreenshot(conn, screenshot); err != nil {
			log.Printf("Error sending screenshot: %v", err)
			return
		}
	}
}

func (c *ControllerExt) captureScreenshot() ([]byte, error) {
	var buf []byte
	err := chromedp.Run(c.Ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			sCtx, cancel := context.WithTimeout(ctx, ScreenshotTimeout)
			defer cancel()

			var errCap error
			buf, errCap = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatJpeg).
				WithCaptureBeyondViewport(false).
				Do(sCtx)
			return errCap
		}),
	)
	return buf, err
}

func (c *ControllerExt) sendScreenshot(conn *websocket.Conn, screenshot []byte) error {
	encoded := base64.StdEncoding.EncodeToString(screenshot)

	msg := WebSocketMessage{
		Type:    "screenshot",
		Payload: encoded,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.ChromeController.ClientMu.Lock()
	defer c.ChromeController.ClientMu.Unlock()

	return conn.WriteMessage(websocket.TextMessage, jsonData)
}

func send_button(msg *WebSocketMessage, c *ControllerExt) {
	if button, ok := msg.Payload.(string); ok {
		log.Printf("Executing send_button: %s", button)
		handleButtons(button, c)
	} else {
		log.Printf("Invalid payload for send_button: %v", msg.Payload)
	}
}
