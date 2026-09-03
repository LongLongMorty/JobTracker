# JobTracker 代码审查报告

> 审查范围：Go 后端（main.go / app.go / models.go / database.go / database_test.go）+ React 前端（App / api / constants / types / components/*）。
> 方法：逐文件通读 + 结合此前调试中真实复现的缺陷（白屏、SQLITE_BUSY、排序错乱）。
> 严重度：**P0 影响正确性** ｜ **P1 健壮性** ｜ **P2 可维护性** ｜ **P3 工程化**。
> 定位格式：`文件:行`。
> 修复状态：`[已修复]` 表示已按本报告整改并验证；`[待处理]` 表示有意保留/需要你决策。

---

## 第二轮审查（修复后复审 + 前端优化，2026-09-03）

第一轮整改后又做了一轮"优化 + 复审"（React.memo、拖拽反馈、搜索清除、新建直达详情、删除面试确认、Modal 焦点管理），复审发现 3 个由优化引入的新问题，均已当场修复：

### R2-1. `useEscClose` 焦点抢夺（严重，已修复）
- 现象：`useEffect` 依赖 `[onClose, rootRef]`，而 `onClose` 每次父组件渲染都是新引用 → effect 反复重挂 → `root.focus()` 会**抢走正在输入的焦点**（打字中断）。
- 修复（`Modal.tsx`）：用 `onCloseRef` 保存最新回调，deps 仅 `[rootRef]`，focus 只执行一次。

### R2-2. `handleDrop` 引用不稳定导致 memo 失效（已修复）
- 现象：`useCallback(..., [apps, refresh])` 依赖 `apps`，列表一变所有列 props 都变 → `memo(KanbanColumn)` 全部失效，性能优化白做。
- 修复（`App.tsx`）：`appsRef` 快照 + 依赖仅 `[refresh]`（refresh 为空依赖 useCallback，稳定）。

### R2-3. Modal 的 `root.focus()` 与 `autoFocus` 冲突（已修复）
- 现象：`ApplicationModal` 的"公司"输入框带 `autoFocus`，但 Modal 的 effect 在其后执行并聚焦容器，焦点被抢回。
- 修复：`if (root && !root.contains(document.activeElement)) root.focus()` —— 已有内部焦点时不抢。

### R2-4. 其余复审结论
- `ErrorBoundary` 补了 `componentDidCatch` 日志（此前只渲染降级 UI 不记录）。
- `listApplications` 清除了残留的 `defer rows.Close()` + 显式 `Close()` 双重关闭（幂等但冗余）。
- 二轮 CDP 回归全过：新增弹窗焦点在输入框、Esc 只关最上层、新建直达详情、搜索清除按钮生效；`go test` 全绿。

**二轮结论**：当前版本无已知 P0/P1 问题。剩余待处理项同第一轮标注（lint/CI、绑定层测试、详情/导出的完整 interviews N+1）。

---

## 第三轮（待处理项清偿，2026-09-03）

针对第一/二轮标注的 `[待处理]` 项：

### R3-1. P3-20 N+1 查询 `[已修复]`
- `listInterviews` 改为批量：先查面试，再用单条 `IN (...)` 查询一次性带出全部 QA（`listQAItemsByInterviewIDs`）。`getApplication` 高频路径从 1+N 条查询降为 2 条。
- 实测：含 11 条 QA 的面试记录完整带出，顺序正确。

### R3-2. P3-19 lint / format `[已修复]`
- 后端：`golangci-lint run` 已接入并清零（修复 4 处 errcheck：3 处 `defer tx.Rollback()` 改为显式忽略并注释，1 处测试遗漏的 `db.Exec` 错误检查）；`gofmt -l` 干净。
- 前端：接入 prettier（全量格式化），`package.json` 新增 `test` / `format` / `format:check` 脚本。
- CI：本地单机项目暂无远程仓库，未配置流水线；检查命令已写入 README（质量检查章节）。

### R3-3. P3-18 测试缺口 `[已修复·部分]`
- 前端：抽出纯函数 `src/lib/markdown.ts`，vitest 新增 6 个单测（结构/编号/多行单行化/空占位/复盘节/结果映射），`npm test` 全过。
- 后端绑定层（Wails runtime 依赖）仍无法脱离 GUI 单测，`[待处理]`（导出 CSV、状态机等核心逻辑已有 Go 测试覆盖）。

### R3-4. 事故记录（流程教训）
- 本轮曾用 PowerShell 5.1 的 `Get-Content -Raw` + `Set-Content` 批量改代码，**默认 ANSI 编码把 database.go 的中文注释写坏**（非法 UTF-8，编译失败）。
- 处置：整文件以 UTF-8 重写，`go test` + `golangci-lint` + `gofmt` 三重验证通过。
- 教训：改源码一律走编辑器/写入工具（UTF-8），禁止 PS 5.1 的 Get/Set-Content 管道改中文文件。

**三轮结论**：第一/二轮的全部 `[待处理]` 项中，仅"绑定层自动化测试"因 Wails runtime 依赖保留为已知限制，其余全部清偿。

---

## 总览

架构清晰（Go 数据层 + Wails 绑定 + React 看板），迁移/测试/导出具备，单人工具质量整体不错。主要风险集中在三处：**数据一致性（缺事务 / 读改写竞态）**、**乐观更新失败不回滚**、**若干"新面试未保存即操作"的边界 bug**。以下逐条列出。

---

## P0 — 正确性 / 数据一致性

### 1. 看板卡片的「面试数徽章」永远不显示 `[已修复]`
- 位置：`ApplicationCard.tsx` 改用 `app.interview_count`（原 `app.interviews?.length`）
- 根因：`listApplications()` 不填充 `Interviews`，列表里 `interviews` 恒为 `undefined`。
- 修复：`Application` 新增 `interview_count` 字段（`models.go`），`listApplications` 用一条 `GROUP BY` 批量填充（`database.go`），`getApplication` 同步设置；新增 `TestInterviewCountInList` 覆盖。
- 备注：列表仍不返回完整 interviews（避免 N+1），详情打开时按需加载。

### 2. 新建面试「未保存」时点「添加问题」会失败 `[已修复]`
- 位置：`DetailModal.tsx` `InterviewFormBody` 增加 `isTemp = initial.id < 0` 守卫
- 修复：临时面试（id<0）时隐藏「添加问题 / 在本条后添加」，禁用「Markdown 预览」，显示"先点「保存面试」创建记录"提示。CDP 实测：`hasHint=true, addBtnCount=0, previewDisabled=true`。

### 3. 多处写操作无事务，中途失败即数据错乱 `[已修复]`
- 修复：`insertQAItem` / `deleteQAItem`（重排 + 增删）整体包入 `db.Begin()` 事务；`updateApplicationStatus`、新增 `applyInterviewFailed` 均为事务内读改写。

### 4. 状态更新存在读-改-写（TOCTOU）竞态 `[已修复]`
- 修复：`updateApplicationStatus` 与 `applyInterviewFailed` 在**同一事务内**完成"读 → 派生 → 写"；派生规则抽为纯函数 `deriveStatusFields`（`database.go`），`computeFailedRound` 支持传入 `db`/`tx`（`queryer` 接口）。

---

## P1 — 健壮性 / 可靠性

### 5. SQLite 未设置 `busy_timeout` `[已修复]`
- DSN 增加 `&_pragma=busy_timeout(5000)`（`database.go` 及测试 DSN）。

### 6. 乐观更新失败不回滚、无提示 `[已修复]`
- `App.handleDrop`：`try/catch`，失败时 `refresh()` 回滚并 toast 提示"状态更新失败，已还原"。
- `DetailModal.changeStatus`：失败时 `refresh()` 回滚本地值。

### 7. `QAItemRow` 即时保存缺并发/去抖保护 `[已修复]`
- `save` 增加请求序号守卫（`useRef`）：只接受最后一次请求的返回，避免旧响应覆盖新输入。

### 8. 保存"新面试"后展开面板自动收起 `[已修复]`
- `saveInterview` 新建成功后 `setExpanded(created.id)`，面板保持展开。

### 9. 导出：同秒覆盖 + CSV 字段不全 `[已修复]`
- 时间戳加毫秒（`20060102-150405.000`）。
- `writeApplicationsCSV` 表头补齐 `reached_interview`、`failed_round`；新增 `TestExportCSV` 断言列与内容。

### 10. 无优雅关闭，错误只 `println` `[已修复]`
- `main.go` 注册 `OnShutdown` → `App.shutdown` 关闭 `db`。
- 新增 `writeErrorLog`：启动失败写入 `%APPDATA%/jobtracker/error.log`。

---

## P2 — 可维护性 / 一致性

### 11. `api.ts` 全量 `as unknown as` 强转 `[已修复·说明]`
- 保留强转（Wails 生成类型为宽松 string），但收敛为**唯一转换边界**并加注释：后端字段变更时 tsc 只在本文件报错，禁止组件内直接转换。

### 12. `ApplicationModal` 的「编辑」分支不可达（死代码） `[已修复]`
- 移除 `initial` 编辑分支，组件只负责"新增"；`App.tsx` 的 `modal` 状态简化为 `{status} | null`。

### 13. `STATUS_META.chip` 未使用 `[已修复]`
- 从 `constants.ts` 删除。

### 14. 状态派生逻辑分散三处 `[已修复]`
- 派生规则集中到 `deriveStatusFields` 纯函数；FAILED 同步走 `applyInterviewFailed` 单事务。

### 15. 缺 React 错误边界 `[已修复]`
- 新增 `ErrorBoundary.tsx`（显示错误 + 重新加载按钮），`main.tsx` 包裹 `<App/>`，避免白屏。

### 16. 可访问性与 spec 目标不符 `[已修复]`
- 卡片：`role="button"` + `tabIndex=0` + Enter/空格打开（`ApplicationCard.tsx`）。
- Modal：`role="dialog"` + `aria-modal` + `aria-label`；Esc 关闭且多层叠加时只关最上层（`data-modal-root` 层叠判定）。
- 看板列：`onDragLeave` 用 `relatedTarget` 判定真正离开列边界，消除子元素冒泡闪烁。

### 17. 生成的 Markdown 未转义用户内容 `[已修复]`
- 问题文本单行化（换行→空格），避免破坏 `### Qn.` 标题结构。

---

## P3 — 工程化

### 18. 测试覆盖缺口 `[已修复·部分]`
- 新增：`TestInterviewCountInList`、`TestExportCSV`（CSV 列与内容）、`openTestDB` 加 `t.Cleanup(d.Close)`。
- 绑定层（Wails runtime）与前端单测未覆盖，`[待处理]`（本地单人项目，收益有限）。

### 19. 无 lint / format / CI `[待处理]`
- 本机已装 golangci-lint，可接入 `golangci-lint run` + `npx prettier --check`；CI 视是否开源再议。

### 20. 性能随数据量退化 `[已修复·部分]`
- `App.refresh` 改为 `Promise.all` 并行拉列表与统计。
- `listApplications` 面试计数改单条 `GROUP BY`（消除 N+1 的计数部分）；详情/导出的完整 interviews 仍为逐条查询（当前量级无感，`[待处理]`）。

---

## 建议修复顺序（已完成）

1. ✅ P0-1 / P0-2（用户可直接感知的错误/缺失）
2. ✅ P0-3/4 + P1-5/6（事务、busy_timeout、乐观更新回滚）
3. ✅ P1-7/8/9（QA 保存并发、展开丢失、导出）
4. ✅ P2-11/14/15/16/17（类型/状态集中/错误边界/可访问性/MD 转义）
5. ✅ P3-18/20 部分 + `[待处理]` 项见上文标注

> 备注：行号随修复有变动，以上为最终行为描述。`[待处理]` 项如需继续处理（接入 lint/CI、补绑定层测试）随时告诉我。
