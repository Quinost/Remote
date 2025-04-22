package chrome

import (
	"context"
	"log"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"
)

type ChromeController struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Clients  map[*websocket.Conn]bool
	ClientMu sync.Mutex
}

func RunChrome() ChromeController {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts()...)
	taskCtx, cancelTask := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	var initialURL = "https://example.com/"

	err := chromedp.Run(taskCtx, chromedp.Navigate((initialURL)))

	if err != nil {
		log.Fatalf("Faile to run chrome: %v", err)
	}
	log.Println("Chrome is running successfully")

	cancelAll := func() {
		cancelTask()  // zamknij kontekst zadaniowy
		cancelAlloc() // zamknij allocator
	}

	return ChromeController{
		Ctx:     taskCtx,
		Cancel:  cancelAll,
		Clients: make(map[*websocket.Conn]bool),
	}
}

func opts() []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("mute-audio", false),
		chromedp.WindowSize(520, 900),
		chromedp.UserDataDir(""),
	)

	return opts
}

func Open_Url(payload any, c *ChromeController) {
	if url, ok := payload.(string); ok && url != "" {
		log.Printf("Executing open_url: %v", url)

		go func(targetUrl string) {
			err := chromedp.Run(c.Ctx, chromedp.Navigate(targetUrl), chromedp.WaitVisible(`body`, chromedp.ByQuery))

			if err != nil {
				log.Printf("Failed to open URL: %v", err)
			} else {
				log.Printf("Opened URL: %s", targetUrl)
			}
		}(url)
	} else {
		log.Printf("Invalid payload for open_url: %v", payload)
	}
}

func Click_At(payload any, c *ChromeController) {
	if payloadMap, ok := payload.(map[string]any); ok {
		xVal, xOk := payloadMap["x"].(float64)
		yVal, yOk := payloadMap["y"].(float64)

		if xOk && yOk {
			log.Printf("Executing click_at: x=%f, y=%f", xVal, yVal)

			go func(clickX, clickY float64) {
				chromedp.Run(c.Ctx, chromedp.MouseClickXY(clickX, clickY))

			}(xVal, yVal)
		} else {
			log.Printf("Wrong values for click_at: %v", payloadMap)
		}
	} else {
		log.Printf("Invalid payload for click_at: %v", payload)
	}
}

func Exit_Fullscreen(c *ChromeController) {
	err := chromedp.Run(c.Ctx,
		chromedp.Evaluate(`document.exitFullscreen()`, nil),
	)
	if err == nil {
		log.Println("Success: document.exitFullscreen()")
	}
}

func Scroll(payload any, c *ChromeController) {
	panic("not implemented")
}
