package wpsev

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/quic-go/quic-go/http3"
)

const (
	HTTP1 = iota + 1
	HTTP2
	HTTP3
)

type Server struct {
	server      *http.Server
	server1     *http.Server
	server2     *http.Server
	server3     *http3.Server
	patterns    map[string]map[string][]http.Handler
	init        bool
	tlsClose    bool
	mutex       sync.Mutex
	versionHttp uint32
}

func NewServer(server *http.Server, versionHttp uint32) *Server {

	return &Server{
		server:      server,
		patterns:    make(map[string]map[string][]http.Handler),
		mutex:       sync.Mutex{},
		versionHttp: versionHttp,
	}
}

func (s *Server) Start(addr string, port int) error {
	s.initRouter()

	if s.server.Addr == "" {
		s.server.Addr = fmt.Sprintf("%s:%d", addr, port)
	}
	return s.server.ListenAndServe()
}

func (s *Server) StartTLS(addr string, port int, certFile, keyFile string) error {
	s.initRouter()
	addr = fmt.Sprintf("%s:%d", addr, port)
	var err error

	switch s.versionHttp {
	case HTTP1:
		s.server1 = &http.Server{
			Addr:    addr,
			Handler: s.server.Handler,
		}
		err = s.server1.ListenAndServeTLS(certFile, keyFile)
	case HTTP2:
		s.server2 = &http.Server{
			Addr: addr,
		}
		err = http2.ConfigureServer(s.server2, &http2.Server{})
		if err != nil {
			break
		}
		err = s.server2.ListenAndServeTLS(certFile, keyFile)
	case HTTP3:
		s.server3 = &http3.Server{}
		err = s.ListenAndServeHttp3(addr, certFile, keyFile, s.server.Handler)
	default:
		err = fmt.Errorf("Wrong http protocol version: %d\n", s.versionHttp)
	}

	if err == nil {
		s.tlsClose = true
	}
	return err
}

func (s *Server) Stop() error {
	if !s.tlsClose {
		return s.server.Close()
	}

	var err error
	switch s.versionHttp {
	case HTTP1:
		err = s.server1.Close()
	case HTTP2:
		err = s.server2.Close()
	case HTTP3:
		err = s.server3.Close()
	}

	return err
}

func (s *Server) initRouter() {
	s.mutex.Lock()
	if s.init {
		return
	}
	s.init = true

	if _, ok := s.patterns["/"]; !ok {
		s.AddRouter(http.MethodOptions, "/", notFound)
	}

	for k := range s.patterns {
		http.Handle(k, s.getHandlers())
	}

	s.mutex.Unlock()
}

/*
AddRouter method accepts pattern, request type and chain handlers.
When specifying a pattern, you must ensure that there are no duplicates, otherwise the service will not start.
server.AddRouter
/test/:id/*file and /test/:name/*path are considered duplicates.
*/
func (s *Server) AddRouter(method, pattern string, handlers ...interface{}) {

	hs := make([]http.Handler, 0)

	for _, handler := range handlers {
		switch t := handler.(type) {
		case http.Handler:
			hs = append(hs, t)
		case func(http.ResponseWriter, *http.Request):
			hs = append(hs, http.HandlerFunc(t))
		default:
			panic(fmt.Errorf("error handler type %v\n", t))
		}
	}

	if err := s.checkPattern(method, pattern); err != nil {
		panic(err)
	}

	_, ok := s.patterns[pattern]
	if !ok {
		m := make(map[string][]http.Handler)
		m[method] = hs
		s.patterns[pattern] = m
		return
	}

	s.patterns[pattern][method] = hs
}

func (s *Server) getHandlers() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		BreakConn(r, false)

		if vh, ok := s.patterns[r.URL.Path][r.Method]; ok {
			for _, h := range vh {
				if checkBreak(r) {
					return
				}
				h.ServeHTTP(w, r)
			}
			return
		}

		p := s.searchPattern(r.URL.Path)
		if p == "" {
			SendMsg(w, "not found", http.StatusNotFound)
			return
		}

		params := getParseUrl(r.URL.Path)
		for i, elem := range getParseUrl(p) {
			if elem == "" {
				continue
			}

			if elem[0] == ':' {
				setParam(r, elem[1:], params[i])
				continue
			}

			if elem[0] == '*' {
				setParam(r, elem[1:], filepath.Join(params[i:]...))
				break
			}
		}

		setParam(r, "pattern", p)

		handlers, ok := s.patterns[p][r.Method]
		if !ok {
			SendMsg(w, "not found", http.StatusNotFound)
			return
		}

		for _, h := range handlers {
			if checkBreak(r) {
				return
			}
			h.ServeHTTP(w, r)
		}
	}
}

func setParam(r *http.Request, key, value string) {
	*r = *r.WithContext(context.WithValue(r.Context(), key, value))
}

func GetParam(r *http.Request, key string) string {
	v, ok := r.Context().Value(key).(string)
	if !ok {
		return ""
	}
	return v
}

func SendMsg(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func BreakConn(r *http.Request, interrupt bool) {
	*r = *r.WithContext(context.WithValue(r.Context(), "break", interrupt))
}

func checkBreak(r *http.Request) bool {
	flag, ok := r.Context().Value("break").(bool)
	if !ok {
		return false
	}
	return flag
}

func (s *Server) ReloadTLS(certFile, keyFile string) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for range c {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		switch s.versionHttp {
		case HTTP1:
			s.server1.TLSConfig = tlsConfig
		case HTTP2:
			s.server2.TLSConfig = tlsConfig
		case HTTP3:
			s.server3.TLSConfig = tlsConfig
		}

	}
	return nil
}

func (s *Server) ListenAndServeHttp3(addr, certFile, keyFile string, handler http.Handler) error {

	var err error
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: certs,
	}

	if addr == "" {
		addr = ":https"
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	s.server3.TLSConfig = config
	s.server3.Handler = s.server.Handler

	hErr := make(chan error)
	qErr := make(chan error)
	go func() {
		hErr <- http.ListenAndServeTLS(addr, certFile, keyFile, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.server3.SetQuicHeaders(w.Header())
			handler.ServeHTTP(w, r)
		}))
	}()
	go func() {
		qErr <- s.server3.Serve(udpConn)
	}()

	select {
	case err := <-hErr:
		s.server3.Close()
		return err
	case err := <-qErr:
		return err
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
