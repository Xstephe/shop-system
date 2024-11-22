package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"

	"mxshop-api/order-web/global"
)

// 链路跟踪
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},

			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:           true,
				LocalAgentHostPort: fmt.Sprintf("%s:%d", global.ServerConfig.JaegerInfo.Host, global.ServerConfig.JaegerInfo.Port),
			},
			ServiceName: global.ServerConfig.JaegerInfo.Name,
		}
		tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(jaegerlog.StdLogger))
		if err != nil {
			panic(err)
		}
		defer closer.Close()
		startSpan := tracer.StartSpan(c.Request.URL.Path)
		defer startSpan.Finish()
		c.Set("tracer", tracer)
		c.Set("parentSpan", startSpan)
		c.Next()
	}
}
