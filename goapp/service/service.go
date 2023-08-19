//go:build !wasm

package service

import (
	"context"
	"errors"
	"fmt"
	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/goappnew/goapp"
	"github.com/mlctrez/goappnew/goapp/compo"
	"github.com/mlctrez/servicego"
	"io/fs"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func Entry() {
	compo.Routes()
	servicego.Run(&Service{})
}

var _ servicego.Service = (*Service)(nil)

const WasmHeader = "Wasm-Content-Length"

type Service struct {
	servicego.Defaults
	listener      net.Listener
	handler       *app.Handler
	engine        *gin.Engine
	server        *http.Server
	staticHandler http.Handler
	isDev         bool
	mappings      map[string]gin.HandlerFunc
}

func (s *Service) Start(_ service.Service) (err error) {

	s.isDev = os.Getenv("DEV") != ""
	s.mappings = make(map[string]gin.HandlerFunc)
	s.staticHandler = http.FileServer(http.FS(goapp.WebFs))

	steps := []func() error{
		s.setupHandler, s.setupEngine,
		s.goAppMappings, s.staticMappings, s.mappingsToGin,
		s.setupApiEndpoints, s.listen,
	}

	for _, step := range steps {
		if err = step(); err != nil {
			return err
		}
	}
	go s.runHttpServer()

	return nil
}

func (s *Service) setupEngine() (err error) {
	if !s.isDev {
		gin.SetMode(gin.ReleaseMode)
	}

	s.engine = gin.New()
	// required for go-app to work correctly
	s.engine.RedirectTrailingSlash = false

	middleware := []gin.HandlerFunc{gin.Recovery()}
	if s.isDev {
		config := gin.LoggerConfig{SkipPaths: []string{"/app-worker.js"}}
		middleware = append(middleware, gin.LoggerWithConfig(config))
	}
	if os.Getenv("GOAPP_USE_COMPRESSION") == "" {
		s.engine.Use(middleware...)
		return nil
	}

	var wasmSize int64
	if wasmSize, err = s.wasmSize(); err != nil {
		return err
	}
	s.handler.WasmContentLengthHeader = WasmHeader
	wasmHeaderHandler := (&fixedHeader{key: WasmHeader, value: fmt.Sprintf("%d", wasmSize)}).HandlerFunc
	middleware = append(middleware, wasmHeaderHandler, brotli.Brotli(brotli.DefaultCompression))

	s.engine.Use(middleware...)
	return nil
}

func (s *Service) wasmSize() (wasmSize int64, err error) {
	var wasmFile fs.File
	if wasmFile, err = goapp.WebFs.Open("web/app.wasm"); err != nil {
		return 0, err
	}
	defer func() { _ = wasmFile.Close() }()

	var stat fs.FileInfo
	if stat, err = wasmFile.Stat(); err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func (s *Service) setupHandler() (err error) {
	s.handler = &app.Handler{Env: make(app.Environment)}
	s.handler.Scripts = make([]string, 0)
	s.handler.Styles = make([]string, 0)
	if s.isDev {
		s.handler.AutoUpdateInterval = time.Second * 3
		s.handler.Version = ""
		s.handler.Env["DEV"] = "1"
	} else {
		s.handler.AutoUpdateInterval = time.Hour
		s.handler.Version = fmt.Sprintf("%s@%s", goapp.Version, goapp.Commit)
	}

	return nil
}

func (s *Service) goAppMappings() (err error) {
	goAppHandlerFunc := (&goAppHandler{handler: s.handler}).HandlerFunc
	s.mappings["/"] = goAppHandlerFunc
	s.mappings["/app.js"] = goAppHandlerFunc
	s.mappings["/app-worker.js"] = goAppHandlerFunc
	s.mappings["/wasm_exec.js"] = goAppHandlerFunc
	s.mappings["/app.css"] = goAppHandlerFunc
	s.mappings["/manifest.webmanifest"] = goAppHandlerFunc
	s.mappings["/manifest.json"] = goAppHandlerFunc
	return nil
}

func (s *Service) staticMappings() (err error) {
	var dir []fs.DirEntry
	if dir, err = goapp.WebFs.ReadDir("web"); err != nil {
		return err
	}

	for _, entry := range dir {
		name := entry.Name()
		path := fmt.Sprintf("/%s", name)
		if _, exists := s.mappings[path]; exists {
			s.mappings[path] = (&staticRemap{name: name, httpHandler: s.staticHandler}).HandlerFunc
		} else {
			if strings.HasSuffix(name, ".js") {
				s.handler.Scripts = append(s.handler.Scripts, fmt.Sprintf("/web/%s", name))
			}
			if strings.HasSuffix(name, ".css") {
				s.handler.Styles = append(s.handler.Styles, fmt.Sprintf("/web/%s", name))
			}
		}
	}

	s.mappings["/web/:path"] = (&staticHandler{httpHandler: s.staticHandler}).HandlerFunc

	return nil
}

func (s *Service) mappingsToGin() (err error) {
	var sortedMappings []string
	for k := range s.mappings {
		sortedMappings = append(sortedMappings, k)
	}

	sort.Strings(sortedMappings)
	for _, k := range sortedMappings {
		s.engine.GET(k, s.mappings[k])
	}

	return nil
}

func (s *Service) Stop(_ service.Service) (err error) {
	if s.server != nil {
		stopContext, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		err = s.server.Shutdown(stopContext)
		if errors.Is(err, context.Canceled) {
			os.Exit(-1)
		}
		_ = s.Log().Info("http.Server.Shutdown success")
	}
	return
}

func (s *Service) setupApiEndpoints() (err error) {
	// setup additional api endpoints here
	return
}

func ListenAddress() string {
	if address := os.Getenv("ADDRESS"); address != "" {
		return address
	}

	if port := os.Getenv("PORT"); port == "" {
		return "localhost:8080"
	} else {
		return "localhost:" + port
	}
}

func (s *Service) listen() (err error) {
	address := ListenAddress()
	if s.listener, err = net.Listen("tcp4", address); err != nil {
		return err
	}
	return nil
}

func (s *Service) runHttpServer() {
	s.server = &http.Server{Handler: s.engine}

	//goland:noinspection ALL
	addressString := s.listener.Addr().String()
	_ = s.Log().Infof("listening on http://%s\n", addressString)

	var serveErr error
	if strings.HasSuffix(addressString, ":443") {
		serveErr = s.server.ServeTLS(s.listener, "cert.pem", "cert.key")
	} else {
		serveErr = s.server.Serve(s.listener)
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		_ = s.Log().Error(serveErr)
	}
}
