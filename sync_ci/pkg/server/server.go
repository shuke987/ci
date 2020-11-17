package server

import (
	"context"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ci/sync_ci/pkg/model"
	"github.com/pingcap/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	cfg *model.Config
}

func NewServer(cfg *model.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run() {
	httpServer := s.setupHttpServer()
	go httpServer.ListenAndServe()

	ch := make(chan os.Signal)
	defer close(ch)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.S().Errorw("http server shutdown", "err", err)
	}
}

func (s *Server) setupDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(s.cfg.Dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	d, err := db.DB()
	if err != nil {
		return nil, err
	}
	d.SetMaxIdleConns(10)
	d.SetMaxOpenConns(100)
	d.SetConnMaxIdleTime(time.Hour)
	return db, nil
}

func (s *Server) setupHttpServer() (httpServer *http.Server) {
	router := gin.Default()
	logger := log.L()
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	db, err := s.setupDB()
	if err != nil {
		panic(fmt.Sprintf("setup db failed: %v", err))
	}
	jenkins, err := gojenkins.CreateJenkins(nil, "https://internal.pingcap.net/idc-jenkins/").Init()
	if err != nil {
		panic(fmt.Sprintf("setup jenkins failed: %v", err))
	}
	syncHandler := &SyncHandler{db, jenkins}
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	apiv1 := router.Group("/api/v1")
	apiv1.POST("/ci/job/sync", syncHandler.syncData)

	addr := fmt.Sprintf("0.0.0.0:%s", s.cfg.Port)
	log.S().Info(fmt.Sprintf("listening on %s", addr))
	httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return httpServer
}

type SyncHandler struct {
	db      *gorm.DB
	jenkins *gojenkins.Jenkins
}

func (h *SyncHandler) syncData(c *gin.Context) {
	//b, err := h.jenkins.GetBuild("tidb_ghpr_unit_test", 58291)
	//if err != nil {
	//	log.S().Error(err)
	//}
	//parameters := b.GetParameters()
	//log.S().Info(parameters)
	c.JSON(http.StatusOK, "resource")
}