# 前言碎碎念
最近在找实习，然后一开始只是想像网上一样用excel记录一下投递和面试状态，今天突发奇想不如做一个桌面端的小助手帮助记录，可以省去大量填表时间而且管理起来更方便，于是我随即vibe coding了此项目，数据全部存在本地，非常符合项目定位.....
# JobTracker
单用户求职投递追踪桌面应用。基于 Kanban 看板管理求职进度，记录面试、复盘笔记，并统计各简历版本的面试转化率。

## 功能

- **Kanban 看板**：6 个状态列 —— 已投递 / 面试中 / Offer / 简历挂 / 面试挂 / 已拒绝（我拒绝对方）
  - 拖拽卡片即可更新状态
  - 新建记录或拖入「已投递」时自动填充投递日期
  - 首次进入「面试中」或「面试挂」自动标记 `reached_interview`（用于面试率统计，只增不改）
- **面试挂自动同步**：任一轮面试标记为「未通过」时，卡片自动移到「面试挂」列；改回后自动回到「面试中」
- **投递详情**：公司、岗位、简历版本、JD 原文、薪资、地点、联系方式、Markdown 备注
- **面试记录**：每个投递可添加多轮面试（轮次名称、时间、结果、Markdown 问题与复盘笔记）
- **统计**：投递总数 / 面试率 / Offer 率，以及按简历版本统计面试转化率
- **搜索**：按公司 / 岗位 / 简历版本 / 地点过滤
- **导出**：一键导出 CSV（applications / interviews 两张表）+ JSON 完整结构

## 技术栈

- **桌面框架**: [Wails v2](https://wails.io)（Go 后端 + 前端桥接）
- **前端**: React 19 + Vite + Tailwind CSS v4 + Lucide Icons + react-markdown
- **数据库**: SQLite（[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) 纯 Go 驱动，无需 CGO）
- 数据存储于用户配置目录：`%APPDATA%\jobtracker\app.db`（Windows）

## 环境要求

- Go ≥ 1.25
- Node.js ≥ 20
- Wails CLI v2：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## 开发

```bash
# 实时开发（前端热更新，浏览器调试端口 http://localhost:34115）
wails dev

# 运行后端测试
go test ./...

# 生产构建（输出到 build/bin/）
wails build
```

> 注意：`wails dev` / `wails build` 会自动重新生成 `frontend/wailsjs` 绑定文件；若修改了 Go 端 struct 或方法，无需手动操作。

## 项目结构

```
├── main.go            # 应用入口
├── app.go             # Wails 绑定方法 (Applications / Interviews / Stats / Export)
├── models.go          # 数据结构
├── database.go        # SQLite 连接、迁移、CRUD
├── database_test.go   # 数据库层测试
└── frontend/
    └── src/
        ├── App.tsx            # 主布局与状态
        ├── components/        # 看板列、卡片、各弹窗、Markdown 编辑器
        ├── api.ts             # 前端对 Go 绑定的封装
        ├── constants.ts       # 状态元数据（标签/颜色）
        └── types.ts           # 前端类型
```

## 数据库 Schema

- `applications`：投递记录（`status`：APPLIED / INTERVIEWING / OFFERED / RESUME_REJECTED / INTERVIEW_FAILED / DECLINED，含 `reached_interview`、`applied_at` 等）
- `interviews`：面试记录（`application_id` 外键，级联删除）
- 触发器：`applications` 更新时自动刷新 `updated_at`
- 索引：`status` 与 `application_id`
- 旧版本状态（WISHLIST / REJECTED / ARCHIVED）启动时自动迁移到新状态集
