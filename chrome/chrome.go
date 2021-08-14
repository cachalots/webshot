package chrome

import (
	"context"
	"fmt"
	"github.com/4everland/screenshot/lib"
	"github.com/chromedp/chromedp"
	"time"
)

type Chrome struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewLocalChrome(execPath, proxy string) *Chrome {
	if execPath != "" {
		LocalChromeOptions = append(LocalChromeOptions, chromedp.ExecPath(execPath))
	}

	if proxy != "" {
		LocalChromeOptions = append(LocalChromeOptions, chromedp.ProxyServer(proxy))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), LocalChromeOptions...)
	return &Chrome{
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (c Chrome) Screenshot(o ScreenshotOptions) (b []byte) {
	lib.Logger().Info(fmt.Sprintf("%s start screenshot, available sceond: %f",
		o.URL.String(), o.EndTime.Sub(time.Now()).Seconds()))

	timeoutCtx, cancel := context.WithTimeout(c.Ctx, o.EndTime.Sub(time.Now()))
	defer cancel()

	ctx, cancel := chromedp.NewContext(timeoutCtx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.EmulateViewport(o.Width, o.Height),
		chromedp.Navigate(o.URL.String()),
		chromedp.Sleep(o.Delay * time.Millisecond),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			lib.Logger().Info(fmt.Sprintf("%s screenshoting, available sceond: %d",
				o.URL.String(), time.Now().Unix()))
			if o.Full {
				return chromedp.FullScreenshot(&b, 100).Do(ctx)
			}

			return chromedp.CaptureScreenshot(&b).Do(ctx)
		}),
	}); err != nil {
		lib.Logger().Error("chrome screenshot err:"+err.Error(), lib.ChromeLog)
		return
	}

	lib.Logger().Info(fmt.Sprintf("%s screenshot success %d", o.URL.String(), time.Now().Unix()))

	return b
}
