import Modal from './Modal';
import type {Stats} from '../types';

interface Props {
    stats: Stats | null;
    onClose: () => void;
}

function pct(v: number): string {
    return `${Math.round(v * 100)}%`;
}

export default function StatsModal({stats, onClose}: Props) {
    const s = stats;

    const metrics = s
        ? [
              {label: '累计投递', value: String(s.total_applications), sub: '全部记录'},
              {label: '面试率', value: pct(s.interview_rate), sub: `进入面试 ${s.reached_interview}`},
              {label: 'Offer 率', value: pct(s.offer_rate), sub: `Offer ${s.offered}`},
              {label: '面试中', value: String(s.interviewing), sub: '进行中'},
              {label: '简历挂', value: String(s.resume_rejected), sub: '简历阶段'},
              {label: '面试挂', value: String(s.interview_failed), sub: '面试阶段'},
              {label: '已拒绝', value: String(s.declined), sub: '我拒绝对方'},
          ]
        : [];

    return (
        <Modal title="数据统计" onClose={onClose} width="max-w-2xl">
            <div className="space-y-6">
                {s ? (
                    <>
                        <div className="grid grid-cols-4 gap-3">
                            {metrics.map((m) => (
                                <div key={m.label} className="rounded-xl border border-slate-200 bg-slate-50/50 p-3.5">
                                    <div className="text-[11px] text-slate-400">{m.label}</div>
                                    <div className="mt-1 text-2xl font-semibold font-mono text-slate-800">{m.value}</div>
                                    <div className="mt-0.5 text-[11px] text-slate-400">{m.sub}</div>
                                </div>
                            ))}
                        </div>

                        <div>
                            <h3 className="mb-2 text-sm font-semibold text-slate-700">按简历版本 · 面试率</h3>
                            {s.by_resume_version.length === 0 ? (
                                <div className="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-xs text-slate-300">
                                    还没有带简历版本的投递记录
                                </div>
                            ) : (
                                <div className="space-y-2.5">
                                    {s.by_resume_version.map((v) => {
                                        const w = Math.min(100, Math.max(4, v.interview_rate * 100));
                                        return (
                                            <div key={v.resume_version} className="flex items-center gap-3">
                                                <span className="w-36 shrink-0 truncate font-mono text-xs text-slate-600">{v.resume_version}</span>
                                                <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-slate-100">
                                                    <div
                                                        className="h-full rounded-full bg-gradient-to-r from-amber-500 to-amber-600"
                                                        style={{width: `${w}%`}}
                                                    />
                                                </div>
                                                <span className="w-24 shrink-0 text-right font-mono text-xs text-slate-500">
                                                    {v.reached_interview}/{v.total_applied} · {pct(v.interview_rate)}
                                                </span>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    </>
                ) : (
                    <div className="py-10 text-center text-sm text-slate-400">加载中…</div>
                )}
            </div>
        </Modal>
    );
}
