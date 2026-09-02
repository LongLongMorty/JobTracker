package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := initDB(); err != nil {
		runtime.LogErrorf(ctx, "初始化数据库失败: %v", err)
	}
}

// ---- Applications ----

func (a *App) ListApplications() ([]*Application, error) {
	return listApplications()
}

func (a *App) GetApplication(id int64) (*Application, error) {
	return getApplication(id)
}

func (a *App) CreateApplication(in ApplicationInput) (*Application, error) {
	return insertApplication(in)
}

func (a *App) UpdateApplication(id int64, in ApplicationInput) (*Application, error) {
	return updateApplication(id, in)
}

func (a *App) UpdateApplicationStatus(id int64, status string) (*Application, error) {
	return updateApplicationStatus(id, status)
}

func (a *App) DeleteApplication(id int64) error {
	return deleteApplication(id)
}

// ---- Interviews ----

func (a *App) ListInterviews(applicationID int64) ([]*Interview, error) {
	return listInterviews(applicationID)
}

func (a *App) CreateInterview(in InterviewInput) (*Interview, error) {
	return insertInterview(in)
}

func (a *App) UpdateInterview(id int64, in InterviewInput) (*Interview, error) {
	return updateInterview(id, in)
}

func (a *App) DeleteInterview(id int64) error {
	return deleteInterview(id)
}

// ---- Interview QA Items ----

func (a *App) CreateQAItem(in QAItemInput) (*QAItem, error) {
	return insertQAItem(in)
}

func (a *App) UpdateQAItem(id int64, in QAItemInput) (*QAItem, error) {
	return updateQAItem(id, in)
}

func (a *App) DeleteQAItem(id int64) error {
	return deleteQAItem(id)
}

// ---- Stats ----

func (a *App) GetStats() (*Stats, error) {
	return getStats()
}

// ExportData 让用户选择导出目录, 写入 CSV(两个文件) 和 JSON(完整结构)。
// 返回导出目录; 用户取消时返回空字符串。
func (a *App) ExportData() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择导出文件夹",
	})
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", nil
	}

	apps, err := listApplications()
	if err != nil {
		return "", err
	}
	for _, ap := range apps {
		ivs, err := listInterviews(ap.ID)
		if err != nil {
			return "", err
		}
		ap.Interviews = ivs
	}

	ts := time.Now().Format("20060102-150405")

	// JSON (完整结构, 面试嵌套在申请内)
	jsonData, err := json.MarshalIndent(map[string]any{"exported_at": time.Now().Format(time.RFC3339), "applications": apps}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("jobtracker-export-%s.json", ts)), jsonData, 0o644); err != nil {
		return "", err
	}

	// CSV applications
	if err := writeApplicationsCSV(filepath.Join(dir, fmt.Sprintf("jobtracker-applications-%s.csv", ts)), apps); err != nil {
		return "", err
	}

	// CSV interviews
	if err := writeInterviewsCSV(filepath.Join(dir, fmt.Sprintf("jobtracker-interviews-%s.csv", ts)), apps); err != nil {
		return "", err
	}

	return dir, nil
}

func writeApplicationsCSV(path string, apps []*Application) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"id", "company", "position", "status", "resume_version", "jd_text", "applied_at", "location", "salary_range", "contact_info", "notes", "created_at", "updated_at"}); err != nil {
		return err
	}
	for _, a := range apps {
		if err := w.Write([]string{
			itoa(a.ID), a.Company, a.Position, a.Status, a.ResumeVersion, a.JDText,
			a.AppliedAt, a.Location, a.SalaryRange, a.ContactInfo, a.Notes, a.CreatedAt, a.UpdatedAt,
		}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeInterviewsCSV(path string, apps []*Application) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"id", "application_id", "round_name", "scheduled_at", "questions_and_notes", "outcome", "created_at", "qa_items"}); err != nil {
		return err
	}
	for _, a := range apps {
		for _, iv := range a.Interviews {
			qaJSON := ""
			if len(iv.QAItems) > 0 {
				if b, err := json.Marshal(iv.QAItems); err == nil {
					qaJSON = string(b)
				}
			}
			if err := w.Write([]string{
				itoa(iv.ID), itoa(iv.ApplicationID), iv.RoundName, iv.ScheduledAt,
				iv.QuestionsAndNotes, iv.Outcome, iv.CreatedAt, qaJSON,
			}); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
