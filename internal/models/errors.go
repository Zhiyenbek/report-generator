package models

import "errors"

var (
	ErrInvalidInput   = errors.New("INVALID_INPUT")
	ErrHTMLGenerate   = errors.New("HTML_GENERATE_ERROR")
	ErrPDFGenerate    = errors.New("PDF_GENERATE_ERROR")
	ErrPDFSave        = errors.New("PDF_SAVE_ERROR")
	ErrInternalServer = errors.New("INTERNAL_SERVER_ERROR")
)
