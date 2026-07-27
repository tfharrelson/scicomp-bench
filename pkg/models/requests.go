package models

type SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SubmitJobRequest struct {
	Type      JobType `json:"type"`
	JobName   string  `json:"job_name"`
	InputFile string  `json:"input_file"`
}
