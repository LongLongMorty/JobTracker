package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func openTestDB(t *testing.T) {
	t.Helper()
	d, err := sql.Open("sqlite", "file:memdb"+t.Name()+"?mode=memory&cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	d.SetMaxOpenConns(1)
	t.Cleanup(func() { d.Close() })
	db = d
	if err := migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestApplicationLifecycle(t *testing.T) {
	openTestDB(t)

	// 新建记录默认 APPLIED, 且自动填充投递日期
	a, err := insertApplication(ApplicationInput{Company: "Acme", Position: "Backend Engineer", ResumeVersion: "v1.0"})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if a.Status != StatusApplied {
		t.Fatalf("expected APPLIED, got %s", a.Status)
	}
	if a.AppliedAt == "" {
		t.Fatal("applied_at should be auto-filled on insert")
	}

	// 进入 INTERVIEWING 置 reached_interview = 1
	a, err = updateApplicationStatus(a.ID, StatusInterviewing)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if a.ReachedInterview != 1 {
		t.Fatalf("reached_interview should be 1, got %d", a.ReachedInterview)
	}

	// 面试记录标记 FAILED → 申请自动变为面试挂, 记录轮次
	iv, err := insertInterview(InterviewInput{ApplicationID: a.ID, RoundName: "1st Technical", ScheduledAt: "2026-09-05T10:00", Outcome: "FAILED"})
	if err != nil {
		t.Fatalf("insert interview: %v", err)
	}
	a, _ = getApplication(a.ID)
	if a.Status != StatusInterviewFailed {
		t.Fatalf("expected INTERVIEW_FAILED after FAILED interview, got %s", a.Status)
	}
	if a.FailedRound != 1 {
		t.Fatalf("failed_round should be 1, got %d", a.FailedRound)
	}

	// 面试结果改回 PASSED → 状态回到面试中, 轮次清除
	if _, err = updateInterview(iv.ID, InterviewInput{ApplicationID: a.ID, RoundName: "1st Technical", ScheduledAt: "2026-09-05T10:00", Outcome: "PASSED"}); err != nil {
		t.Fatalf("update interview: %v", err)
	}
	a, _ = getApplication(a.ID)
	if a.Status != StatusInterviewing {
		t.Fatalf("expected INTERVIEWING after outcome reverted, got %s", a.Status)
	}
	if a.FailedRound != 0 {
		t.Fatalf("failed_round should be cleared, got %d", a.FailedRound)
	}

	// 第二轮面试 FAILED → 轮次应为 2
	iv1 := iv
	iv2, err := insertInterview(InterviewInput{ApplicationID: a.ID, RoundName: "2nd Technical", ScheduledAt: "2026-09-08T14:00", Outcome: "FAILED"})
	if err != nil {
		t.Fatalf("insert 2nd interview: %v", err)
	}
	a, _ = getApplication(a.ID)
	if a.Status != StatusInterviewFailed || a.FailedRound != 2 {
		t.Fatalf("expected INTERVIEW_FAILED round 2, got %s/%d", a.Status, a.FailedRound)
	}

	// 手动拖离面试挂 → failed_round 重算 (仍有 FAILED 面试时保留轮次)
	if _, err = updateApplicationStatus(a.ID, StatusApplied); err != nil {
		t.Fatalf("drag away: %v", err)
	}
	a, _ = getApplication(a.ID)
	if a.FailedRound != 2 {
		t.Fatalf("failed_round should stay 2 while FAILED interview exists, got %d", a.FailedRound)
	}
	_ = iv1
	_ = iv2

	// 补录场景: 先创建时间晚的面试, 再补录时间早的并标记 FAILED
	// 轮次应按面试时间排序 → 补录的"1st"时间最早 → 挂在第 1 轮
	c, err := insertApplication(ApplicationInput{Company: "Gamma", Position: "Ops"})
	if err != nil {
		t.Fatalf("insert c: %v", err)
	}
	late, err := insertInterview(InterviewInput{ApplicationID: c.ID, RoundName: "2nd Technical", ScheduledAt: "2026-09-08T14:00", Outcome: "PASSED"})
	if err != nil {
		t.Fatalf("insert late interview: %v", err)
	}
	early, err := insertInterview(InterviewInput{ApplicationID: c.ID, RoundName: "1st Technical", ScheduledAt: "2026-09-05T10:00", Outcome: "PASSED"})
	if err != nil {
		t.Fatalf("insert early interview: %v", err)
	}
	// 把补录的"1st"标记为 FAILED → 应按时间排序挂在第 1 轮
	if _, err = updateInterview(early.ID, InterviewInput{ApplicationID: c.ID, RoundName: "1st Technical", ScheduledAt: "2026-09-05T10:00", Outcome: "FAILED"}); err != nil {
		t.Fatalf("fail early interview: %v", err)
	}
	c, _ = getApplication(c.ID)
	if c.Status != StatusInterviewFailed || c.FailedRound != 1 {
		t.Fatalf("expected INTERVIEW_FAILED round 1 (time-sorted), got %s/%d", c.Status, c.FailedRound)
	}
	// 若改为按创建顺序则会是 2, 校验区分
	_ = late

	// 逐条问题记录: 增删改查
	qa, err := insertQAItem(QAItemInput{InterviewID: iv2.ID, Question: "如何设计高并发系统？", Answer: "分片+缓存", Reflection: "多聊限流"})
	if err != nil {
		t.Fatalf("insert qa: %v", err)
	}
	if qa.SortOrder != 1 {
		t.Fatalf("sort_order should be 1, got %d", qa.SortOrder)
	}
	qa2, err := insertQAItem(QAItemInput{InterviewID: iv2.ID, Question: "HTTP 与 HTTPS 区别？", Answer: "TLS", Reflection: ""})
	if err != nil {
		t.Fatalf("insert qa2: %v", err)
	}
	if qa2.SortOrder != 2 {
		t.Fatalf("sort_order should be 2, got %d", qa2.SortOrder)
	}
	items, err := listQAItems(iv2.ID)
	if err != nil || len(items) != 2 {
		t.Fatalf("list qa: %v len=%d", err, len(items))
	}
	if items[0].Question != "如何设计高并发系统？" {
		t.Fatalf("qa order wrong: %s", items[0].Question)
	}
	updated, err := updateQAItem(qa.ID, QAItemInput{InterviewID: iv2.ID, Question: "如何设计高并发订单系统？", Answer: "分片+缓存+MQ", Reflection: "注意热点拆分"})
	if err != nil || updated.Question != "如何设计高并发订单系统？" {
		t.Fatalf("update qa: %v %+v", err, updated)
	}
	// 插入到中间位置: sort_order 重排 ([qa(1), qa2(2)] → [qa(1), mid(2), qa2(3)])
	mid, err := insertQAItem(QAItemInput{InterviewID: iv2.ID, Question: "插入中间的问题？", Answer: "", Reflection: "", SortOrder: 2})
	if err != nil {
		t.Fatalf("insert mid qa: %v", err)
	}
	if mid.SortOrder != 2 {
		t.Fatalf("mid sort_order should be 2, got %d", mid.SortOrder)
	}
	items, _ = listQAItems(iv2.ID)
	if len(items) != 3 {
		t.Fatalf("after mid insert, len should be 3, got %d", len(items))
	}
	want := []string{"如何设计高并发订单系统？", "插入中间的问题？", "HTTP 与 HTTPS 区别？"}
	for i, it := range items {
		if it.Question != want[i] {
			t.Fatalf("order wrong at %d: got %q want %q", i, it.Question, want[i])
		}
	}
	// 删除中间条目后顺序连续
	if err := deleteQAItem(mid.ID); err != nil {
		t.Fatalf("delete mid qa: %v", err)
	}
	items, _ = listQAItems(iv2.ID)
	if items[1].Question != "HTTP 与 HTTPS 区别？" || items[1].SortOrder != 2 {
		t.Fatalf("after delete mid, order should be compact: %+v", items)
	}
	// 删除末尾条目
	if err := deleteQAItem(qa2.ID); err != nil {
		t.Fatalf("delete qa2: %v", err)
	}
	items, _ = listQAItems(iv2.ID)
	if len(items) != 1 {
		t.Fatalf("after delete, qa len should be 1, got %d", len(items))
	}
	// 面试记录应带出 QA
	ivs, _ := listInterviews(a.ID)
	for _, iv := range ivs {
		if iv.ID == iv2.ID && len(iv.QAItems) != 1 {
			t.Fatalf("interview should carry qa items, got %d", len(iv.QAItems))
		}
	}

	// 删除整个面试 → QA 级联删除
	if err := deleteInterview(iv2.ID); err != nil {
		t.Fatalf("delete interview: %v", err)
	}
	items, _ = listQAItems(iv2.ID)
	if len(items) != 0 {
		t.Fatal("qa items should be cascade deleted with interview")
	}

	// 创建时显式指定其他状态 (如已拒绝)
	b, err := insertApplication(ApplicationInput{Company: "Beta", Position: "FE", ResumeVersion: "v2.0", Status: StatusDeclined})
	if err != nil {
		t.Fatalf("insert b: %v", err)
	}
	if b.Status != StatusDeclined {
		t.Fatalf("expected DECLINED, got %s", b.Status)
	}

	// 统计: 3 条记录 (Acme/Beta/Gamma), 2 条进过面试
	st, err := getStats()
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.TotalApplications != 3 || st.ReachedInterview != 2 || st.Offered != 0 {
		t.Fatalf("stats mismatch: %+v", st)
	}
	if st.InterviewRate <= 0.665 || st.InterviewRate >= 0.667 {
		t.Fatalf("interview rate should be 2/3, got %f", st.InterviewRate)
	}
	if len(st.ByResumeVersion) != 2 {
		t.Fatalf("expected 2 resume version stats, got %d", len(st.ByResumeVersion))
	}

	// 级联删除
	if err := deleteApplication(a.ID); err != nil {
		t.Fatalf("delete app: %v", err)
	}
	ivs, _ = listInterviews(a.ID)
	if len(ivs) != 0 {
		t.Fatal("interviews should be cascade deleted")
	}
}

func TestLegacyStatusMigration(t *testing.T) {
	openTestDB(t)

	// 模拟旧版本数据
	res, err := db.Exec(`INSERT INTO applications (company, position, status, reached_interview) VALUES ('Old', 'X', 'WISHLIST', 0)`)
	if err != nil {
		t.Fatalf("seed legacy: %v", err)
	}
	id, _ := res.LastInsertId()
	if _, err := db.Exec(`UPDATE applications SET status = 'REJECTED', reached_interview = 1 WHERE id = ?`, id); err != nil {
		t.Fatalf("seed legacy status: %v", err)
	}

	if err := migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	a, err := getApplication(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if a.Status != StatusInterviewFailed {
		t.Fatalf("legacy REJECTED+reached should map to INTERVIEW_FAILED, got %s", a.Status)
	}
}

// 空数据库: 列表应返回 [] 而非 null, 统计应返回 0 而非报错
func TestEmptyDatabase(t *testing.T) {
	openTestDB(t)

	apps, err := listApplications()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if apps == nil {
		t.Fatal("listApplications should return empty slice, not nil")
	}

	ivs, err := listInterviews(1)
	if err != nil {
		t.Fatalf("list interviews: %v", err)
	}
	if ivs == nil {
		t.Fatal("listInterviews should return empty slice, not nil")
	}

	st, err := getStats()
	if err != nil {
		t.Fatalf("stats on empty db: %v", err)
	}
	if st.TotalApplications != 0 || st.InterviewRate != 0 || st.OfferRate != 0 {
		t.Fatalf("empty stats mismatch: %+v", st)
	}
	if st.ByResumeVersion == nil {
		t.Fatal("by_resume_version should be empty slice, not nil")
	}
}

// 列表应批量填充 interview_count (看板徽章依赖)
func TestInterviewCountInList(t *testing.T) {
	openTestDB(t)

	a, err := insertApplication(ApplicationInput{Company: "Acme", Position: "BE"})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := insertInterview(InterviewInput{ApplicationID: a.ID, RoundName: "1st", ScheduledAt: "2026-09-01T10:00"}); err != nil {
		t.Fatalf("insert interview: %v", err)
	}

	apps, err := listApplications()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(apps) != 1 || apps[0].InterviewCount != 1 {
		t.Fatalf("interview_count should be 1 in list, got %d", apps[0].InterviewCount)
	}
	detail, err := getApplication(a.ID)
	if err != nil || detail.InterviewCount != 1 {
		t.Fatalf("interview_count should be 1 in detail, got %d", detail.InterviewCount)
	}
}

// CSV 导出: 表头应包含派生字段, 内容与数据库一致
func TestExportCSV(t *testing.T) {
	openTestDB(t)

	a, err := insertApplication(ApplicationInput{Company: "Acme", Position: "BE", Status: StatusInterviewFailed, Notes: "复盘"})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := insertInterview(InterviewInput{ApplicationID: a.ID, RoundName: "1st", ScheduledAt: "2026-09-01T10:00", Outcome: "FAILED"}); err != nil {
		t.Fatalf("insert interview: %v", err)
	}
	apps, err := listApplications()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, ap := range apps {
		ivs, _ := listInterviews(ap.ID)
		ap.Interviews = ivs
	}

	dir := t.TempDir()
	appsPath := filepath.Join(dir, "apps.csv")
	ivsPath := filepath.Join(dir, "ivs.csv")
	if err := writeApplicationsCSV(appsPath, apps); err != nil {
		t.Fatalf("write apps csv: %v", err)
	}
	if err := writeInterviewsCSV(ivsPath, apps); err != nil {
		t.Fatalf("write ivs csv: %v", err)
	}

	appsData, err := os.ReadFile(appsPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	appsCSV := string(appsData)
	for _, col := range []string{"reached_interview", "failed_round", "updated_at"} {
		if !strings.Contains(appsCSV, col) {
			t.Fatalf("apps csv missing column %q", col)
		}
	}
	if !strings.Contains(appsCSV, "Acme") {
		t.Fatal("apps csv missing company data")
	}

	ivsData, err := os.ReadFile(ivsPath)
	if err != nil {
		t.Fatalf("read ivs: %v", err)
	}
	if !strings.Contains(string(ivsData), "qa_items") {
		t.Fatal("ivs csv missing qa_items column")
	}
}
