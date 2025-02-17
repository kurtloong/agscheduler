package services

import (
	"fmt"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kurtloong/agscheduler"
)

type cHTTPService struct {
	cn *agscheduler.ClusterNode
}

func (chs *cHTTPService) nodes(c *gin.Context) {
	c.JSON(200, gin.H{"data": chs.cn.NodeMap(), "error": ""})
}

type clusterHTTPService struct {
	Cn *agscheduler.ClusterNode
}

func (s *clusterHTTPService) registerRoutes(r *gin.Engine, shs *cHTTPService) {
	r.GET("/cluster/nodes", shs.nodes)
}

func (s *clusterHTTPService) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	s.registerRoutes(r, &cHTTPService{cn: s.Cn})

	slog.Info(fmt.Sprintf("cluster HTTP Service listening at: %s", s.Cn.EndpointHTTP))

	go func() {
		if err := r.Run(s.Cn.EndpointHTTP); err != nil {
			slog.Error(fmt.Sprintf("Cluster HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}
