olx-building-plot-ads:
	curl http://localhost:9998/extract \
      -X POST \
      -H "Content-Type: application/json" \
      --data '{"url": "https://www.olx.pl/nieruchomosci/dzialki/$(city)/?search[dist]=$(distance)&page=$(page)", "js_file": "plugins/olx-ads-extract.js"}'

nike-get-products:
	curl http://localhost:9998/extract \
      -X POST \
      -H "Content-Type: application/json" \
      --data '{"url": "https://www.nike.com/pl/w/mezczyzni-buty-nik1zy7ok", "js_file": "plugins/nike-scroll-extract.js"}'
