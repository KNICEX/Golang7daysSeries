package gee

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{
		engine:      engine,
		middlewares: []HandlerFunc{Recovery(), Logger()},
	}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHtmlGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) Error(handler func(panicValue any)) {
	errorHandler := func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				handler(err)
			}
		}()
		c.Next()
	}
	// 替换默认错误处理中间件
	engine.middlewares = append([]HandlerFunc{errorHandler}, engine.middlewares[1:]...)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	// 处理重复/
	req.URL, _ = url.Parse(normalizeURL(req.URL.Path))
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func checkPattern(pattern string) {
	if len(pattern) == 0 || pattern[0] == ':' || pattern[0] == '*' {
		panic("Unsupported pattern")
	}
}

func normalizeURL(url string) string {
	if len(url) == 0 {
		return "/"
	}
	cleaned := path.Clean(url)
	if url[len(url)-1] == '/' && cleaned[len(cleaned)-1] != '/' {
		return cleaned + "/"
	}
	return cleaned
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	checkPattern(comp)
	// 去除可能重复的 //
	pattern := group.prefix + comp
	pattern = normalizeURL(pattern)
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	checkPattern(prefix)
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	group.engine.groups = append(group.engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) GET(pattern string, handlerFunc HandlerFunc) {
	group.addRoute(http.MethodGet, pattern, handlerFunc)
}

func (group *RouterGroup) POST(pattern string, handlerFunc HandlerFunc) {
	group.addRoute(http.MethodPost, pattern, handlerFunc)
}

func (group *RouterGroup) PUT(pattern string, handlerFunc HandlerFunc) {
	group.addRoute(http.MethodPut, pattern, handlerFunc)
}

func (group *RouterGroup) DELETE(pattern string, handlerFunc HandlerFunc) {
	group.addRoute(http.MethodDelete, pattern, handlerFunc)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	// 请求到来后 从url中裁剪掉参数1 即 /static/a -> /a 然后直接去 fs对应的根目录中查找 a
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(ctx *Context) {
		//file := ctx.Param("filepath")
		//if _, err := fs.Open(file); err != nil {
		//	ctx.Status(http.StatusNotFound)
		//	return
		//}
		fileServer.ServeHTTP(ctx.Writer, ctx.Req)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}
