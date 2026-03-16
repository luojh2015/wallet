package server

import (
	"context"
	"embed"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/luojh/wallet/api/grpc/v1"
	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed all:swagger-ui
var swaggerUI embed.FS

// NewHTTPServer 创建 HTTP 服务器（集成 gRPC-Gateway）
func NewHTTPServer(
	cfg *config.Config,
	middleware *middleware.GinMiddleware,
) *http.Server {
	router := gin.New()
	middleware.Apply(router)

	gatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
		runtime.WithErrorHandler(customErrorHandler),
		// 配置传入 Header 匹配器，将 HTTP Header 转换为 gRPC metadata
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	// 使用 bufconn 实现进程内连接
	client, err := grpc.NewClient(
		"passthrough:///localhost:"+strconv.Itoa(cfg.App.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return net.Dial("tcp", ":"+strconv.Itoa(cfg.App.GRPCPort))
		}),
	)
	if err != nil {
		panic(err)
	}

	if err := v1.RegisterWalletServiceHandler(context.Background(), gatewayMux, client); err != nil {
		panic(err)
	}

	// 注册路由
	registerGatewayRoutes(router, gatewayMux)

	return &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.App.HTTPPort),
		Handler: router,
	}
}

// headerMatcher 配置 HTTP Header 到 gRPC metadata 的映射
func headerMatcher(key string) (string, bool) {
	// 转换 Authorization header
	switch strings.ToLower(key) {
	case "authorization":
		return "authorization", true
	case "x-session-token":
		return "x-session-token", true
	default:
		// 其他 header 使用默认处理
		return runtime.DefaultHeaderMatcher(key)
	}
}

// registerGatewayRoutes 注册 gRPC-Gateway 路由到 Gin
func registerGatewayRoutes(router *gin.Engine, gatewayMux http.Handler) {
	// 定义需要代理到 gRPC-Gateway 的路由
	// routes := []struct {
	// 	method string
	// 	path   string
	// }{
	// 	{"POST", "/v1/wallets"},
	// 	{"GET", "/v1/wallets/:wallet_id"},
	// 	{"POST", "/v1/wallets/transfer"},
	// 	{"GET", "/v1/transactions/:wallet_id"},
	// 	{"POST", "/v1/auth/login"},
	// 	{"POST", "/v1/auth/logout"},
	// }

	// for _, route := range routes {
	// 	switch route.method {
	// 	case "GET":
	// 		router.GET(route.path, wrapGatewayHandler(gatewayMux))
	// 	case "POST":
	// 		router.POST(route.path, wrapGatewayHandler(gatewayMux))
	// 	case "PUT":
	// 		router.PUT(route.path, wrapGatewayHandler(gatewayMux))
	// 	case "DELETE":
	// 		router.DELETE(route.path, wrapGatewayHandler(gatewayMux))
	// 	}
	// }

	router.GET("/healthy", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 注册 Swagger UI 路由
	registerSwaggerRoutes(router)

	router.Any("/v1/*path", wrapGatewayHandler(gatewayMux))
}

// registerSwaggerRoutes 注册 Swagger UI 和 API 文档路由
func registerSwaggerRoutes(router *gin.Engine) {
	// Swagger UI 静态文件 - 使用 StripPrefix 去除 /swagger-ui/ 前缀
	router.GET("/swagger-ui/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if filepath == "" || filepath == "/" {
			filepath = "/index.html"
		}
		// 从 embed fs 中读取文件，路径需要包含 swagger-ui 前缀
		data, err := swaggerUI.ReadFile("swagger-ui" + filepath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		// 根据文件扩展名设置 Content-Type
		switch {
		case strings.HasSuffix(filepath, ".html"):
			c.Header("Content-Type", "text/html")
		case strings.HasSuffix(filepath, ".js"):
			c.Header("Content-Type", "application/javascript")
		case strings.HasSuffix(filepath, ".css"):
			c.Header("Content-Type", "text/css")
		case strings.HasSuffix(filepath, ".json"):
			c.Header("Content-Type", "application/json")
		}
		c.Data(http.StatusOK, c.ContentType(), data)
	})

	// Swagger UI 入口页面 - 重定向到 index.html
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger-ui/index.html")
	})

	// Swagger JSON 文件
	router.GET("/swagger.json", func(c *gin.Context) {
		data, err := swaggerUI.ReadFile("swagger-ui/wallet.swagger.json")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	})
}

// wrapGatewayHandler 包装 gRPC-Gateway 处理器
func wrapGatewayHandler(gatewayMux http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		gatewayMux.ServeHTTP(c.Writer, c.Request)
	}
}

// customErrorHandler 自定义错误处理
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// 使用标准 JSON 格式返回错误
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
}
