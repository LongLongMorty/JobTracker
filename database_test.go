package main

import (
	"database/sql"
	"testing"
)

func openTestDB(t *testing.T) {
	t.Helper()
	d, err := sql.Open("sqlite", "file::memory:?cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	d.SetMaxOpenConns(1)
	db = d
	if err := migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestApplicationLifecycle(t *testing.T) {
	openTestDB(t)

	a, err := insertApplication(ApplicationInput{Company: "Acme", Position: "Backend Engineer", ResumeVersion: "v1.0"})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if a.Status != "WISHLIST" {
		t.Fatalf("expected WISHLIST, got %s", a.Status)
	}

	// 状态 -> APPLIED 应自动填充 applied_at
	a, err = updateApplicationStatus(a.ID, "APPLIED")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if a.AppliedAt == "" {
		t.Fatal("applied_at should be auto-filled")
	}

	// 状态 -> INTERVIEWING 应置 reached_interview = 1
	a, err = updateApplicationStatus(a.ID, "INTERVIEWING")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if a.ReachedInterview != 1 {
		t.Fatalf("reached_interview should be 1, got %d", a.ReachedInterview)
	}

	// 状态再变不会重置 reached_interview
	a, err = updateApplicationStatus(a.ID, "OFFERED")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if a.ReachedInterview != 1 {
		t.Fatalf("reached_interview should stay 1, got %d", a.ReachedInterview)
	}

	// 面试记录
	iv, err := insertInterview(InterviewInput{ApplicationID: a.ID, RoundName: "1st Technical", ScheduledAt: "2026-09-05T10:00"})
	if err != nil {
		t.Fatalf("insert interview: %v", err)
	}
	if iv.Outcome != "PENDING" {
		t.Fatalf("expected PENDING, got %s", iv.Outcome)
	}
	ivs, err := listInterviews(a.ID)
	if err != nil || len(ivs) != 1 {
		t.Fatalf("list interviews: %v len=%d", err, len(ivs))
	}

	// 级联删除
	if err := deleteApplication(a.ID); err != nil {
		t.Fatalf("delete app: %v", err)
	}
	ivs, _ = listInterviews(a.ID)
	if len(ivs) != 0 {
		t.Fatal("interviews should be cascade deleted")
	}

	// 统计
	b, _ := insertApplication(ApplicationInput{Company: "Beta", Position: "FE", ResumeVersion: "v2.0"})
	updateApplicationStatus(b.ID, "APPLIED")
	updateApplicationStatus(b.ID, "INTERVIEWING")
	st, err := getStats()
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.ReachedInterview != 1 || st.TotalApplied != 1 {
		t.Fatalf("stats mismatch: %+v", st)
	}
	if len(st.ByResumeVersion) != 1 {
		t.Fatalf("expected 1 resume version stat, got %d", len(st.ByResumeVersion))
	}
	if st.ByResumeVersion[0].ResumeVersion != "v2.0" || st.ByResumeVersion[0].ReachedInterview != 1 {
		t.Fatalf("unexpected resume stat: %+v", st.ByResumeVersion[0])
	}
}
