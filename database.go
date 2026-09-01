package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// dbPath 数据库文件路径: <config dir>/jobtracker/app.db
func dbPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户配置目录: %w", err)
	}
	dir = filepath.Join(dir, "jobtracker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("无法创建配置目录 %s: %w", dir, err)
	}
	return filepath.Join(dir, "app.db"), nil
}

// initDB 打开数据库并执行迁移
func initDB() error {
	path, err := dbPath()
	if err != nil {
		return err
	}
	db, err = sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	return migrate()
}

func migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company TEXT NOT NULL,
    position TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'APPLIED',
    reached_interview INTEGER NOT NULL DEFAULT 0,
    resume_version TEXT DEFAULT '',
    jd_text TEXT DEFAULT '',
    applied_at TEXT DEFAULT '',
    location TEXT DEFAULT '',
    salary_range TEXT DEFAULT '',
    contact_info TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS trg_applications_updated_at
AFTER UPDATE ON applications
FOR EACH ROW
BEGIN
    UPDATE applications SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS interviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER NOT NULL,
    round_name TEXT NOT NULL,
    scheduled_at TEXT NOT NULL,
    questions_and_notes TEXT DEFAULT '',
    outcome TEXT DEFAULT 'PENDING',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_interviews_application_id ON interviews(application_id);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// 新增列 (旧库升级): failed_round 记录面试挂的轮次
	has, err := columnExists("applications", "failed_round")
	if err != nil {
		return err
	}
	if !has {
		if _, err := db.Exec(`ALTER TABLE applications ADD COLUMN failed_round INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}

	// 旧版本状态迁移 (WISHLIST/REJECTED/ARCHIVED 已废弃)
	legacy := []string{
		`UPDATE applications SET status = 'APPLIED' WHERE status = 'WISHLIST'`,
		`UPDATE applications SET status = 'INTERVIEW_FAILED', reached_interview = 1 WHERE status = 'REJECTED' AND reached_interview = 1`,
		`UPDATE applications SET status = 'RESUME_REJECTED' WHERE status = 'REJECTED'`,
		`UPDATE applications SET status = 'DECLINED' WHERE status = 'ARCHIVED'`,
	}
	for _, stmt := range legacy {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// columnExists 检查某列是否存在
func columnExists(table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, typ, dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name.String == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// ---- Applications ----

const applicationCols = `id, company, position, status, reached_interview, resume_version, jd_text,
	applied_at, location, salary_range, contact_info, notes, created_at, updated_at, failed_round`

func scanApplication(row interface{ Scan(...any) error }) (*Application, error) {
	var a Application
	err := row.Scan(&a.ID, &a.Company, &a.Position, &a.Status, &a.ReachedInterview,
		&a.ResumeVersion, &a.JDText, &a.AppliedAt, &a.Location, &a.SalaryRange,
		&a.ContactInfo, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &a.FailedRound)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func getApplication(id int64) (*Application, error) {
	row := db.QueryRow("SELECT "+applicationCols+" FROM applications WHERE id = ?", id)
	a, err := scanApplication(row)
	if err != nil {
		return nil, err
	}
	ivs, err := listInterviews(id)
	if err != nil {
		return nil, err
	}
	a.Interviews = ivs
	return a, nil
}

func listApplications() ([]*Application, error) {
	rows, err := db.Query("SELECT " + applicationCols + " FROM applications ORDER BY updated_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	apps := []*Application{}
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func insertApplication(in ApplicationInput) (*Application, error) {
	appliedAt := in.AppliedAt
	if appliedAt == "" {
		appliedAt = time.Now().Format("2006-01-02")
	}
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status == "" {
		status = StatusApplied
	}
	res, err := db.Exec(`INSERT INTO applications
		(company, position, status, resume_version, jd_text, applied_at, location, salary_range, contact_info, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Company, in.Position, status, in.ResumeVersion, in.JDText, appliedAt,
		in.Location, in.SalaryRange, in.ContactInfo, in.Notes)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return getApplication(id)
}

func updateApplication(id int64, in ApplicationInput) (*Application, error) {
	_, err := db.Exec(`UPDATE applications SET
		company = ?, position = ?, resume_version = ?, jd_text = ?, applied_at = ?,
		location = ?, salary_range = ?, contact_info = ?, notes = ?
		WHERE id = ?`,
		in.Company, in.Position, in.ResumeVersion, in.JDText, in.AppliedAt,
		in.Location, in.SalaryRange, in.ContactInfo, in.Notes, id)
	if err != nil {
		return nil, err
	}
	return getApplication(id)
}

// updateApplicationStatus 更新状态, 并处理:
//   - 进入 APPLIED 且 applied_at 为空时自动填充当天日期
//   - 进入 INTERVIEWING / INTERVIEW_FAILED 时 reached_interview 置 1 (只增不改)
//   - 离开 INTERVIEW_FAILED 时重算/清除 failed_round
func updateApplicationStatus(id int64, status string) (*Application, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	a, err := getApplication(id)
	if err != nil {
		return nil, err
	}
	appliedAt := a.AppliedAt
	if status == StatusApplied && appliedAt == "" {
		appliedAt = time.Now().Format("2006-01-02")
	}
	reached := a.ReachedInterview
	if (status == StatusInterviewing || status == StatusInterviewFailed) && reached == 0 {
		reached = 1
	}
	round := a.FailedRound
	if status != StatusInterviewFailed {
		// 拖离面试挂: 根据现有 FAILED 面试重算 (通常为 0)
		r, err := computeFailedRound(id)
		if err != nil {
			return nil, err
		}
		round = r
	}
	_, err = db.Exec(`UPDATE applications SET status = ?, applied_at = ?, reached_interview = ?, failed_round = ? WHERE id = ?`,
		status, appliedAt, reached, round, id)
	if err != nil {
		return nil, err
	}
	return getApplication(id)
}

func deleteApplication(id int64) error {
	_, err := db.Exec("DELETE FROM applications WHERE id = ?", id)
	return err
}

// ---- Interviews ----

const interviewCols = `id, application_id, round_name, scheduled_at, questions_and_notes, outcome, created_at`

func scanInterview(row interface{ Scan(...any) error }) (*Interview, error) {
	var iv Interview
	err := row.Scan(&iv.ID, &iv.ApplicationID, &iv.RoundName, &iv.ScheduledAt,
		&iv.QuestionsAndNotes, &iv.Outcome, &iv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &iv, nil
}

func listInterviews(applicationID int64) ([]*Interview, error) {
	rows, err := db.Query("SELECT "+interviewCols+" FROM interviews WHERE application_id = ? ORDER BY scheduled_at, id", applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ivs := []*Interview{}
	for rows.Next() {
		iv, err := scanInterview(rows)
		if err != nil {
			return nil, err
		}
		ivs = append(ivs, iv)
	}
	return ivs, rows.Err()
}

func insertInterview(in InterviewInput) (*Interview, error) {
	res, err := db.Exec(`INSERT INTO interviews (application_id, round_name, scheduled_at, questions_and_notes, outcome)
		VALUES (?, ?, ?, ?, ?)`,
		in.ApplicationID, in.RoundName, in.ScheduledAt, in.QuestionsAndNotes, defaultOutcome(in.Outcome))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	var iv Interview
	err = db.QueryRow("SELECT "+interviewCols+" FROM interviews WHERE id = ?", id).Scan(
		&iv.ID, &iv.ApplicationID, &iv.RoundName, &iv.ScheduledAt, &iv.QuestionsAndNotes, &iv.Outcome, &iv.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := syncStatusOnOutcome(iv.ApplicationID, "", iv.Outcome); err != nil {
		return nil, err
	}
	return &iv, nil
}

func updateInterview(id int64, in InterviewInput) (*Interview, error) {
	var oldOutcome string
	err := db.QueryRow("SELECT outcome FROM interviews WHERE id = ?", id).Scan(&oldOutcome)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`UPDATE interviews SET round_name = ?, scheduled_at = ?, questions_and_notes = ?, outcome = ? WHERE id = ?`,
		in.RoundName, in.ScheduledAt, in.QuestionsAndNotes, defaultOutcome(in.Outcome), id)
	if err != nil {
		return nil, err
	}
	var iv Interview
	err = db.QueryRow("SELECT "+interviewCols+" FROM interviews WHERE id = ?", id).Scan(
		&iv.ID, &iv.ApplicationID, &iv.RoundName, &iv.ScheduledAt, &iv.QuestionsAndNotes, &iv.Outcome, &iv.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := syncStatusOnOutcome(iv.ApplicationID, oldOutcome, iv.Outcome); err != nil {
		return nil, err
	}
	return &iv, nil
}

// syncStatusOnOutcome 面试结果与申请状态自动同步:
//   - 某轮面试标记为 FAILED → 卡片自动移到面试挂, 并记录挂的轮次
//   - 从 FAILED 改回 PASSED/PENDING 且当前是面试挂 → 回到面试中
func syncStatusOnOutcome(applicationID int64, oldOutcome, newOutcome string) error {
	newOutcome = strings.ToUpper(newOutcome)
	oldOutcome = strings.ToUpper(oldOutcome)
	if newOutcome == oldOutcome {
		return nil
	}
	if newOutcome == "FAILED" {
		if _, err := updateApplicationStatus(applicationID, StatusInterviewFailed); err != nil {
			return err
		}
		round, err := computeFailedRound(applicationID)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE applications SET failed_round = ? WHERE id = ?", round, applicationID)
		return err
	}
	if oldOutcome == "FAILED" {
		a, err := getApplication(applicationID)
		if err != nil {
			return err
		}
		if a.Status == StatusInterviewFailed {
			_, err := updateApplicationStatus(applicationID, StatusInterviewing)
			return err
		}
	}
	return nil
}

// computeFailedRound 计算该申请挂在第几轮:
// 按面试时间 (scheduled_at, 时间相同再按 id) 排序, FAILED 那条面试的 1-based 序号; 无 FAILED 面试则返回 0
func computeFailedRound(applicationID int64) (int64, error) {
	rows, err := db.Query(`SELECT outcome FROM interviews WHERE application_id = ? ORDER BY scheduled_at, id`, applicationID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var round int64
	for rows.Next() {
		var outcome string
		if err := rows.Scan(&outcome); err != nil {
			return 0, err
		}
		round++
		if strings.ToUpper(outcome) == "FAILED" {
			return round, nil
		}
	}
	return 0, rows.Err()
}

func deleteInterview(id int64) error {
	_, err := db.Exec("DELETE FROM interviews WHERE id = ?", id)
	return err
}

func defaultOutcome(o string) string {
	if o == "" {
		return "PENDING"
	}
	return strings.ToUpper(o)
}

// ---- Stats ----

func getStats() (*Stats, error) {
	var s Stats
	s.ByResumeVersion = []ResumeVersionStat{}
	err := db.QueryRow(`SELECT
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN status = 'INTERVIEWING' THEN 1 ELSE 0 END), 0) AS interviewing,
		COALESCE(SUM(reached_interview), 0) AS reached,
		COALESCE(SUM(CASE WHEN status = 'OFFERED' THEN 1 ELSE 0 END), 0) AS offered,
		COALESCE(SUM(CASE WHEN status = 'RESUME_REJECTED' THEN 1 ELSE 0 END), 0) AS resume_rejected,
		COALESCE(SUM(CASE WHEN status = 'INTERVIEW_FAILED' THEN 1 ELSE 0 END), 0) AS interview_failed,
		COALESCE(SUM(CASE WHEN status = 'DECLINED' THEN 1 ELSE 0 END), 0) AS declined
		FROM applications`).Scan(
		&s.TotalApplications, &s.Interviewing, &s.ReachedInterview, &s.Offered,
		&s.ResumeRejected, &s.InterviewFailed, &s.Declined)
	if err != nil {
		return nil, err
	}
	if s.TotalApplications > 0 {
		s.InterviewRate = float64(s.ReachedInterview) / float64(s.TotalApplications)
		s.OfferRate = float64(s.Offered) / float64(s.TotalApplications)
	}

	rows, err := db.Query(`SELECT resume_version, COUNT(*) AS applied, SUM(reached_interview) AS reached
		FROM applications WHERE resume_version <> ''
		GROUP BY resume_version ORDER BY reached DESC, applied DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v ResumeVersionStat
		if err := rows.Scan(&v.ResumeVersion, &v.TotalApplied, &v.ReachedInterview); err != nil {
			return nil, err
		}
		if v.TotalApplied > 0 {
			v.InterviewRate = float64(v.ReachedInterview) / float64(v.TotalApplied)
		}
		s.ByResumeVersion = append(s.ByResumeVersion, v)
	}
	return &s, rows.Err()
}
