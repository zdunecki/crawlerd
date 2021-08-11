package v1

import (
	"net/http"

	"crawlerd/api"
	"crawlerd/pkg/runner/storage"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

type v1 struct {
	addr string
	api  api.API

	storage storage.Storage

	log *log.Entry
}

func New(addr string, storage storage.Storage) *v1 {
	r := chi.NewRouter()

	// TODO: cors config
	r.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	app := api.New(r)

	v1 := &v1{
		addr: addr,
		api:  app,

		storage: storage,

		log: log.WithField("service", "runner"),
	}

	app.Post("/v1/extract", v1.extract)

	return v1
}

func (v1 *v1) ListenAndServe() {
	v1.log.Info("listening on: ", v1.addr)

	if err := http.ListenAndServe(v1.addr, v1.api.Handler()); err != nil {
		log.Error(err)
	}
}

func (v1 *v1) extract(c api.Context) {
	req := &RequestExtract{}
	if err := c.Bind(req); err != nil {
		log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	//chromedp.NewRemoteAllocator(context.Background(), "WS_URL")
	//ctx, _ := chromedp.NewExecAllocator(c.RequestContext(), []chromedp.ExecAllocatorOption{
	//	//chromedp.Flag("headless", true),
	//}...)

	ctx, cancel := chromedp.NewContext(c.RequestContext())
	defer cancel()

	fn, err := v1.storage.Functions().Get(c.RequestContext(), req.ID)
	if err != nil {
		log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	var res interface{}
	if err := chromedp.Run(ctx,
		chromedp.Navigate(req.URL),
		//chromedp.Evaluate(scriptContent, &res),
		chromedp.Evaluate(fn, &res, func(params *runtime.EvaluateParams) *runtime.EvaluateParams {
			return params.WithAwaitPromise(true)
		}),
	); err != nil {
		log.Error(err)
		c.InternalError().JSON("something went wrong")
		return
	}

	c.JSON(res)
	return
}
