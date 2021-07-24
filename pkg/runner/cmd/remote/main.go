// Command remote is a chromedp example demonstrating how to connect to an
// existing Chrome DevTools instance using a remote WebSocket URL.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	devtoolsWsURL := flag.String("devtools-ws-url", "", "DevTools WebSsocket URL")
	flag.Parse()
	if *devtoolsWsURL == "" {
		*devtoolsWsURL = "ws://127.0.0.1:9222/"
		//log.Fatal("must specify -devtools-ws-url")
	}

	// create allocator context for use with creating a browser context later
	allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), *devtoolsWsURL)
	defer cancel()

	allocatorContext = context.Background()
	//ctxt, cancel := chromedp.NewExecAllocator(allocatorContext, []chromedp.ExecAllocatorOption{
	//	chromedp.Flag("headless", false),
	//}...)
	//defer cancel()

	allocatorContext, _ = chromedp.NewExecAllocator(allocatorContext, []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false),
	}...)
	// create context
	ctxt, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()

	// run task list
	var body string
	if err := chromedp.Run(ctxt,
		//chromedp.Flag("headless", false),
		//chromedp.Flag("hide-scrollbars", false),
		chromedp.Navigate("https://duckduckgo.com"),
		chromedp.WaitVisible("#logo_homepage_link"),
		chromedp.OuterHTML("html", &body),
		chromedp.Sleep(time.Second*5),
	); err != nil {
		log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
	}

	log.Println("Body of duckduckgo.com starts with:")
	log.Println(body[0:100])
}
