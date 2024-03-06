package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"report-generator/config"
	"report-generator/internal/models"

	"go.uber.org/zap"
)

type reportRepo struct {
	log *zap.SugaredLogger
	cfg *config.GeneratorConfig
}

func NewReportRepo(log *zap.SugaredLogger, cfg *config.GeneratorConfig) Report {
	return &reportRepo{log, cfg}
}

func (r *reportRepo) Generate(report *models.Report) ([]byte, error) {
	reports, err := r.generateReportsPage(report)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// prepareGenerate prepares reports cover pdf files to be merged creates
func (r *reportRepo) prepareGenerate(reports, cover []byte) ([]byte, error) {
	mpw, buffer, err := r.prepareMerge(cover, reports)
	if err != nil {
		return nil, err
	}
	return r.request(mpw, r.cfg.MergeURL, buffer)
}

// prepareMerge prepares cover and reports files to be merged
// creates buffer and multipart writer from it, takes two pdf files cover and reports
// adds both to buffer and returns writer, buffer and error
func (r *reportRepo) prepareMerge(cover []byte, reports []byte) (*multipart.Writer, *bytes.Buffer, error) {
	bodyBuffer := &bytes.Buffer{}
	mpw := multipart.NewWriter(bodyBuffer)
	defer r.close(mpw)
	err := addFormFile(mpw, bytes.NewBuffer(cover), models.CoverPDFName)
	if err != nil {
		r.log.Errorf("adding cover pdf error: %s", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	err = addFormFile(mpw, bytes.NewBuffer(reports), models.ReportsPDFName)
	if err != nil {
		r.log.Errorf("adding reports pdf error: %s", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	return mpw, bodyBuffer, nil
}

// generateCover returns cover page pdf as slice of bytes
func (r *reportRepo) generateCover(report *models.Report) ([]byte, error) {
	mpw, bodyBuffer, err := r.prepareCoverConvert(report)
	if err != nil {
		return nil, err
	}
	return r.request(mpw, r.cfg.ConvertURL, bodyBuffer)
}

// generateReportsPage returns reports pdf as slice of bytes
func (r *reportRepo) generateReportsPage(report *models.Report) ([]byte, error) {
	mpw, buffer, err := r.prepareReportsConvert(report)
	if err != nil {
		return nil, err
	}
	return r.request(mpw, r.cfg.ConvertURL, buffer)
}

// margins holds arguments for query
var margins = []string{"marginTop", "marginBottom", "marginLeft", "marginRight"}

// setMargins sets all margins to 0 value for cover page
func (r *reportRepo) setMargins(mpw *multipart.Writer) error {
	for _, margin := range margins {
		err := r.addFormField(mpw, margin, "0")
		if err != nil {
			return err
		}
	}
	return nil
}

func addFormFile(mpw *multipart.Writer, file io.Reader, name string) error {
	formFile, err := mpw.CreateFormFile("files", name)
	if err != nil {
		return fmt.Errorf("error during creating form file header: %s", err.Error())
	}
	_, err = io.Copy(formFile, file)
	if err != nil {
		return fmt.Errorf("error during copying data from buffer to form file header: %s", err.Error())
	}
	return nil
}

func (r *reportRepo) addFormField(mpw *multipart.Writer, name string, value string) error {
	header, err := mpw.CreateFormField(name)
	if err != nil {
		return fmt.Errorf("error during creating form field: %s", err.Error())
	}
	n, err := header.Write([]byte(value))
	if err != nil {
		return fmt.Errorf("error during setting value to a form field: %s", err.Error())
	}
	if n != len(value) {
		return fmt.Errorf("couldn't write all form field bytes, wrote %d out of %d", n, len(value))
	}
	return nil
}

// request function takes multipart writer, url to which make call, and byte buffer
// returns response body, in case of error occurence returns nil and error
func (r *reportRepo) request(mpw *multipart.Writer, url string, body *bytes.Buffer) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		r.log.Errorf("request initiliazation error: %s", err.Error())
		return nil, models.ErrInternalServer
	}
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	fmt.Println(body.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.log.Errorf("request error: %s", err.Error())
		return nil, models.ErrPDFGenerate
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		r.log.Error("error during reading response body: %s", err.Error())
		return nil, models.ErrInternalServer
	}
	return responseBody, nil
}

func (r *reportRepo) close(mpw *multipart.Writer) {
	err := mpw.Close()
	if err != nil {
		r.log.Errorf("error during closing multipart writer: %s", err.Error())
	}
}

// prepareCoverConvert prepares cover page for making request to gottenberg
// creates bytes buffer and multipart writer from it. Executes cover template using report data
// puts that template into buffer and then returns writer, buffer and error
func (r *reportRepo) prepareCoverConvert(report *models.Report) (*multipart.Writer, *bytes.Buffer, error) {
	var document bytes.Buffer
	err := r.cfg.TemplateFile.Execute(&document, report)
	if err != nil {
		r.log.Errorf("error during cover template exec: %s", err.Error())
		return nil, nil, models.ErrHTMLGenerate
	}
	bodyBuffer := &bytes.Buffer{}
	mpw := multipart.NewWriter(bodyBuffer)
	defer r.close(mpw)
	err = addFormFile(mpw, &document, models.IndexHTML)
	if err != nil {
		r.log.Error("error during adding template form file: %s", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	err = r.addFormField(mpw, "nativePageRanges", "1")
	if err != nil {
		r.log.Error("error during setting page range field to 1", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	err = r.setMargins(mpw)
	if err != nil {
		r.log.Errorf("error during setting margins to 0", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	return mpw, bodyBuffer, nil
}

// prepareReportsConvert prepares pages with reports for gottenberg. creates bytes buffer and multipart from it.
// Executes reports template using report data then puts its output into buffer and then returns writer, buffer and error
func (r *reportRepo) prepareReportsConvert(report *models.Report) (*multipart.Writer, *bytes.Buffer, error) {
	var document bytes.Buffer
	err := r.cfg.TemplateFile.Execute(&document, report)
	if err != nil {
		r.log.Errorf("error during report template exec: %s", err.Error())
		return nil, nil, models.ErrHTMLGenerate
	}
	bodyBuffer := &bytes.Buffer{}
	mpw := multipart.NewWriter(bodyBuffer)
	defer r.close(mpw)
	fmt.Println(document.String())
	err = addFormFile(mpw, &document, models.IndexHTML)
	if err != nil {
		r.log.Errorf("add report file error: %s", err.Error())
		return nil, nil, models.ErrInternalServer
	}

	return mpw, bodyBuffer, nil
}
