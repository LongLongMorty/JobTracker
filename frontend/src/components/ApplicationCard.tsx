import { memo, useState } from "react";
import { Calendar, MapPin, MessageSquare } from "lucide-react";
import type { Application } from "../types";
import { STATUS_META } from "../constants";

interface Props {
  app: Application;
  onOpen: (app: Application) => void;
}

function ApplicationCard({ app, onOpen }: Props) {
  const meta = STATUS_META[app.status];
  const interviewCount = app.interview_count ?? 0;
  const failedRound =
    app.status === "INTERVIEW_FAILED" && app.failed_round > 0
      ? app.failed_round
      : 0;
  const [dragging, setDragging] = useState(false);

  return (
    <div
      draggable
      role="button"
      tabIndex={0}
      aria-label={`${app.company} ${app.position}`}
      onDragStart={(e) => {
        setDragging(true);
        e.dataTransfer.setData("text/plain", String(app.id));
        e.dataTransfer.effectAllowed = "move";
      }}
      onDragEnd={() => setDragging(false)}
      onClick={() => onOpen(app)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onOpen(app);
        }
      }}
      className={`group cursor-pointer rounded-lg border-l-4 border border-slate-200 bg-white p-3 shadow-sm hover:shadow-md hover:-translate-y-px transition-all select-none focus-visible:ring-2 focus-visible:ring-amber-500 ${dragging ? "opacity-50 scale-[0.98]" : ""}`}
      style={{ borderLeftColor: meta.railHex }}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="text-[15px] font-semibold text-slate-800 truncate leading-snug">
            {app.company}
          </div>
          <div className="text-xs text-slate-500 truncate mt-0.5">
            {app.position}
          </div>
        </div>
        {interviewCount > 0 && (
          <span className="shrink-0 inline-flex items-center gap-1 text-[11px] text-slate-500 bg-slate-50 rounded-full px-1.5 py-0.5">
            <MessageSquare size={11} />
            {interviewCount}
          </span>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2.5 text-[11px] text-slate-400">
        {failedRound > 0 && (
          <span className="rounded-full bg-red-50 px-1.5 py-0.5 font-medium text-red-600">
            第{failedRound}轮挂
          </span>
        )}
        {app.resume_version && (
          <span className="font-mono text-slate-500">{app.resume_version}</span>
        )}
        {app.location && (
          <span className="inline-flex items-center gap-0.5">
            <MapPin size={11} />
            {app.location}
          </span>
        )}
        {app.applied_at && (
          <span className="inline-flex items-center gap-0.5">
            <Calendar size={11} />
            {app.applied_at}
          </span>
        )}
      </div>
    </div>
  );
}

// memo: 拖拽/编辑其他卡片时避免全列表重渲染 (onOpen 由 useCallback 保持稳定)
export default memo(ApplicationCard);
