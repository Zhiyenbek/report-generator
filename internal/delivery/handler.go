package delivery

import (
	"encoding/json"
	"errors"
	"io"
	"report-generator/internal/models"
	"report-generator/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

type Handler struct {
	service *service.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *service.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	router.Use(cors.New(config))
	router.POST("/report/generate", h.Generate)
	return router
}

func (h *Handler) Generate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error(err)
		c.JSON(400, CreateResponse(-1, nil, models.ErrInvalidInput))
		return
	}
	report := &models.Report{}
	err = json.Unmarshal(body, report)
	if err != nil {
		h.log.Error(err.Error())
		c.JSON(400, CreateResponse(-1, nil, models.ErrInvalidInput))
		return
	}
	report.FinishDate = time.Unix(report.FinishedAt, 0).Format(models.DateFormat)
	report.StartDate = time.Unix(report.StartedAt, 0).Format(models.DateFormat)
	link, err := h.service.Report.Generate(report)
	if err != nil {
		err = errorFilter(err)
		c.JSON(500, CreateResponse(-1, nil, err))
		return
	}
	c.JSON(201, CreateResponse(0, link, nil))
}

func errorFilter(err error) error {
	switch {
	case errors.Is(err, models.ErrHTMLGenerate):
		return models.ErrHTMLGenerate
	case errors.Is(err, models.ErrPDFGenerate):
		return models.ErrPDFGenerate
	case errors.Is(err, models.ErrPDFSave):
		return models.ErrPDFSave
	}
	return models.ErrInternalServer
}

func CreateResponse(status int, data interface{}, err error) gin.H {
	if err != nil {
		return gin.H{
			"response": gin.H{
				"status": status,
				"data":   data,
			},
			"error": gin.H{
				"message": err.Error(),
			},
		}
	}
	return gin.H{
		"response": gin.H{
			"status": status,
			"data":   data,
		},
		"error": nil,
	}
}
