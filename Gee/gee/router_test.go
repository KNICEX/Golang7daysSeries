package gee

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
	ok = reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
	ok = reflect.DeepEqual(parsePattern("/p/:name/b"), []string{"p", ":name", "b"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
	ok = reflect.DeepEqual(parsePattern("/p/:name/"), []string{"p", ":name", "*"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
	ok = reflect.DeepEqual(parsePattern("/p/*name"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
	ok = reflect.DeepEqual(parsePattern("/"), []string{"*"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}
func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/a", nil)
	r.addRoute("GET", "/a/b", nil)
	r.addRoute("GET", "/:a/b", nil)
	r.addRoute("GET", "/:a/*b", nil)
	r.addRoute("GET", "/:a/b/", nil)
	r.addRoute("GET", "/:a/:b/", nil)
	return r
}
func TestGetRoute(t *testing.T) {
	r := newTestRouter()

	getRouteItemTest("/a", "/a", r, t)
	getRouteItemTest("/b", "/", r, t)
	getRouteItemTest("/a/b", "/a/b", r, t)
	getRouteItemTest("/b/b", "/:a/b", r, t)
	getRouteItemTest("/a/b/", "/:a/b/", r, t)
	getRouteItemTest("/a/dd", "/:a/*b", r, t)
	getRouteItemTest("/a/b/c", "/:a/b/", r, t)
	getRouteItemTest("/a/c/c", "/:a/:b/", r, t)

}

func getRouteItemTest(url, target string, r *router, t *testing.T) map[string]string {
	n, params := r.getRoute("GET", url)
	if n == nil {
		t.Fatal("not found")
	}
	fmt.Printf("%s -> %s %v \n", url, n.pattern, params)
	if n.pattern != target {
		t.Fatal("not match")
	}
	return params

}

func TestPath(t *testing.T) {
	//str := "//a/"
	//fmt.Println(path.Clean(str))

	URL, _ := url.Parse("//a/b")
	fmt.Println(URL.Path)
}
