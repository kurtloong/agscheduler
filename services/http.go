package services

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kurtloong/agscheduler"
)

type sHTTPService struct {
	scheduler *agscheduler.Scheduler
}

func (shs *sHTTPService) handleJob(j agscheduler.Job, err error) gin.H {
	if j.Id == "" {
		return gin.H{"data": nil, "error": shs.handleErr(err)}
	} else {
		return gin.H{"data": j, "error": shs.handleErr(err)}
	}
}

func (shs *sHTTPService) handleErr(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}

func (shs *sHTTPService) addJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.AddJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) getJob(c *gin.Context) {
	j, err := shs.scheduler.GetJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) getAllJobs(c *gin.Context) {
	js, err := shs.scheduler.GetAllJobs()
	c.JSON(200, gin.H{"data": js, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) updateJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.UpdateJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) deleteJob(c *gin.Context) {
	err := shs.scheduler.DeleteJob(c.Param("id"))
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) deleteAllJobs(c *gin.Context) {
	shs.scheduler.DeleteAllJobs()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) pauseJob(c *gin.Context) {
	j, err := shs.scheduler.PauseJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) resumeJob(c *gin.Context) {
	j, err := shs.scheduler.ResumeJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) runJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	err = shs.scheduler.RunJob(j)
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) start(c *gin.Context) {
	shs.scheduler.Start()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) stop(c *gin.Context) {
	shs.scheduler.Stop()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

type HTMLRoute struct {
	URLPath  string
	HTMLFile string
}

type SchedulerHTTPService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36370`
	Address      string
	StaticPaths  map[string]string     // New field for static paths
	ExtraRoutes  []func(r *gin.Engine) // New field for additional routes
	HTMLRoutes   []HTMLRoute           // New field for HTML routes
	HTMLGlobPath string                // New field for HTML glob path
}

func (s *SchedulerHTTPService) registerRoutes(r *gin.Engine, shs *sHTTPService) {
	r.POST("/scheduler/job", shs.addJob)
	r.GET("/scheduler/job/:id", shs.getJob)
	r.GET("/scheduler/jobs", shs.getAllJobs)
	r.PUT("/scheduler/job", shs.updateJob)
	r.DELETE("/scheduler/job/:id", shs.deleteJob)
	r.DELETE("/scheduler/jobs", shs.deleteAllJobs)
	r.POST("/scheduler/job/:id/pause", shs.pauseJob)
	r.POST("/scheduler/job/:id/resume", shs.resumeJob)
	r.POST("/scheduler/job/run", shs.runJob)
	r.POST("/scheduler/start", shs.start)
	r.POST("/scheduler/stop", shs.stop)
}

// New method to add static paths
func (s *SchedulerHTTPService) AddStaticPath(urlPath, localPath string) {
	if s.StaticPaths == nil {
		s.StaticPaths = make(map[string]string)
	}
	s.StaticPaths[urlPath] = localPath
}

// New method to add extra routes
func (s *SchedulerHTTPService) AddRoute(routeFunc func(r *gin.Engine)) {
	s.ExtraRoutes = append(s.ExtraRoutes, routeFunc)
}

// New method to add HTML resource routes
func (s *SchedulerHTTPService) AddHTMLRoute(urlPath, htmlFile string) {
	s.HTMLRoutes = append(s.HTMLRoutes, HTMLRoute{URLPath: urlPath, HTMLFile: htmlFile})
}

func (s *SchedulerHTTPService) AddHTMLGlobPath(globPath string) {
	s.HTMLGlobPath = globPath
}

func (s *SchedulerHTTPService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36370"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	s.registerRoutes(r, &sHTTPService{scheduler: s.Scheduler})

	slog.Info(fmt.Sprintf("Scheduler HTTP Service listening at: %s", s.Address))
	if condition := s.HTMLGlobPath != ""; condition {
		r.LoadHTMLGlob(s.HTMLGlobPath)
	}
	// Setting up static paths
	for urlPath, localPath := range s.StaticPaths {
		r.Static(urlPath, localPath)
	}

	// Adding extra routes
	for _, routeFunc := range s.ExtraRoutes {
		routeFunc(r)
	}

	// Serve HTML files
	for _, route := range s.HTMLRoutes {
		s.serveHTML(r, route.URLPath, route.HTMLFile)
	}

	go func() {
		if err := r.Run(s.Address); err != nil {
			slog.Error(fmt.Sprintf("Scheduler HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}

// Helper method to serve an HTML file
func (s *SchedulerHTTPService) serveHTML(r *gin.Engine, urlPath, htmlFile string) {
	r.GET(urlPath, func(c *gin.Context) {
		c.HTML(http.StatusOK, htmlFile, nil)
	})
}
