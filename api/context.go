package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ContextFn func(ctx Context)

type Context interface {
	Created() Context
	NoContent() Context
	BadRequest() Context
	NotFound() Context
	RequestEntityTooLarge() Context
	InternalError() Context

	Bind(interface{}) error
	Request() *http.Request

	ParamInt(key string) (int, error)

	JSON(interface{}) error

	RequestContext() context.Context
}

type ctx struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (c ctx) Created() Context {
	c.writer.WriteHeader(http.StatusCreated)
	return c
}

func (c ctx) NoContent() Context {
	c.writer.WriteHeader(http.StatusNoContent)
	return c
}

func (c ctx) BadRequest() Context {
	c.writer.WriteHeader(http.StatusBadRequest)
	return c
}

func (c ctx) NotFound() Context {
	c.writer.WriteHeader(http.StatusNotFound)
	return c
}

func (c ctx) RequestEntityTooLarge() Context {
	c.writer.WriteHeader(http.StatusRequestEntityTooLarge)
	return c
}

func (c ctx) InternalError() Context {
	c.writer.WriteHeader(http.StatusInternalServerError)
	return c
}

func (c ctx) Bind(i interface{}) error {
	if err := json.NewDecoder(c.request.Body).Decode(&i); err != nil {
		return err
	}

	return nil
}

func (c ctx) Request() *http.Request {
	return c.request
}

func (c ctx) ParamInt(key string) (int, error) {
	urlID := chi.URLParam(c.request, key)
	i, err := strconv.Atoi(urlID)

	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c ctx) JSON(i interface{}) error {
	c.writer.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	_, err = c.writer.Write(b)

	return err
}

func (c ctx) RequestContext() context.Context {
	return c.request.Context()
}
