package runnerv1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/zdunecki/crawlerd/api"
	apiv1 "github.com/zdunecki/crawlerd/api/v1"
	"github.com/zdunecki/crawlerd/pkg/meta/metav1"
	"github.com/zdunecki/crawlerd/pkg/store"
	"github.com/zdunecki/crawlerd/pkg/util"
)

type v1 struct {
	addr string
	api  api.API

	cfg Config

	//apiClient apiv1.V1
	store store.Repository

	log *log.Entry
}

type V1 interface {
	Run(*metav1.RunnerUpCreate) (interface{}, error)
}

func New(addr string, store store.Repository, cfg Config) *v1 {
	l := log.WithField("service", "runner")
	r := chi.NewRouter() // TODO: router should be an abstraction

	//apiAddr := util.BaseAddr(cfg.APIURL)
	////apiClient, err := client.NewWithOpts(client.WithHTTPAddr(apiAddr))
	//if err != nil {
	//	l.Error(err)
	//	return nil
	//}

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

	// TODO:
	//v := viper.New()
	//v.AutomaticEnv()
	//viper.Unmarshal()

	v1 := &v1{
		addr: addr,
		api:  app,

		cfg: cfg,

		//apiClient: apiClient,
		store: store,

		log: l,
	}

	app.Post("/v1/run", v1.run)

	return v1
}

func (v1 *v1) ListenAndServe() {
	v1.log.Info("listening on: ", v1.addr)

	if err := http.ListenAndServe(v1.addr, v1.api.Handler()); err != nil {
		v1.log.Error(err)
	}
}

