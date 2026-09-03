import { describe, expect, it } from "vitest";
import { buildInterviewMarkdown } from "./markdown";
import type { Interview } from "../types";

function makeInterview(overrides: Partial<Interview> = {}): Interview {
  return {
    id: 1,
    application_id: 1,
    round_name: "1st Technical",
    scheduled_at: "2026-09-03T10:00",
    questions_and_notes: "",
    outcome: "PENDING",
    created_at: "2026-09-03 10:00:00",
    qa_items: [],
    ...overrides,
  };
}

describe("buildInterviewMarkdown", () => {
  it("基础结构: 标题/时间/结果", () => {
    const md = buildInterviewMarkdown(makeInterview({ outcome: "PASSED" }));
    expect(md).toContain("## 1st Technical（2026-09-03 10:00）");
    expect(md).toContain("**结果：** 通过");
  });

  it("逐条 QA 按序生成并编号", () => {
    const iv = makeInterview({
      qa_items: [
        {
          id: 1,
          interview_id: 1,
          question: "Q-A",
          answer: "A1",
          reflection: "R1",
          sort_order: 1,
          created_at: "",
        },
        {
          id: 2,
          interview_id: 1,
          question: "Q-B",
          answer: "",
          reflection: "R2",
          sort_order: 2,
          created_at: "",
        },
      ],
    });
    const md = buildInterviewMarkdown(iv);
    expect(md).toContain("### Q1. Q-A");
    expect(md).toContain("**我的回答：** A1");
    expect(md).toContain("**复盘改进：** R1");
    expect(md).toContain("### Q2. Q-B");
    // 无回答的条目不生成"我的回答"段
    const q2AnswerIdx = md.indexOf("**我的回答：** A1");
    expect(md.indexOf("**我的回答：**", q2AnswerIdx + 1)).toBe(-1);
    // Q2 的复盘在
    expect(md).toContain("**复盘改进：** R2");
  });

  it("多行问题被单行化, 不破坏标题结构", () => {
    const iv = makeInterview({
      qa_items: [
        {
          id: 1,
          interview_id: 1,
          question: "第一行\n第二行",
          answer: "",
          reflection: "",
          sort_order: 1,
          created_at: "",
        },
      ],
    });
    const md = buildInterviewMarkdown(iv);
    expect(md).toContain("### Q1. 第一行 第二行");
    expect(md.split("\n").filter((l) => l.startsWith("### "))).toHaveLength(1);
  });

  it("空问题占位符", () => {
    const iv = makeInterview({
      qa_items: [
        {
          id: 1,
          interview_id: 1,
          question: "",
          answer: "",
          reflection: "",
          sort_order: 1,
          created_at: "",
        },
      ],
    });
    expect(buildInterviewMarkdown(iv)).toContain("### Q1. （未填写问题）");
  });

  it("整体复盘节仅在非空时出现", () => {
    expect(buildInterviewMarkdown(makeInterview())).not.toContain("整体复盘");
    expect(
      buildInterviewMarkdown(
        makeInterview({ questions_and_notes: "表现不错" }),
      ),
    ).toContain("## 整体复盘");
  });

  it("结果标签映射: PENDING/PASSED/FAILED", () => {
    expect(
      buildInterviewMarkdown(makeInterview({ outcome: "PENDING" })),
    ).toContain("待定");
    expect(
      buildInterviewMarkdown(makeInterview({ outcome: "PASSED" })),
    ).toContain("通过");
    expect(
      buildInterviewMarkdown(makeInterview({ outcome: "FAILED" })),
    ).toContain("未通过");
  });
});
