package chrome

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	cfg "remote/config"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"

	cu "github.com/Davincible/chromedp-undetected"
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

func RunChrome(cfg *cfg.Config) ChromeController {
	taskCtx, cancelTask, _ := cu.New(
		cu.NewConfig(cu.WithChromeFlags(opts(cfg)...)))

	var initialURL = cfg.DefaultWebpage

	err := chromedp.Run(taskCtx, chromedp.Navigate((initialURL)))

	listenTarget(&taskCtx)

	if err != nil {
		log.Fatalf("Failed to run chrome: %v", err)
	}
	log.Println("Chrome is running successfully")

	cancelAll := func() {
		cancelTask()
	}

	return ChromeController{
		Ctx:     taskCtx,
		Cancel:  cancelAll,
		Clients: make(map[*websocket.Conn]bool),
	}
}

func opts(cfg *cfg.Config) []chromedp.ExecAllocatorOption {
	execPath, _ := os.Executable()
	dir := filepath.Join(filepath.Dir(execPath), "ChromeProfile")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("hide-scrollbars", false),

		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("high-dpi-support", true),
		chromedp.Flag("force-device-scale-factor", "1.0"),
		chromedp.Flag("no-first-runs", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("use-gl", "desktop"),
		chromedp.Flag("enable-webgl", true),
		chromedp.Flag("hide-crash-restore-bubble", true),
		chromedp.WindowSize(cfg.Resolution.Width, cfg.Resolution.Height),
		chromedp.UserDataDir(dir),
		chromedp.Flag("profile-directory", cfg.Profile),
	)

	return opts
}

func listenTarget(taskCtx *context.Context) {
	chromedp.ListenTarget(*taskCtx, func(ev any) {
		if ev, ok := ev.(*target.EventTargetCreated); ok {
			if ev.TargetInfo.Type == "page" || ev.TargetInfo.Type == "window" {
				go func() {
					c := chromedp.FromContext(*taskCtx)
					err := target.CloseTarget(ev.TargetInfo.TargetID).Do(cdp.WithExecutor(*taskCtx, c.Browser))
					if err != nil {
						log.Printf("Failed to close %s: %v", ev.TargetInfo.TargetID, err)
					} else {
						log.Printf("Closed %s", ev.TargetInfo.TargetID)
					}
				}()
			}
		}
	})
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

func Type_Enter(payload any, c *ChromeController) {
	var text string
	var valid bool

	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if textVal, textOk := payloadMap["text"].(string); textOk {
			text = textVal
			valid = true
		}
	} else if textStr, ok := payload.(string); ok {
		text = textStr
		valid = true
	}

	if !valid {
		log.Printf("Invalid payload for type_enter: %T %v", payload, payload)
		return
	}

	log.Printf("Executing type_enter: text=%s", text)

	go func(textToType string) {
		jsText := fmt.Sprintf("%q", textToType)

		err := chromedp.Run(c.Ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				var focused = document.activeElement;
				if (focused && (focused.tagName === 'INPUT' || focused.tagName === 'TEXTAREA' || focused.contentEditable === 'true')) {
					focused.value = %s;
					focused.dispatchEvent(new Event('input', { bubbles: true }));
					focused.dispatchEvent(new Event('change', { bubbles: true }));
				}
			`, jsText), nil),
			chromedp.Sleep(50*time.Millisecond),
		)
		if err != nil {
			log.Printf("Error during type: %v", err)
		} else {
			log.Printf("Successfully typed: %s", textToType)
		}
	}(text)
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
	var direction string
	var percent float64
	var valid bool

	if payloadMap, ok := payload.(map[string]interface{}); ok {
		dir, dirOk := payloadMap["direction"].(string)
		pct, pctOk := payloadMap["percent"].(float64)

		if dirOk && pctOk {
			direction = dir
			percent = pct
			valid = true
		}
	} else if scrollPayload, ok := payload.(ScrollPayload); ok {
		direction = scrollPayload.Direction
		percent = scrollPayload.Percent
		valid = true
	}

	if !valid {
		log.Printf("Wrong payload for scroll: %T %v", payload, payload)
		return
	}

	log.Printf("Executing scroll: direction=%s, percent=%.2f%%", direction, percent)

	var scrollScript string
	if direction == "up" {
		scrollScript = fmt.Sprintf(`
            (function() {
                var scrollAmount = -%f * window.innerHeight / 100;
                window.scrollBy(0, Math.max(scrollAmount, -window.scrollY));
                return true;
            })();
        `, percent)
	} else {
		scrollScript = fmt.Sprintf(`
            (function() {
                var scrollAmount = %f * window.innerHeight / 100;
				var maxScroll = document.documentElement.scrollHeight - document.documentElement.clientHeight;
                window.scrollBy(0, Math.min(scrollAmount, maxScroll - window.scrollY));
                return true;
            })();
        `, percent)
	}

	go func(script string) {
		var result bool
		err := chromedp.Run(c.Ctx,
			chromedp.Evaluate(script, &result),
		)
		if err != nil {
			log.Printf("Error during scroll: %v", err)
		}

	}(scrollScript)
}
