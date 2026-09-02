package main

// 状态常量: 无备投, 记录即已投递
const (
	StatusApplied         = "APPLIED"         // 已投递
	StatusInterviewing    = "INTERVIEWING"    // 面试中
	StatusOffered         = "OFFERED"         // 已获 Offer
	StatusResumeRejected  = "RESUME_REJECTED" // 简历挂
	StatusInterviewFailed = "INTERVIEW_FAILED"
	StatusDeclined        = "DECLINED" // 我拒绝对方
)

// Application 一条投递记录
type Application struct {
	ID            int64  `json:"id"`
	Company       string `json:"company"`
	Position      string `json:"position"`
	Status        string `json:"status"` // APPLIED/INTERVIEWING/OFFERED/RESUME_REJECTED/INTERVIEW_FAILED/DECLINED
	ResumeVersion string `json:"resume_version"`
	JDText        string `json:"jd_text"`
	AppliedAt     string `json:"applied_at"` // YYYY-MM-DD
	Location      string `json:"location"`
	SalaryRange   string `json:"salary_range"`
	ContactInfo   string `json:"contact_info"`
	Notes         string `json:"notes"`
	// 首次进入面试流程时置 1, 只增不改 (统计用)
	ReachedInterview int64 `json:"reached_interview"`
	// 面试挂时的轮次 (0 = 无, 由 FAILED 面试自动计算)
	FailedRound    int64        `json:"failed_round"`
	CreatedAt      string       `json:"created_at"`
	UpdatedAt      string       `json:"updated_at"`
	Interviews     []*Interview `json:"interviews,omitempty"`
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
	Status        string `json:"status"` // 仅创建时生效
}

// Interview 面试记录 (1-N 关联 Application)
type Interview struct {
	ID                int64  `json:"id"`
	ApplicationID     int64  `json:"application_id"`
	RoundName         string `json:"round_name"`
	ScheduledAt       string `json:"scheduled_at"` // YYYY-MM-DDTHH:mm
	QuestionsAndNotes string `json:"questions_and_notes"`
	Outcome           string `json:"outcome"` // PENDING/PASSED/FAILED
	CreatedAt         string `json:"created_at"`
	QAItems           []*QAItem `json:"qa_items,omitempty"` // 逐条问题记录
}

// InterviewInput 创建/更新面试记录
type InterviewInput struct {
	ApplicationID     int64  `json:"application_id"`
	RoundName         string `json:"round_name"`
	ScheduledAt       string `json:"scheduled_at"`
	QuestionsAndNotes string `json:"questions_and_notes"`
	Outcome           string `json:"outcome"`
}

// QAItem 面试中的逐条问题记录
type QAItem struct {
	ID          int64  `json:"id"`
	InterviewID int64  `json:"interview_id"`
	Question    string `json:"question"`
	Answer      string `json:"answer"`
	Reflection  string `json:"reflection"` // 复盘改进
	SortOrder   int64  `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
}

// QAItemInput 创建/更新问题条目
type QAItemInput struct {
	InterviewID int64  `json:"interview_id"`
	Question    string `json:"question"`
	Answer      string `json:"answer"`
	Reflection  string `json:"reflection"`
	SortOrder   int64  `json:"sort_order"` // <=0 = 追加到末尾; >0 = 插入到该序号位置
}

// Stats 统计信息
type Stats struct {
	TotalApplications int64   `json:"total_applications"`
	Interviewing      int64   `json:"interviewing"`      // 当前处于面试中
	ReachedInterview  int64   `json:"reached_interview"` // 进入过面试的申请数
	Offered           int64   `json:"offered"`
	ResumeRejected    int64   `json:"resume_rejected"`  // 简历挂
	InterviewFailed   int64   `json:"interview_failed"` // 面试挂
	Declined          int64   `json:"declined"`         // 我拒绝对方
	InterviewRate     float64 `json:"interview_rate"`   // ReachedInterview / Total
	OfferRate         float64 `json:"offer_rate"`       // Offered / Total
	ByResumeVersion   []ResumeVersionStat  `json:"by_resume_version"`
}

// ResumeVersionStat 按简历版本统计
type ResumeVersionStat struct {
	ResumeVersion    string  `json:"resume_version"`
	TotalApplied     int64   `json:"total_applied"`
	ReachedInterview int64   `json:"reached_interview"`
	InterviewRate    float64 `json:"interview_rate"`
}
