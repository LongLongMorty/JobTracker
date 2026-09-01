package main

// Application 一条投递记录
type Application struct {
	ID               int64       `json:"id"`
	Company          string      `json:"company"`
	Position         string      `json:"position"`
	Status           string      `json:"status"`             // WISHLIST/APPLIED/INTERVIEWING/OFFERED/REJECTED/ARCHIVED
	ResumeVersion    string      `json:"resume_version"`
	JDText           string      `json:"jd_text"`
	AppliedAt        string      `json:"applied_at"` // YYYY-MM-DD
	Location         string      `json:"location"`
	SalaryRange      string      `json:"salary_range"`
	ContactInfo      string      `json:"contact_info"`
	Notes            string      `json:"notes"`
	ReachedInterview int64       `json:"reached_interview"` // 首次进入 INTERVIEWING 时置 1, 只增不改
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
	Interviews       []*Interview `json:"interviews,omitempty"`
}

// ApplicationInput 创建/更新投递记录时前端传入的字段
type ApplicationInput struct {
	Company       string `json:"company"`
	Position      string `json:"position"`
	ResumeVersion string `json:"resume_version"`
	JDText        string `json:"jd_text"`
	AppliedAt     string `json:"applied_at"`
	Location      string `json:"location"`
	SalaryRange   string `json:"salary_range"`
	ContactInfo   string `json:"contact_info"`
	Notes         string `json:"notes"`
}

// Interview 面试记录 (1-N 关联 Application)
type Interview struct {
	ID               int64  `json:"id"`
	ApplicationID    int64  `json:"application_id"`
	RoundName        string `json:"round_name"`
	ScheduledAt      string `json:"scheduled_at"` // YYYY-MM-DDTHH:mm
	QuestionsAndNotes string `json:"questions_and_notes"`
	Outcome          string `json:"outcome"` // PENDING/PASSED/FAILED
	CreatedAt        string `json:"created_at"`
}

// InterviewInput 创建/更新面试记录
type InterviewInput struct {
	ApplicationID     int64  `json:"application_id"`
	RoundName         string `json:"round_name"`
	ScheduledAt       string `json:"scheduled_at"`
	QuestionsAndNotes string `json:"questions_and_notes"`
	Outcome           string `json:"outcome"`
}

// Stats 统计信息
type Stats struct {
	TotalApplications int64   `json:"total_applications"`
	TotalApplied      int64   `json:"total_applied"`       // 已投递数 (APPLIED 及之后的非 wishlist/archived)
	Interviewing      int64   `json:"interviewing"`        // 当前处于 INTERVIEWING
	ReachedInterview  int64   `json:"reached_interview"`   // 进入过面试的申请数
	Offered           int64   `json:"offered"`
	Rejected          int64   `json:"rejected"`
	Archived          int64   `json:"archived"`
	InterviewRate     float64 `json:"interview_rate"` // ReachedInterview / TotalApplied
	OfferRate         float64 `json:"offer_rate"`     // Offered / TotalApplied
	ByResumeVersion   []ResumeVersionStat `json:"by_resume_version"`
}

// ResumeVersionStat 按简历版本统计
type ResumeVersionStat struct {
	ResumeVersion    string  `json:"resume_version"`
	TotalApplied     int64   `json:"total_applied"`
	ReachedInterview int64   `json:"reached_interview"`
	InterviewRate    float64 `json:"interview_rate"`
}
