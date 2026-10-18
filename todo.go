// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"todo/internal/config"
	"todo/internal/handler"
	"todo/internal/middleware"
	"todo/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/todo-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithFileServer("/", http.Dir("web")))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	server.Use(middleware.Auth(ctx))
	handler.RegisterHandlers(server, ctx)

	httpx.SetErrorHandler(handler.HandleError)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
