package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")[1:]

	parts := make([]string, 0)
	for index, item := range vs {
		if item == "" {
			item = "*"
		}
		parts = append(parts, item)
		if item[0] == '*' {
			// // 或 /*xx/ 不被允许
			if len(vs) > index+1 {
				panic("illegal pattern")
			}
			break
		}
	}
	return parts
}

func parseUrlPath(path string) []string {
	if path == "" {
		path = "/"
	}
	vs := strings.Split(path, "/")
	last := vs[len(vs)-1:]

	res := make([]string, 0)
	for _, item := range vs[0 : len(vs)-1] {
		if item != "" {
			res = append(res, item)
		}
	}
	return append(res, last...)
}
func (r *router) addRoute(method, pattern string, handlerFunc HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handlerFunc
}
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parseUrlPath(path)
	_, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := r.roots[method].search(searchParts, 0)
	if n != nil {
		params := make(map[string]string)
		parts := parsePattern(n.n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n.n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])

	} else {
		c.handlers = append(c.handlers, func(ctx *Context) {
			ctx.String(http.StatusNotFound, "404 NOT FOUND Request path : %v", c.Req.URL)
		})
	}
	c.Next()
}
