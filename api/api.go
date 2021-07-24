package api

import (
	"net/http"
)

type RouteMethods interface {
	Get(string, http.HandlerFunc)
	Put(string, http.HandlerFunc)
	Post(string, http.HandlerFunc)
	Patch(string, http.HandlerFunc)
	Delete(string, http.HandlerFunc)
}

type Router interface {
	ServeHTTP(http.ResponseWriter, *http.Request)

	RouteMethods
}

type API interface {
	Handler() http.Handler

	Get(string, ContextFn)
	Put(string, ContextFn)
	Post(string, ContextFn, ...MiddleWare)
	Patch(string, ContextFn, ...MiddleWare)
	Delete(string, ContextFn)
}

type api struct {
	router Router
}

func New(router Router) API {
	return &api{
		router: router,
	}
}

func (a *api) Get(s string, context ContextFn) {
	a.router.Get(s, func(writer http.ResponseWriter, request *http.Request) {
		context(ctx{
			writer:  writer,
			request: request,
		})
	})
}

func (a *api) Post(s string, context ContextFn, middlewares ...MiddleWare) {
	a.router.Post(s, func(writer http.ResponseWriter, request *http.Request) {
		for _, m := range middlewares {
			m(writer, request)
		}

		context(ctx{
			writer:  writer,
			request: request,
		})
	})
}

func (a *api) Put(s string, context ContextFn) {
	a.router.Put(s, func(writer http.ResponseWriter, request *http.Request) {
		context(ctx{
			writer:  writer,
			request: request,
		})
	})
}

func (a *api) Patch(s string, context ContextFn, middlewares ...MiddleWare) {
	a.router.Patch(s, func(writer http.ResponseWriter, request *http.Request) {
		for _, m := range middlewares {
			m(writer, request)
		}

		context(ctx{
			writer:  writer,
			request: request,
		})
	})
}

func (a *api) Delete(s string, context ContextFn) {
	a.router.Delete(s, func(writer http.ResponseWriter, request *http.Request) {
		context(ctx{
			writer:  writer,
			request: request,
		})
	})
}

func (a *api) Handler() http.Handler {
	return a.router
}
