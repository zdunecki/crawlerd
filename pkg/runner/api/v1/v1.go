package v1

import (
	"fmt"
	"net/http"
	"strings"

	"crawlerd/api"
	apiv1 "crawlerd/api/v1"
	"crawlerd/api/v1/client"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/runner"
	"crawlerd/pkg/util"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

type v1 struct {
	addr string
	api  api.API

	cfg Config

	apiClient apiv1.V1
	store     runner.Store

	log *log.Entry
}

type V1 interface {
	Run(*metav1.RunnerUpCreate) (interface{}, error)
}

func New(addr string, store runner.Store, cfg Config) *v1 {
	l := log.WithField("service", "runner")
	r := chi.NewRouter() // TODO: router should be an abstraction

	apiAddr := util.BaseAddr(cfg.APIURL)
	apiClient, err := client.NewWithOpts(client.WithHTTPAddr(apiAddr))
	if err != nil {
		l.Error(err)
		return nil
	}

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

		apiClient: apiClient,
		store:     store,

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

// TODO: update RunnerStatus
func (v1 *v1) run(c api.Context) {
	req := &metav1.RunnerUpCreate{}
	if err := c.Bind(req); err != nil {
		v1.log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	rs := v1.store.Runner()

	runID, err := rs.Create(c.RequestContext(), &metav1.RunnerCreate{
		RunAt:  util.NowInt(),
		Status: metav1.RunnerStatusQueued,
	})

	if err != nil {
		v1.log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	//chromedp.NewRemoteAllocator(context.Background(), "WS_URL")

	// non-headless mode
	//ctx, cancel := chromedp.NewExecAllocator(c.RequestContext(), []chromedp.ExecAllocatorOption{
	//	chromedp.Flag("headless", false),
	//}...)
	//ctx, cancel = chromedp.NewContext(ctx)

	ctx, cancel := chromedp.NewContext(c.RequestContext())
	defer cancel()

	fn, err := v1.store.Functions().GetByID(c.RequestContext(), req.ID)
	if err != nil {
		v1.log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	// TODO: remove after debugging but console api will be useful in future releases
	gotException := make(chan bool, 1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
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

	windowVariables := map[string]string{
		"CRAWLERD_API_URL": v1.cfg.APIURL,
		"CRAWLERD_RUN_ID":  runID,
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

	var res interface{}
	if err := chromedp.Run(ctx,
		chromedp.Navigate(req.URL),
		chromedp.Evaluate(onLoadScript, nil),
		chromedp.Evaluate(fn, &res, func(params *runtime.EvaluateParams) *runtime.EvaluateParams {
			return params.WithAwaitPromise(true)
		}),
	); err != nil {
		v1.log.Error(err)
		c.InternalError().JSON("something went wrong")
		return
	}

	// TODO: how to determine failed status
	// TODO: timeout status
	if err := rs.UpdateByID(c.RequestContext(), runID, &metav1.RunnerPatch{
		EndAt:  util.NowInt(),
		Status: metav1.RunnerStatusSuccessed,
	}); err != nil {
		v1.log.Error(err)

		c.InternalError().JSON("something went wrong")
		return
	}

	c.JSON(res)
	return
}
