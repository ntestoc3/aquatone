package agents

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/michenriksen/aquatone/core"
)

type URLScreenshotter struct {
	session *core.Session
}

func NewURLScreenshotter() *URLScreenshotter {
	return &URLScreenshotter{}
}

func (a *URLScreenshotter) ID() string {
	return "agent:url_screenshotter"
}

func (a *URLScreenshotter) Register(s *core.Session) error {
	s.EventBus.SubscribeAsync(core.URLResponsive, a.OnURLResponsive, false)
	s.EventBus.SubscribeAsync(core.SessionEnd, a.OnSessionEnd, false)
	a.session = s
	return nil
}

func (a *URLScreenshotter) OnURLResponsive(url string) {
	a.session.Out.Debug("[%s] Received new responsive URL %s\n", a.ID(), url)
	page := a.session.GetPage(url)
	if page == nil {
		a.session.Out.Error("Unable to find page for URL: %s\n", url)
		return
	}

	a.session.WaitGroup.Add()
	go func(page *core.Page) {
		defer a.session.WaitGroup.Done()
		a.screenshotPage(page)
	}(page)
}

func (a *URLScreenshotter) OnSessionEnd() {
	a.session.Out.Debug("[%s] Received SessionEnd event\n", a.ID())
}

func (a *URLScreenshotter) screenshotPage(page *core.Page) {
	filePath := fmt.Sprintf("screenshots/%s.png", page.BaseFilename())

	var actx context.Context

	if *a.session.Options.UseRemoteChrome {
		u, err := url.Parse(*a.session.Options.ChromeDevToolsURL)
		if err != nil {
			a.session.Stats.IncrementScreenshotFailed()
			a.session.Out.Error("%s screenshot %s not valid remote chrome dev tools url, error: %s\n", page.URL, *a.session.Options.ChromeDevToolsURL, err)
			return
		}
		queryString := u.Query()
		queryString.Add("--user-agent", RandomUserAgent())
		queryString.Add("--window-size", *a.session.Options.Resolution)
		queryString.Add("--ignore-certificate-errors", "true")
		if *a.session.Options.Proxy != "" {
			queryString.Add("--proxy-server", *a.session.Options.Proxy)
		}
		u.RawQuery = queryString.Encode()
		a.session.Out.Debug("[%s] remote chrome dev tools url: %s\n", a.ID(), u.String())
		rctx, cancelActx := chromedp.NewRemoteAllocator(context.Background(), u.String())
		defer cancelActx()

		actx = rctx

	} else {
		var opts = append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("window-size", *a.session.Options.Resolution),
			chromedp.Flag("user-agent", RandomUserAgent()),
		)

		if os.Geteuid() == 0 {
			opts = append(opts, chromedp.Flag("no-sandbox", true))
		}
		if *a.session.Options.Proxy != "" {
			opts = append(opts, chromedp.ProxyServer(*a.session.Options.Proxy))
		}

		ectx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()

		actx = ectx
	}

	nctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	ctx, cancel := context.WithTimeout(nctx, time.Duration(*a.session.Options.ScreenshotTimeout)*time.Millisecond)
	defer cancel()

	// capture screenshot of an element
	var buf []byte

	// capture entire browser viewport, returning png with quality=70
	if err := chromedp.Run(ctx, fullScreenshot(page.URL, int64(*a.session.Options.ImageQuality), &buf)); err != nil {
		a.session.Stats.IncrementScreenshotFailed()
		a.session.Out.Debug("[%s] Error: %v\n", a.ID(), err)
		if ctx.Err() == context.DeadlineExceeded {
			a.session.Out.Error("%s: screenshot timed out\n", page.URL)
			return
		}
		a.session.Out.Error("%s: screenshot failed: %s\n", page.URL, err)
		return
	}
	writePath := a.session.GetFilePath(filePath)
	if err := ioutil.WriteFile(writePath, buf, 0644); err != nil {
		a.session.Stats.IncrementScreenshotFailed()
		a.session.Out.Error("%s: screenshot write File %s error: %s\n", page.URL, writePath, err)
		return
	}

	a.session.Stats.IncrementScreenshotSuccessful()
	a.session.Out.Info("%s: %s\n", page.URL, Green("screenshot successful"))
	page.ScreenshotPath = filePath
	page.HasScreenshot = true
}

func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
