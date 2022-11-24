package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zdunecki/crawlerd/pkg/analyze"
	"golang.org/x/net/html"
	"net/http"
	"os"
)

// TODO: performances (zero-copy etc.)
func main() {
	u := "https://www.canva.com/newsroom/news"
	u = "https://livesession.io/blog"
	u = "https://livesession.io"
	u = "https://tech.onesignal.com/"
	u = "https://onesignal.com/blog/tag/development"
	u = "https://www.egnyte.com/blog/"

	{
		resp, err := http.Get(u)
		if err != nil {
			panic(err)
		}

		tree, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}

		w, err := os.Create("./original.html")
		if err != nil {
			panic(err)
		}

		if err := html.Render(w, tree); err != nil {
			panic(err)
		}
	}

	{
		tree, device, err := analyze.CrawlPage(u)
		if err != nil {
			panic(err)
		}
		doc := goquery.NewDocumentFromNode(tree)
		fmt.Println(analyze.IsArticle(device, doc))
	}
}
