package repository

import (
	"database/sql"
	"report-generator/config"
	"report-generator/internal/models"

	"go.uber.org/zap"
)

type Repository struct {
	Report
}

type Report interface {
	//Generate takes data from report model and generates
	//two pdf files from it cover and report pages then
	//it merges them with another call to gottenberg
	//returns pdf file as slice of bytes in case of error
	//occurence returns nil and error
	Generate(*models.Report) ([]byte, error)
}

func NewRepository(log *zap.SugaredLogger, cfg *config.GeneratorConfig, db *sql.DB) *Repository {
	return &Repository{
		Report: NewReportRepo(log, cfg),
	}
}
