package analyze

import (
	"encoding/json"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"golang.org/x/net/html"
	"os"
)

// TODO: faster serialization
// TODO: serialization should be done via browser engine itself? (without javascript)
var jsScriptSerializeDOM = `() => {
    function dynamicCssContent(el) {
        const cssRules = el && el.sheet && el.sheet.cssRules
        
        if (!cssRules) {
            return ""
        }

		let text = ""
		
		for (const rule of cssRules) {
			text += rule.cssText + " \n"
		}
        
        return text
    }
    
    function search(el, parent, i) {
    	if (!el) {
            return
    	}
        
        i++
        
        // TODO: computed styles
        const data = {
            i,
            nT: el.nodeType,
			tN: el.nodeName,
			a: {},
			tC: "",
			cN: [],
			r: el.getBoundingClientRect ? el.getBoundingClientRect() : undefined
        }
        
        if (el.attributes) {
			for (const a of el.attributes) {
				data.a[a.name] = a.value
			}	
        }
        
        const backgroundImage = (el && el.computedStyleMap && el.computedStyleMap().get("background-image").toString()) || ""
               
        if (backgroundImage.includes("url")) {
          	data.a["crawlerd-is-image"] = "1"
        }
        
        switch (el.nodeType) {
            case Node.COMMENT_NODE:
            case Node.TEXT_NODE:
				data.tC = el.textContent
                
			   	if (!data.tC) {
			   		data.tC = dynamicCssContent(parent)
			   	}
               
               	break
        }
   
                
        if (!el.childNodes) {
            return
        }
        
        for (const e of el.childNodes) {
            const child = search(e, el, i)
            data.cN.push(child)
        }
        
        return data
	}
	
    return JSON.stringify(search(document, null, 0))
}
`

func CrawlPage(u string) (*html.Node, devices.Device, error) {
	device := devices.LaptopWithMDPIScreen.Landescape()
	device.Screen.Horizontal.Height = 1329
	page := rod.New().
		DefaultDevice(device).
		MustConnect().
		MustPage(u).
		MustWaitLoad()

	jsDOMString := page.MustEval(jsScriptSerializeDOM).Str()

	var jsDOM *DOM
	if err := json.Unmarshal([]byte(jsDOMString), &jsDOM); err != nil {
		return nil, device, err
	}

	//tree := buildHTMLTree(jsDOM, nil) method 1
	tree := buildVisibleHTMLTree(device, jsDOM, nil) // method 2

	w, err := os.Create("./converted.html")
	if err != nil {
		return nil, device, err
	}

	if err := html.Render(w, tree); err != nil {
		return nil, device, err
	}

	return tree, device, nil
}
