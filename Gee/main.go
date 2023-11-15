package main

import (
	"fmt"
	"gee"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	r := gee.New()
	sse(r)
	r.Run(":8080")
}

func sse(r *gee.Engine)  {

}

func testServer(r *gee.Engine)  {
	r.GET("/:a", func(c *gee.Context) {
		c.JSON(http.StatusOK, c.Param("a"))
	})
	r.GET("/a/*c", func(ctx *gee.Context) {
		ctx.JSON(http.StatusOK, ctx.Param("c"))
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/haha", func(ctx *gee.Context) {
			ctx.JSON(http.StatusOK, "/v1/haha")
		})
		v2 := v1.Group("/v2")
		{
			v2.GET("/Ciallo", func(ctx *gee.Context) {
				ctx.JSON(http.StatusOK, "Ciallo!!!")
			})
			v2.Use(func(ctx *gee.Context) {
				fmt.Println("this is v2 middleware path:", ctx.Path)
			})
		}
		v1.Use(func(ctx *gee.Context) {
			fmt.Println("this is v1 middleware path:", ctx.Path)
		})
	}

	authorized := r.Group("/auth")
	{
		authorized.GET("/logout", func(ctx *gee.Context) {
			ctx.JSON(http.StatusOK, "for logout")
		})
		authorized.GET("/panic", func(ctx *gee.Context) {
			panic("test panic!!")
		})
		authorized.Use(func(ctx *gee.Context) {
			fmt.Println("need authorize... path: ", ctx.Path)

		})
	}

	r.Error(func(panicValue any) {
		log.Println("this is my error handler ", panicValue)
		log.Printf("%s\n", debug.Stack())
	})

	r.Static("/static", "./static")
}