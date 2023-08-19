//go:build !wasm

package service

import (
	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"net/http"
)

type staticHandler struct {
	httpHandler http.Handler
}

func (sh *staticHandler) HandlerFunc(c *gin.Context) {
	sh.httpHandler.ServeHTTP(c.Writer, c.Request)
}

type staticRemap struct {
	name        string
	httpHandler http.Handler
}

func (sr *staticRemap) HandlerFunc(c *gin.Context) {
	c.Request.URL.Path = "/web/" + sr.name
	sr.httpHandler.ServeHTTP(c.Writer, c.Request)
}

type goAppHandler struct {
	handler *app.Handler
}

func (ga *goAppHandler) HandlerFunc(c *gin.Context) {
	ga.handler.ServeHTTP(c.Writer, c.Request)
}

type fixedHeader struct {
	key   string
	value string
}

func (ws *fixedHeader) HandlerFunc(c *gin.Context) {
	c.Writer.Header().Set(ws.key, ws.value)
	c.Next()
}
