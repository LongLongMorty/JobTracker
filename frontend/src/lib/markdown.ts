import type { Interview } from "../types";

// 生成面试的完整 Markdown 文档 (纯函数, 单测覆盖)
export function buildInterviewMarkdown(iv: Interview): string {
  const lines: string[] = [];
  lines.push(
    `## ${iv.round_name || "面试"}${iv.scheduled_at ? `（${iv.scheduled_at.replace("T", " ")}）` : ""}`,
  );
  lines.push("");
  const outcomeLabel =
    iv.outcome === "PASSED"
      ? "通过"
      : iv.outcome === "FAILED"
        ? "未通过"
        : "待定";
  lines.push(`**结果：** ${outcomeLabel}`);
  lines.push("");
  (iv.qa_items ?? []).forEach((q, i) => {
    // 问题单行化, 避免换行破坏 Markdown 标题结构
    lines.push(
      `### Q${i + 1}. ${(q.question || "（未填写问题）").replace(/\r?\n/g, " ")}`,
    );
    lines.push("");
    if (q.answer) {
      lines.push(`**我的回答：** ${q.answer}`);
      lines.push("");
    }
    if (q.reflection) {
      lines.push(`**复盘改进：** ${q.reflection}`);
      lines.push("");
    }
  });
  if (iv.questions_and_notes) {
    lines.push(`## 整体复盘`);
    lines.push("");
    lines.push(iv.questions_and_notes);
    lines.push("");
  }
  return lines.join("\n");
}
