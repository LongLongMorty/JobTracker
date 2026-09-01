#  Job Tracker Application Spec

## 1. Project Overview

A lightweight, single-user desktop application designed to track and manage job applications, interview schedules, and interview reflections. The tool must prioritize high responsiveness, rapid local data input, and keyboard accessibility.

## 2. Tech Stack & Architecture

- **Framework**: Wails v2 (Go backend + Frontend bridge)
- **Frontend**: React (Vite) + Tailwind CSS + Lucide Icons + `shadcn/ui` (optional, keep components minimal)
- **Database**: SQLite3 (using `mattn/go-sqlite3` or modern `modernc.org/sqlite` pure Go driver)
- **Storage**: Single local DB file stored in the user config directory (`~/.jobtracker/app.db`)

## 3. Core Features & Functional Requirements

### Feature 1: Kanban Board (Application Pipeline)

- **Columns** (6 statuses, 无备投列, 记录即已投递; 被动被拒拆分为简历挂/面试挂, 主动拒为已拒绝):
  1. `Applied` (已投递)
  2. `Interviewing` (面试中)
  3. `Offered` (已获 Offer)
  4. `Resume Rejected` (简历挂)
  5. `Interview Failed` (面试挂)
  6. `Declined` (已拒绝, 我拒绝对方)
- **Interactions**:
  - Drag-and-drop cards between columns to update status (`status` field).
  - Quick-add card from the Kanban header with minimal fields (Company, Role, Resume Version).
  - **`applied_at` 自动填充**: 新建记录时自动记录当天日期；卡片拖入 `APPLIED` 列时若为空也会自动填充；用户可在详情中手动修改。
  - **面试结果自动同步**: 任一轮面试标记为 `FAILED` 时，卡片自动移到「面试挂」；改回 `PASSED`/`PENDING` 时若当前为「面试挂」则回到「面试中」。

### Feature 2: Application Detailed Record

- **Fields**:
  - `id`: INTEGER PRIMARY KEY
  - `company`: TEXT (Required)
  - `position`: TEXT (Required, 岗位名称)
  - `status`: ENUM string (`APPLIED`, `INTERVIEWING`, `OFFERED`, `RESUME_REJECTED`, `INTERVIEW_FAILED`, `DECLINED`)
  - `resume_version`: TEXT (e.g., "v1.2-AI-Infra", "v2.0-Backend")
  - `jd_text`: TEXT (Raw Job Description)
  - `applied_at`: DATETIME (新建/拖入 Applied 列时自动填充, 可手动修改)
  - `location`: TEXT
  - `salary_range`: TEXT
  - `contact_info`: TEXT (Recruiter name/email/referral note)
  - `notes`: TEXT (Markdown supported)
  - `created_at` / `updated_at`: DATETIME (`updated_at` 由 trigger 自动更新)

### Feature 3: Interview Timelines & Reflection Notes

- Each application can have **multiple** interview records (`1-to-N`).
- **Interview Fields**:
  - `id`: INTEGER PRIMARY KEY
  - `application_id`: INTEGER (FK)
  - `round_name`: TEXT (e.g., "1st Technical", "HR Round")
  - `scheduled_at`: DATETIME
  - `questions_and_notes`: TEXT (Markdown: Question Bank, Answers, Improvements)
  - `outcome`: ENUM (`PENDING`, `PASSED`, `FAILED`)
- 某轮面试标记为 `FAILED` 时自动把所属申请的 `status` 同步为 `INTERVIEW_FAILED` (从 `FAILED` 改回时若状态为 `INTERVIEW_FAILED` 则回到 `INTERVIEWING`)

### Feature 4: Basic Analytics (Stats Drawer/Modal)

- Top metrics: Total Applied, Interview Rate, Offer Rate, plus 简历挂 / 面试挂 / 已拒绝 计数。
  - **Interview Rate** = 进入过面试的申请数 / 全部记录数。通过 `reached_interview` 标志统计（见 schema, 状态首次变为 `INTERVIEWING` 或 `INTERVIEW_FAILED` 时置 1, 只增不改）
  - **Offer Rate** = 已获 Offer 数 / 全部记录数
- Breakdown by Resume Version (shows which version gets the highest interview yield).

### Feature 5: Data Export (数据导出)

- 一键导出全部数据为 **CSV**（applications 与 interviews 分表导出）和 **JSON**（完整结构, 含外键关系）。
- 用于备份、迁移及日后写简历复盘参考。导出路径由用户选择（文件保存对话框）。

## 4. Database Schema (SQLite)

SQL

```
CREATE TABLE IF NOT EXISTS applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company TEXT NOT NULL,
    position TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'APPLIED', -- APPLIED / INTERVIEWING / OFFERED / RESUME_REJECTED / INTERVIEW_FAILED / DECLINED
    reached_interview INTEGER NOT NULL DEFAULT 0,  -- 首次进入 INTERVIEWING 或 INTERVIEW_FAILED 时置 1, 只增不改 (统计用)
    resume_version TEXT DEFAULT '',
    jd_text TEXT DEFAULT '',
    applied_at DATETIME,                        -- 新建/拖入 APPLIED 列时自动填充
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
    scheduled_at DATETIME NOT NULL,
    questions_and_notes TEXT DEFAULT '',
    outcome TEXT DEFAULT 'PENDING',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_interviews_application_id ON interviews(application_id);
```

> 旧版本状态迁移（启动时自动执行）：`WISHLIST`→`APPLIED`；`REJECTED` 且有 `reached_interview`→`INTERVIEW_FAILED`，否则→`RESUME_REJECTED`；`ARCHIVED`→`DECLINED`。

## 5. Development Steps for Coding Agent

```
Phase 1: Project Setup & Database Layer
  ├── Initialize Wails v2 app with React + Tailwind template
  ├── Setup Go SQLite connection and run auto-migration script
  └── Implement Go structs and CRUD methods for Applications and Interviews

Phase 2: Frontend Core & Kanban UI
  ├── Setup Tailwind & install dnd-kit (or HTML5 Drag and Drop)
  ├── Build Kanban layout with the 5 status columns
  ├── Implement Create Application Modal (with JD text area and Resume Version tag)
  └── Connect Drag-and-Drop event to call Wails Go backend binding to update status

Phase 3: Detail Modal & Interview Notes
  ├── Build Application Detail Drawer/Modal
  ├── Add dynamic list for adding/editing Interview Rounds per Application
  └── Integrate simple Markdown Editor or Preview for Interview Questions & Notes

Phase 4: Polish & Refinement
  ├── Add search/filter bar (filter by Company, Position, or Resume Version)
  ├── Implement quick metric stats at the top bar
  └── Implement Data Export (CSV / JSON) via Go backend + save dialog
```