package base

import (
    "net/http"
    "sync"
)

type Loader func(configFile string)

type LoaderRegistry struct {
    sync.RWMutex
    loaders []Loader
}

func NewLoaderRegistry() *LoaderRegistry {
    return &LoaderRegistry{loaders: make([]Loader, 0)}
}

func (r *LoaderRegistry) Register(loader Loader) {
    r.Lock()
    defer r.Unlock()
    if nil == loader {
        return
    }
    r.loaders = append(r.loaders, loader)
}

/********************************************************************************/

type Handler func(mux *http.ServeMux)

type HandlerRegistry struct {
    sync.RWMutex
    handlers []Handler
}

func NewHandlerRegistry() *HandlerRegistry {
    return &HandlerRegistry{handlers: make([]Handler, 0)}
}

func (r *HandlerRegistry) Register(handler Handler) {
    r.Lock()
    defer r.Unlock()
    if nil == handler {
        return
    }
    r.handlers = append(r.handlers, handler)
}

/********************************************************************************/

var (
    loaders = NewLoaderRegistry()
    handlers  = NewHandlerRegistry()
)

func RegisterLoader(loader Loader) {
    loaders.Register(loader)
}

func RegisterHandler(handler Handler) {
    handlers.Register(handler)
}

func Load() {
    for _, loader := range loaders.loaders {
        loader(configFile)
    }
}

func Handle(mux *http.ServeMux) {
    for _, handler := range handlers.handlers {
        handler(mux)
    }
}
