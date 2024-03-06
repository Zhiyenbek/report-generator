package service

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"report-generator/config"
	"report-generator/internal/models"
	"report-generator/internal/repository"

	"go.uber.org/zap"
)

type ReportService struct {
	repo repository.Report
	log  *zap.SugaredLogger
	cfg  *config.Configuration
}

func NewReportService(repo repository.Report, log *zap.SugaredLogger, cfg *config.Configuration) *ReportService {
	return &ReportService{repo, log, cfg}
}

func (r *ReportService) Generate(report *models.Report) (string, error) {
	pdf, err := r.repo.Generate(report)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s%s", report.Name, ".pdf")
	basePath := r.cfg.App.BasePath
	dst, err := os.Create(filepath.Join(basePath, "server/", filename))
	if err != nil {
		r.log.Error(err)
		return "", models.ErrPDFSave
	}
	err = dst.Chmod(fs.FileMode(0777))
	if err != nil {
		r.log.Error(err)
		return "", models.ErrInternalServer
	}
	defer dst.Close()
	if _, err = io.Copy(dst, bytes.NewBuffer(pdf)); err != nil {
		r.log.Error(err)
		return "", models.ErrPDFSave
	}
	return filename, nil
}
