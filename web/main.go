package web

import (
	"context"
	"github.com/knaka/biblioseeq/pbgen/v1/v1connect"
	"github.com/knaka/biblioseeq/web/rpc"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net"
	"net/http"
)

func ListenAndServe(ctx context.Context, addr string) error {
	server := &http.Server{Addr: addr, Handler: GetWrappedRouter()}
	server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	return server.ListenAndServe()
}

func GetWrappedRouter() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(v1connect.NewMainServiceHandler(&rpc.MainServiceHandlerImpl{}))
	corsHandler := cors.New(cors.Options{
		Debug: false,
		// “Credentials are cookies, authorization headers, or TLS client certificates.” とのことなので、credential を gRPC の request body にしか乗せないのであれば false でも良いか // Access-Control-Allow-Credentials - HTTP | MDN https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
		AllowCredentials: true,
		// 環境変数で渡せるようにすべきか
		AllowedOrigins: []string{
			//"*",
			// The value of the 'Access-Control-Allow-Origin' header in the response must not be the wildcard '*' when the request's credentials mode is 'include'. The credentials mode of requests initiated by the XMLHttpRequest is controlled by the withCredentials attribute.
			// Local development
			"http://127.0.0.1:3000",
			"http://localhost:3000",
			"http://127.0.0.1:3001",
			"http://localhost:3001",
			"http://127.0.0.1:8000",
			"http://localhost:8000",
			"http://127.0.0.1:8601",
			"http://localhost:8601",
			// Docker Compose environment
			"http://localhost:65151",
			"http://localhost:65154",
			// Cloud environment
		},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{"*"},
	})
	return h2c.NewHandler(corsHandler.Handler(mux), &http2.Server{})
}