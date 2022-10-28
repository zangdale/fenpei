package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var err404 = []byte(`{"err":"404"}`)
var errHttp = func(err error) []byte {
	if err == nil {
		return []byte(`{"err":""}`)
	}
	return []byte(fmt.Sprintf(`{"err":"%s"}`, err.Error()))
}

type handler struct {
	sync.RWMutex
	mapRouters    map[string]*to
	mapFileServer map[string]http.Handler
	*http.Client
}

func NewHandler(routers map[string]*to) *handler {
	var mapFileServer = make(map[string]http.Handler)
	for k, r := range routers {
		if r.isFile {
			mapFileServer[k] = http.FileServer(http.Dir(r.To))
		}
	}
	return &handler{
		RWMutex:       sync.RWMutex{},
		mapRouters:    routers,
		mapFileServer: mapFileServer,
		Client:        http.DefaultClient,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) == 1 {
		w.WriteHeader(404)
		w.Write(err404)
		return
	}
	index := strings.Index(path[1:], "/")
	if index == -1 {
		var k = path[1:]
		h.RLock()
		to, ok := h.mapRouters[k]
		h.RUnlock()
		if ok {
			if !to.isFile {
				var buf strings.Builder
				buf.WriteString(to.To)
				if r.URL.ForceQuery || r.URL.RawQuery != "" {
					buf.WriteByte('?')
					buf.WriteString(r.URL.RawQuery)
				}
				u := buf.String()
				h.http(u, w, r)
				return
			} else {
				h.file(k, to, w, r)
				return
			}
		} else {
			w.WriteHeader(404)
			w.Write(err404)
			return
		}
	} else {
		var k = path[1:index]
		h.RLock()
		to, ok := h.mapRouters[path[1:index]]
		h.RUnlock()
		if ok {
			if !to.isFile {
				var buf strings.Builder
				buf.WriteString(to.To)
				buf.WriteString(path[index:])
				if r.URL.ForceQuery || r.URL.RawQuery != "" {
					buf.WriteByte('?')
					buf.WriteString(r.URL.RawQuery)
				}
				u := buf.String()
				h.http(u, w, r)
				return
			} else {
				r.URL.Path = path[index:]
				h.file(k, to, w, r)
				return
			}
		} else {
			w.WriteHeader(404)
			w.Write(err404)
			return
		}
	}

}

func (h *handler) http(u string, w http.ResponseWriter, r *http.Request) {

	log.Info().Str("input", r.URL.String()).Str("to", u).Msg("http")

	defer r.Body.Close()
	req, err := http.NewRequest(r.Method, u, r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write(errHttp(err))
		return
	}
	req.Header = r.Header
	resp, err := h.Do(req)
	if err != nil {
		w.WriteHeader(500)
		w.Write(errHttp(err))
		return
	}
	w.WriteHeader(resp.StatusCode)
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

func (h *handler) file(k string, to *to, w http.ResponseWriter, r *http.Request) {
	log.Info().Str("input", r.URL.String()).Str("to", to.To).Msg("file")
	h.Lock()
	s, ok := h.mapFileServer[k]
	h.Unlock()
	if !ok {
		w.WriteHeader(404)
		w.Write(err404)
		return
	}
	s.ServeHTTP(w, r)
}

func (h *handler) UpMapRouters(mapRouters map[string]*to) {
	h.Lock()
	defer h.Unlock()
	h.mapRouters = mapRouters
}

func (h *handler) GetMapRouters() (mapRouters map[string]*to) {
	h.RLock()
	defer h.RUnlock()
	for k, v := range h.mapRouters {
		mapRouters[k] = v
	}
	return
}
