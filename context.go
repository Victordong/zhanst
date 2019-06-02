package zhanst

import (
	"math"
	"net/http"
)

type Param struct {
	Key   string
	Value string
}

const (
	abortIndex int8 = math.MaxInt8 / 2
)

type Params []Param

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	Params Params

	index    int8
	handlers HandlerChain

	engine *Engine

	status int

	key map[string]interface{}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) Next() {
	c.index++
	for handlersLen := int8(len(c.handlers)); c.index < handlersLen; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Status() int {
	return c.status
}

func (c *Context) Render(code int, r Render) {
	c.status = code
	if err := r.Render(c.Writer, code); err != nil {
		panic(err)
	}
}

func (c *Context) JSON(code int, data interface{}) {
	c.Render(code, JSON{Data: data})
}

func (c *Context) reset() {
	c.Params = c.Params[0:0]
	c.index = -1
	c.handlers = c.handlers[0:0]
	c.key = nil
}

func (c *Context) Bind(obj interface{}) error {
	if err := Bind(c.Request, obj); err != nil {
		return err
	}
	return nil
}

func (c *Context) SetValue(key string, value interface{}) {
	c.key[key] = value
}

func (c *Context) GetValue(key string) interface{} {
	if _, ok := c.key[key]; !ok {
		return nil
	}
	return c.key[key]
}
