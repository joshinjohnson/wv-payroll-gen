package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Context struct {
}

func NewContext() Context {
	return Context{}
}

func (m *Context) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxMsg := newRequestCtx()
		ctxMsg.Set("user-agent", r.UserAgent())
		ctxMsg.Set("method", r.Method)
		ctxMsg.Set("path", r.URL.Path)
		ctxMsg.Set("remote", r.RemoteAddr)

		ctx := context.WithValue(context.Background(), "msg", ctxMsg)
		r = r.WithContext(ctx)

		defer func() {
			if rec := recover(); rec != nil {
				logrus.Errorf(fmt.Sprintf("panic: %s", ctxMsg.String()))
			}
		}()

		logrus.Infof(fmt.Sprintf("Incoming Request: %v", ctxMsg.String()))
		next.ServeHTTP(w, r)
	})
}

type requestCtx struct {
	Id     string
	Params map[string]string
}

func newRequestCtx() *requestCtx {
	return &requestCtx{
		Id:     uuid.New().String(),
		Params: make(map[string]string),
	}
}

func (c *requestCtx) Set(key string, value string) {
	c.Params[key] = value
}

func (c *requestCtx) Get(key string) string {
	if v, ok := c.Params[key]; ok {
		return v
	}
	return ""
}

func (c *requestCtx) String() string {
	msg := fmt.Sprintf("id=\"%s\"", c.Id)
	for k, v := range c.Params {
		msg += " " + fmt.Sprintf("%s=\"%s\"", k, v)
	}
	return msg
}
