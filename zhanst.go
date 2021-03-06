package zhanst

import (
	"net/http"
	"os"
	"sync"
)

type HandlerFunc func(*Context)

type HandlerChain []HandlerFunc

type Engine struct {
	RouterGroup
	trees methodTrees
	pool  sync.Pool
}

func New() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: make(HandlerChain, 0),
			basePath: "/",
			root:     true,
		},
		trees: make(methodTrees),
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() interface{} {
		return &Context{
			engine:   engine,
			Params:   make(Params, 0),
			handlers: make(HandlerChain, 0),
			key:      make(map[string]interface{}),
		}
	}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(os.Stdout))
	return engine
}

func (engine *Engine) Run(addr string) {
	http.ListenAndServe(addr, engine)
}

func (engine *Engine) Handle(method string, path string) {
	tree := engine.trees[method]
	handles, params := tree.getValue(path)
	for _, handle := range handles {
		c := Context{Params: params}
		handle(&c)
	}
}

func (engine *Engine) addRoute(method string, path string, handlers HandlerChain) {
	if _, ok := engine.trees[method]; !ok {
		root := new(treeNode)
		root.path = "/"
		root.children = make([]*treeNode, 0)
		root.handlers = make(HandlerChain, 0)
		engine.trees[method] = methodTree{
			method: method,
			root:   root,
		}
	}
	tree := engine.trees[method]
	tree.addRoute(path, handlers)
}

func (engine *Engine) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	c := engine.pool.Get().(*Context)
	c.Request = request
	c.Writer = responseWriter
	c.reset()
	engine.handleRequest(c)
	engine.pool.Put(c)
}

func (engine *Engine) handleRequest(c *Context) {
	method := c.Request.Method
	path := c.Request.URL.Path
	tree := engine.trees[method]
	handlers, params := tree.getValue(path)
	c.Params = params
	c.handlers = handlers
	c.Next()
}
