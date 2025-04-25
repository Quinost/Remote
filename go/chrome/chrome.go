package chrome

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"
)

type ChromeController struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Clients  map[*websocket.Conn]bool
	ClientMu sync.Mutex
}

type ScrollPayload struct {
	Direction string  `json:"direction"`
	Percent   float64 `json:"percent"`
}

func RunChrome() ChromeController {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts()...)
	taskCtx, cancelTask := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	var initialURL = "about:blank" //"https://example.com/"

	err := chromedp.Run(taskCtx, chromedp.Navigate((initialURL)))

	if err != nil {
		log.Fatalf("Faile to run chrome: %v", err)
	}
	log.Println("Chrome is running successfully")

	cancelAll := func() {
		cancelTask()
		cancelAlloc()
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
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("high-dpi-support", true),
		chromedp.Flag("force-device-scale-factor", "1.0"),
		chromedp.WindowSize(530, 900),
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
	if payloadMap, ok := payload.(map[string]any); ok {
		startX, sxOk := payloadMap["startX"].(float64)
		startY, syOk := payloadMap["startY"].(float64)
		endX, exOk := payloadMap["endX"].(float64)
		endY, eyOk := payloadMap["endY"].(float64)
		duration, dOk := payloadMap["duration"].(float64)

		if sxOk && syOk && exOk && eyOk && dOk {
			log.Printf("Executing swipe: from (%f,%f) to (%f,%f) over %f ms",
				startX, startY, endX, endY, duration)

			go func() {
				err := chromedp.Run(c.Ctx,
					chromedp.ActionFunc(func(ctx context.Context) error {
						// Symulacja zdarzenia touchstart
						touchPoints := []*input.TouchPoint{
							{ID: 0, X: startX, Y: startY},
						}

						err := input.DispatchTouchEvent(input.TouchStart, touchPoints).Do(ctx)
						if err != nil {
							return err
						}

						// Symulacja płynnego ruchu
						steps := 10
						stepDuration := time.Duration(duration/float64(steps)) * time.Millisecond

						for i := 1; i <= steps; i++ {
							progress := float64(i) / float64(steps)
							currentX := startX + (endX-startX)*progress
							currentY := startY + (endY-startY)*progress

							// Aktualizacja pozycji dotyku (touchmove)
							touchPoints = []*input.TouchPoint{
								{ID: 0, X: currentX, Y: currentY},
							}

							err := input.DispatchTouchEvent(input.TouchMove, touchPoints).Do(ctx)
							if err != nil {
								return err
							}
							time.Sleep(stepDuration)
						}

						// Zakończenie dotyku (touchend)
						touchPoints = []*input.TouchPoint{
							{ID: 0, X: endX, Y: endY},
						}

						err = input.DispatchTouchEvent(input.TouchEnd, touchPoints).Do(ctx)
						return err
					}),
				)

				if err != nil {
					log.Printf("Failed to execute swipe: %v", err)
				}
			}()
		} else {
			log.Printf("Invalid swipe parameters: %v", payloadMap)
		}
	} else {
		log.Printf("Invalid payload for swipe: %v", payload)
	}
}
