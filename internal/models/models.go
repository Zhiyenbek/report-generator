package models

type Result struct {
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	Evaluation string `json:"evaluation"`
	Score      string `json:"score"`
}

type Report struct {
	Name       string   `json:"name"`
	StartDate  string   `json:"-"`
	FinishDate string   `json:"-"`
	StartedAt  int64    `json:"started_at"`
	FinishedAt int64    `json:"finished_at"`
	ID         string   `json:"report_id"`
	Results    []Result `json:"results"`
}
