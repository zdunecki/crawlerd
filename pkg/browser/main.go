package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-rod/rod"
)

// find closest relation between links, box, image and text in DOM
// if link has image - +1
// if link has image and closest text is very close from bottom or right - +1
// if document.body.innerText is engineering - +1

// link can be:
// * parent - image or text are children
// * child - image can be sibling but something is in root

// TODO: faster serialization
var jsScriptSerializeDOM = `() => {
    function search(el, parent, i) {
    	if (!el) {
            return
    	}
        
        i++
        
        // TODO: computed styles
        const data = {
            i,
            nT: el.nodeType,
			n: el.nodeName,
			tN: "",
			a: {},
			tC: "",
			cN: [],
			r: el.getBoundingClientRect ? el.getBoundingClientRect() : undefined
        }
        
        for (const a of document.body.attributes) {
            data.a[a.name] = a.value
		}	
        
        switch (el.nodeType) {
            case Node.COMMENT_NODE:
            case Node.TEXT_NODE:
                data.tC = el.textContent
                break
        }
   
                
        if (!el.children) {
            return
        }
        
        for (const e of el.children) {
            const child = search(e, data, i)
            data.cN.push(child)
        }
        
        return data
	}
	
    return JSON.stringify(search(document, null, 0))
}
`

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`

	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`

	Left  float64 `json:"left"`
	Right float64 `json:"right"`
}

type DOM struct {
	ID          int                    `json:"i"`
	Type        int                    `json:"nT"`
	Name        string                 `json:"n"`
	TagName     string                 `json:"tN"`
	Attrs       map[string]interface{} `json:"a"`
	TextContent string                 `json:"tC"`
	ChildNodes  []*DOM                 `json:"cN"`
	Rect        *Rect                  `json:"r"`
}

func main() {
	page := rod.New().MustConnect().MustPage("https://www.wikipedia.org/").MustWaitLoad()

	data := page.MustEval(jsScriptSerializeDOM).Str()

	var dom *DOM
	if err := json.Unmarshal([]byte(data), &dom); err != nil {
		panic(err)
	}

	fmt.Println(dom)
}
