import type { Status } from "./types";

export const STATUSES: Status[] = [
  "APPLIED",
  "INTERVIEWING",
  "RESUME_REJECTED",
  "INTERVIEW_FAILED",
  "DECLINED",
  "OFFERED",
];

export const STATUS_META: Record<
  Status,
  {
    label: string;
    dot: string;
    railHex: string;
  }
> = {
  APPLIED: {
    label: "已投递",
    dot: "bg-blue-500",
    railHex: "#3b82f6",
  },
  INTERVIEWING: {
    label: "面试中",
    dot: "bg-amber-500",
    railHex: "#f59e0b",
  },
  OFFERED: {
    label: "Offer",
    dot: "bg-emerald-500",
    railHex: "#10b981",
  },
  RESUME_REJECTED: {
    label: "简历挂",
    dot: "bg-rose-500",
    railHex: "#f43f5e",
  },
  INTERVIEW_FAILED: {
    label: "面试挂",
    dot: "bg-red-600",
    railHex: "#dc2626",
  },
  DECLINED: {
    label: "已拒绝",
    dot: "bg-slate-400",
    railHex: "#94a3b8",
  },
};