// TODO: ScrapeLinksPattern, FollowLinks
// TODO: ScrapeLinksPattern, FollowLinks on frontend side
// TODO: add runner start url to linker
func (v1 *v1) run(c api.Context) {
	req := &metav1.RunnerUpCreate{}
	if err := c.Bind(req); err != nil {
		v1.log.Error(err)

		c.InternalError().JSON(apiobjects.APIError{
			Type: apiv1.ErrorTypeInternal,
		})
		return
	}

	rs := v1.store.Runner()
	rqs := v1.store.RequestQueue()
	rfs := v1.store.RunnerFunctions()

	currentRunnerDepth := metav1.RunnerInitialDepth

	runID, err := rs.Create(c.RequestContext(), &metav1.RunnerCreate{
		RunAt:  util.NowInt(),
		Status: metav1.RunnerStatusQueued,
		Depth:  currentRunnerDepth,
		RunnerConfig: metav1.RunnerConfig{
			ScrapeLinksPattern: req.ScrapeLinksPattern,
			FollowLinks:        req.FollowLinks,
			MaxDepth:           req.MaxDepth,
		},
	})

	if err != nil {
		v1.log.Error(err)

		c.InternalError().JSON(apiobjects.APIError{
			Type: apiv1.ErrorTypeInternal,
		})
		return
	}

	//chromedp.NewRemoteAllocator(context.Background(), "WS_URL")

	// non-headless mode
	//ctx, cancel := chromedp.NewExecAllocator(c.RequestContext(), []chromedp.ExecAllocatorOption{
	//	chromedp.Flag("headless", false),
	//}...)
	//ctx, cancel = chromedp.NewContext(ctx)

	var chromeCtx context.Context
	var crawlFunction string

	{
		ctx, cancel := chromedp.NewContext(c.RequestContext())
		defer cancel()

		crawlFunction, err = rfs.GetByID(c.RequestContext(), req.ID)
		if err != nil {
			v1.log.Error(err)

			c.InternalError().JSON(apiobjects.APIError{
				Type: apiv1.ErrorTypeInternal,
			})
			return
		}

		chromeCtx = ctx

		// TODO: remove after debugging but console api will be useful in future releases
		gotException := make(chan bool, 1)
		chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
			switch ev := ev.(type) {
			case *runtime.EventConsoleAPICalled:
				fmt.Printf("* console.%s call:\n", ev.Type)
				for _, arg := range ev.Args {
					fmt.Printf("%s - %s\n", arg.Type, arg.Value)
				}
			case *runtime.EventExceptionThrown:
				// Since ts.URL uses a random port, replace it.
				s := ev.ExceptionDetails.Error()
				fmt.Printf("* %s\n", s)
				gotException <- true
			}
		})
	}

	// TODO: speedup (it should not run in pool or something else)
	crawl := func(pageURL string, depth uint) (interface{}, error) {
		var res interface{}

		windowVariables := map[string]string{
			"CRAWLERD_API_URL": v1.cfg.APIURL,
			"CRAWLERD_RUN_ID":  runID,
			"CRAWLERD_DEPTH":   fmt.Sprintf("%d", depth),
		}

		onLoadScript := ""
		onLoadScriptAtoms := make([]string, 0)
		for key, value := range windowVariables {
			if value == "" {
				continue
			}
			onLoadScriptAtoms = append(onLoadScriptAtoms, fmt.Sprintf(`window.%s='%s'`, key, value))
		}
		onLoadScript = strings.Join(onLoadScriptAtoms, "\n")

		if err := chromedp.Run(chromeCtx,
			chromedp.Navigate(pageURL),
			chromedp.Evaluate(onLoadScript, nil),
			chromedp.Evaluate(crawlFunction, &res, func(params *runtime.EvaluateParams) *runtime.EvaluateParams {
				return params.WithAwaitPromise(true)
			}),
		); err != nil {
			return nil, err
		}

		return res, nil
	}

	if _, err := crawl(req.URL, currentRunnerDepth); err != nil {
		v1.log.Error(err)

		c.InternalError().JSON(apiobjects.APIError{
			Type: apiv1.ErrorTypeInternal,
		})

		return
	}

	firstRunAndFinal := metav1.RunnerInitialDepth >= req.MaxDepth

	runSuccessed := func() error {
		// TODO: how to determine failed status (gotException inside crawl?)
		// TODO: timeout status
		return rs.UpdateByID(c.RequestContext(), runID, &metav1.RunnerPatch{
			EndAt:  util.NowInt(),
			Status: metav1.RunnerStatusSuccessed,
		})
	}

	if firstRunAndFinal {
		if err := runSuccessed(); err != nil {
			v1.log.Error(err)

			c.InternalError().JSON(apiobjects.APIError{
				Type: apiv1.ErrorTypeInternal,
			})
			return
		}
	} else {
		var deepCrawl func()

		deepCrawl = func() {
			if currentRunnerDepth >= req.MaxDepth {
				if err := runSuccessed(); err != nil {
					v1.log.Error(err)

					c.InternalError().JSON(apiobjects.APIError{
						Type: apiv1.ErrorTypeInternal,
					})
					return
				}

				return
			}

			deeperQueues, err := rqs.List(c.RequestContext(), &metav1.RequestQueueListFilter{
				RunID: &metav1.StringFilter{
					Is: runID,
				},
				Depth: &metav1.UintFilter{
					Is: currentRunnerDepth,
				},
				Status: &metav1.StringFilter{
					Is: string(metav1.RequestQueueStatusQueued),
				},
			})
			if err != nil {
				v1.log.Error(err)

				c.InternalError().JSON(apiobjects.APIError{
					Type: apiv1.ErrorTypeInternal,
				})
				return
			}

			currentRunnerDepth += 1

			// TODO: distributed request queue
			// TODO: bfs
			for _, q := range deeperQueues {
				// TODO: retry
				_, err := crawl(q.URL, currentRunnerDepth)

				var requestQueueStatus metav1.RequestQueueStatus

				if err != nil {
					v1.log.Error(err)
					requestQueueStatus = metav1.RequestQueueStatusFailed
				} else {
					requestQueueStatus = metav1.RequestQueueStatusSuccessed
				}

				if err := rqs.UpdateByID(context.Background(), q.ID, &metav1.RequestQueuePatch{
					Status: requestQueueStatus,
				}); err != nil {
					v1.log.Error(err)
				}
			}

			if err := rs.UpdateByID(context.Background(), runID, &metav1.RunnerPatch{
				Depth: currentRunnerDepth, // TODO: if goes distributed should increment be atomic
			}); err != nil {
				v1.log.Error(err)
			}

			deepCrawl()
		}

		deepCrawl()
	}

	c.JSON(map[string]interface{}{
		"status": "ok",
	})
	return
}
