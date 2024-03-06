package service

import (
	"report-generator/config"
	"report-generator/internal/models"
	"report-generator/internal/repository"

	"go.uber.org/zap"
)

type Service struct {
	Report
	Baseline
}

type Report interface {
	Generate(*models.Report) (string, error)
}

type Baseline interface {
	CheckBaseline(name string) error
}

func NewService(repo *repository.Repository, log *zap.SugaredLogger, cfg *config.Configuration) *Service {
	return &Service{
		Report: NewReportService(repo.Report, log, cfg),
	}
}
