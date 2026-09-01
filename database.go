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
    status TEXT NOT NULL DEFAULT 'WISHLIST',
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
	_, err := db.Exec(schema)
	return err
}

// ---- Applications ----

const applicationCols = `id, company, position, status, reached_interview, resume_version, jd_text,
	applied_at, location, salary_range, contact_info, notes, created_at, updated_at`

func scanApplication(row interface{ Scan(...any) error }) (*Application, error) {
	var a Application
	err := row.Scan(&a.ID, &a.Company, &a.Position, &a.Status, &a.ReachedInterview,
		&a.ResumeVersion, &a.JDText, &a.AppliedAt, &a.Location, &a.SalaryRange,
		&a.ContactInfo, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
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
	var apps []*Application
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
	res, err := db.Exec(`INSERT INTO applications
		(company, position, status, resume_version, jd_text, applied_at, location, salary_range, contact_info, notes)
		VALUES (?, ?, 'WISHLIST', ?, ?, ?, ?, ?, ?, ?)`,
		in.Company, in.Position, in.ResumeVersion, in.JDText, in.AppliedAt,
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
//   - 首次进入 INTERVIEWING 时 reached_interview 置 1 (只增不改)
func updateApplicationStatus(id int64, status string) (*Application, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	a, err := getApplication(id)
	if err != nil {
		return nil, err
	}
	appliedAt := a.AppliedAt
	if status == "APPLIED" && appliedAt == "" {
		appliedAt = time.Now().Format("2006-01-02")
	}
	reached := a.ReachedInterview
	if status == "INTERVIEWING" && reached == 0 {
		reached = 1
	}
	_, err = db.Exec(`UPDATE applications SET status = ?, applied_at = ?, reached_interview = ? WHERE id = ?`,
		status, appliedAt, reached, id)
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
	var ivs []*Interview
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
	return &iv, nil
}

func updateInterview(id int64, in InterviewInput) (*Interview, error) {
	_, err := db.Exec(`UPDATE interviews SET round_name = ?, scheduled_at = ?, questions_and_notes = ?, outcome = ? WHERE id = ?`,
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
	return &iv, nil
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
	err := db.QueryRow(`SELECT
		COUNT(*) AS total,
		SUM(CASE WHEN status IN ('APPLIED','INTERVIEWING','OFFERED','REJECTED') THEN 1 ELSE 0 END) AS applied,
		SUM(CASE WHEN status = 'INTERVIEWING' THEN 1 ELSE 0 END) AS interviewing,
		SUM(reached_interview) AS reached,
		SUM(CASE WHEN status = 'OFFERED' THEN 1 ELSE 0 END) AS offered,
		SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END) AS rejected,
		SUM(CASE WHEN status = 'ARCHIVED' THEN 1 ELSE 0 END) AS archived
		FROM applications`).Scan(
		&s.TotalApplications, &s.TotalApplied, &s.Interviewing, &s.ReachedInterview,
		&s.Offered, &s.Rejected, &s.Archived)
	if err != nil {
		return nil, err
	}
	if s.TotalApplied > 0 {
		s.InterviewRate = float64(s.ReachedInterview) / float64(s.TotalApplied)
		s.OfferRate = float64(s.Offered) / float64(s.TotalApplied)
	}

	rows, err := db.Query(`SELECT resume_version,
		SUM(CASE WHEN status IN ('APPLIED','INTERVIEWING','OFFERED','REJECTED') THEN 1 ELSE 0 END) AS applied,
		SUM(reached_interview) AS reached
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
