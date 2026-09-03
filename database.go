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
	db, err = sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
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

CREATE TABLE IF NOT EXISTS interview_qa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    interview_id INTEGER NOT NULL,
    question TEXT NOT NULL DEFAULT '',
    answer TEXT DEFAULT '',
    reflection TEXT DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(interview_id) REFERENCES interviews(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_interviews_application_id ON interviews(application_id);
CREATE INDEX IF NOT EXISTS idx_qa_interview_id ON interview_qa(interview_id);
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
	a.InterviewCount = int64(len(ivs))
	return a, nil
}

func listApplications() ([]*Application, error) {
	rows, err := db.Query("SELECT " + applicationCols + " FROM applications ORDER BY updated_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	apps := []*Application{}
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	// 必须先关闭外层 rows 再嵌套查询 (SetMaxOpenConns(1) 下避免死锁)
	rows.Close()
	if len(apps) > 0 {
		// 批量填充面试计数 (避免 N+1 逐条查询)
		crows, err := db.Query("SELECT application_id, COUNT(*) FROM interviews GROUP BY application_id")
		if err != nil {
			return nil, err
		}
		counts := map[int64]int64{}
		for crows.Next() {
			var appID, n int64
			if err := crows.Scan(&appID, &n); err != nil {
				crows.Close()
				return nil, err
			}
			counts[appID] = n
		}
		crows.Close()
		if err := crows.Err(); err != nil {
			return nil, err
		}
		for _, a := range apps {
			a.InterviewCount = counts[a.ID]
		}
	}
	return apps, nil
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

// deriveStatusFields 集中计算状态派生字段 (纯函数, 状态机规则唯一来源):
//   - 进入 APPLIED 且 applied_at 为空时自动填充当天日期
//   - 进入 INTERVIEWING / INTERVIEW_FAILED 时 reached_interview 置 1 (只增不改)
func deriveStatusFields(a *Application, status string) (appliedAt string, reached int64) {
	appliedAt = a.AppliedAt
	if status == StatusApplied && appliedAt == "" {
		appliedAt = time.Now().Format("2006-01-02")
	}
	reached = a.ReachedInterview
	if (status == StatusInterviewing || status == StatusInterviewFailed) && reached == 0 {
		reached = 1
	}
	return appliedAt, reached
}

// updateApplicationStatus 更新状态 (事务内读改写, 避免竞态):
//   - 派生字段规则见 deriveStatusFields
//   - 离开 INTERVIEW_FAILED 时根据现有 FAILED 面试重算 failed_round
func updateApplicationStatus(id int64, status string) (*Application, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck

	var a Application
	if err := tx.QueryRow("SELECT "+applicationCols+" FROM applications WHERE id = ?", id).Scan(
		&a.ID, &a.Company, &a.Position, &a.Status, &a.ReachedInterview,
		&a.ResumeVersion, &a.JDText, &a.AppliedAt, &a.Location, &a.SalaryRange,
		&a.ContactInfo, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &a.FailedRound); err != nil {
		return nil, err
	}
	appliedAt, reached := deriveStatusFields(&a, status)
	round := a.FailedRound
	if status != StatusInterviewFailed {
		r, err := computeFailedRound(tx, id)
		if err != nil {
			return nil, err
		}
		round = r
	}
	if _, err := tx.Exec(`UPDATE applications SET status = ?, applied_at = ?, reached_interview = ?, failed_round = ? WHERE id = ?`,
		status, appliedAt, reached, round, id); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
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
	ivs := []*Interview{}
	for rows.Next() {
		iv, err := scanInterview(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		ivs = append(ivs, iv)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	// 必须先关闭外层 rows 再嵌套查询 (SetMaxOpenConns(1) 下避免死锁)
	rows.Close()

	if len(ivs) > 0 {
		items, err := listQAItemsByInterviewIDs(ivs)
		if err != nil {
			return nil, err
		}
		for _, iv := range ivs {
			iv.QAItems = items[iv.ID]
		}
	}
	return ivs, nil
}

// listQAItemsByInterviewIDs 批量查询多个面试的 QA 条目 (单条 IN 查询, 消除 N+1)
func listQAItemsByInterviewIDs(ivs []*Interview) (map[int64][]*QAItem, error) {
	if len(ivs) == 0 {
		return map[int64][]*QAItem{}, nil
	}
	ids := make([]any, len(ivs))
	placeholders := make([]string, len(ivs))
	for i, iv := range ivs {
		ids[i] = iv.ID
		placeholders[i] = "?"
	}
	query := "SELECT " + qaCols + " FROM interview_qa WHERE interview_id IN (" +
		strings.Join(placeholders, ",") + ") ORDER BY sort_order, id"
	rows, err := db.Query(query, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[int64][]*QAItem{}
	for rows.Next() {
		q, err := scanQAItem(rows)
		if err != nil {
			return nil, err
		}
		result[q.InterviewID] = append(result[q.InterviewID], q)
	}
	return result, rows.Err()
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
//   - 某轮面试标记为 FAILED → 卡片自动移到面试挂, 并记录挂的轮次 (单事务)
//   - 从 FAILED 改回 PASSED/PENDING 且当前是面试挂 → 回到面试中
func syncStatusOnOutcome(applicationID int64, oldOutcome, newOutcome string) error {
	newOutcome = strings.ToUpper(newOutcome)
	oldOutcome = strings.ToUpper(oldOutcome)
	if newOutcome == oldOutcome {
		return nil
	}
	if newOutcome == "FAILED" {
		return applyInterviewFailed(applicationID)
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

// applyInterviewFailed 面试挂同步: 事务内一次完成 状态+reached_interview+failed_round
func applyInterviewFailed(applicationID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck

	var a Application
	if err := tx.QueryRow("SELECT "+applicationCols+" FROM applications WHERE id = ?", applicationID).Scan(
		&a.ID, &a.Company, &a.Position, &a.Status, &a.ReachedInterview,
		&a.ResumeVersion, &a.JDText, &a.AppliedAt, &a.Location, &a.SalaryRange,
		&a.ContactInfo, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &a.FailedRound); err != nil {
		return err
	}
	_, reached := deriveStatusFields(&a, StatusInterviewFailed)
	round, err := computeFailedRound(tx, applicationID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE applications SET status = ?, reached_interview = ?, failed_round = ? WHERE id = ?`,
		StatusInterviewFailed, reached, round, applicationID); err != nil {
		return err
	}
	return tx.Commit()
}

// queryer 抽象 db 与 tx, 便于事务内外复用
type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// computeFailedRound 计算该申请挂在第几轮:
// 按面试时间 (scheduled_at, 时间相同再按 id) 排序, FAILED 那条面试的 1-based 序号; 无 FAILED 面试则返回 0
func computeFailedRound(q queryer, applicationID int64) (int64, error) {
	rows, err := q.Query(`SELECT outcome FROM interviews WHERE application_id = ? ORDER BY scheduled_at, id`, applicationID)
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

// ---- Interview QA Items ----

const qaCols = `id, interview_id, question, answer, reflection, sort_order, created_at`

func scanQAItem(row interface{ Scan(...any) error }) (*QAItem, error) {
	var q QAItem
	err := row.Scan(&q.ID, &q.InterviewID, &q.Question, &q.Answer, &q.Reflection, &q.SortOrder, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func listQAItems(interviewID int64) ([]*QAItem, error) {
	rows, err := db.Query("SELECT "+qaCols+" FROM interview_qa WHERE interview_id = ? ORDER BY sort_order, id", interviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*QAItem{}
	for rows.Next() {
		q, err := scanQAItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, q)
	}
	return items, rows.Err()
}

func insertQAItem(in QAItemInput) (*QAItem, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck
	if in.SortOrder > 0 {
		// 插入到指定位置: 该位置及其后的条目后移一位
		if _, err := tx.Exec(`UPDATE interview_qa SET sort_order = sort_order + 1
			WHERE interview_id = ? AND sort_order >= ?`, in.InterviewID, in.SortOrder); err != nil {
			return nil, err
		}
	} else {
		// 追加到末尾
		if err := tx.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM interview_qa WHERE interview_id = ?`,
			in.InterviewID).Scan(&in.SortOrder); err != nil {
			return nil, err
		}
	}
	res, err := tx.Exec(`INSERT INTO interview_qa (interview_id, question, answer, reflection, sort_order)
		VALUES (?, ?, ?, ?, ?)`,
		in.InterviewID, in.Question, in.Answer, in.Reflection, in.SortOrder)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	var q QAItem
	err = db.QueryRow("SELECT "+qaCols+" FROM interview_qa WHERE id = ?", id).Scan(
		&q.ID, &q.InterviewID, &q.Question, &q.Answer, &q.Reflection, &q.SortOrder, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func updateQAItem(id int64, in QAItemInput) (*QAItem, error) {
	_, err := db.Exec(`UPDATE interview_qa SET question = ?, answer = ?, reflection = ? WHERE id = ?`,
		in.Question, in.Answer, in.Reflection, id)
	if err != nil {
		return nil, err
	}
	var q QAItem
	err = db.QueryRow("SELECT "+qaCols+" FROM interview_qa WHERE id = ?", id).Scan(
		&q.ID, &q.InterviewID, &q.Question, &q.Answer, &q.Reflection, &q.SortOrder, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func deleteQAItem(id int64) error {
	var interviewID, sortOrder int64
	err := db.QueryRow("SELECT interview_id, sort_order FROM interview_qa WHERE id = ?", id).Scan(&interviewID, &sortOrder)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck
	if _, err := tx.Exec("DELETE FROM interview_qa WHERE id = ?", id); err != nil {
		return err
	}
	// 紧凑化: 被删条目之后的 sort_order 前移一位
	if _, err := tx.Exec(`UPDATE interview_qa SET sort_order = sort_order - 1
		WHERE interview_id = ? AND sort_order > ?`, interviewID, sortOrder); err != nil {
		return err
	}
	return tx.Commit()
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
