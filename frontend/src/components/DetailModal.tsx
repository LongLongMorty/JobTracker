import {useEffect, useState} from 'react';
import {Plus, Trash2, ChevronDown, ChevronUp, CalendarClock} from 'lucide-react';
import Modal from './Modal';
import MarkdownEditor from './MarkdownEditor';
import type {Application, ApplicationInput, Interview, InterviewInput, Status} from '../types';
import {STATUS_META, STATUSES} from '../constants';
import {api} from '../api';

interface Props {
    app: Application;
    onClose: () => void;
    onChanged: () => Promise<void>;
}

export default function DetailModal({app, onClose, onChanged}: Props) {
    const [form, setForm] = useState({
        company: app.company,
        position: app.position,
        status: app.status,
        resume_version: app.resume_version,
        applied_at: app.applied_at,
        location: app.location,
        salary_range: app.salary_range,
        contact_info: app.contact_info,
        jd_text: app.jd_text,
        notes: app.notes,
    });
    const [interviews, setInterviews] = useState<Interview[]>(app.interviews ?? []);
    const [expanded, setExpanded] = useState<number | null>(null);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');

    const set = (k: keyof typeof form, v: string) => setForm((f) => ({...f, [k]: v}));

    // 面试挂自动同步等后端状态变化时, 同步到表单
    useEffect(() => {
        setForm((f) => (f.status === app.status ? f : {...f, status: app.status}));
    }, [app.status]);

    const refresh = async () => {
        await onChanged();
    };

    const changeStatus = async (status: string) => {
        set('status', status);
        try {
            await api.updateStatus(app.id, status);
            await refresh();
        } catch (e: any) {
            setError(String(e?.message ?? e));
        }
    };

    const saveRecord = async () => {
        if (!form.company.trim() || !form.position.trim()) {
            setError('公司和岗位名称为必填');
            return;
        }
        setBusy(true);
        setError('');
        try {
            const inp: ApplicationInput = {
                company: form.company.trim(),
                position: form.position.trim(),
                resume_version: form.resume_version,
                applied_at: form.applied_at,
                location: form.location,
                salary_range: form.salary_range,
                contact_info: form.contact_info,
                jd_text: form.jd_text,
                notes: form.notes,
                status: app.status,
            };
            await api.updateApplication(app.id, inp);
            await refresh();
            setBusy(false);
        } catch (e: any) {
            setError(String(e?.message ?? e));
            setBusy(false);
        }
    };

    const deleteRecord = async () => {
        if (!confirm(`确认删除「${app.company} - ${app.position}」？此操作不可恢复。`)) return;
        await api.deleteApplication(app.id);
        await refresh();
        onClose();
    };

    const addInterview = () => {
        const temp: Interview = {
            id: -Date.now(),
            application_id: app.id,
            round_name: '',
            scheduled_at: '',
            questions_and_notes: '',
            outcome: 'PENDING',
            created_at: '',
        };
        setInterviews((ivs) => [...ivs, temp]);
        setExpanded(temp.id);
    };

    const saveInterview = async (iv: Interview) => {
        const inp: InterviewInput = {
            application_id: app.id,
            round_name: iv.round_name,
            scheduled_at: iv.scheduled_at,
            questions_and_notes: iv.questions_and_notes,
            outcome: iv.outcome,
        };
        if (iv.id < 0) {
            const created = await api.createInterview(inp);
            setInterviews((ivs) => ivs.map((x) => (x.id === iv.id ? created : x)));
        } else {
            await api.updateInterview(iv.id, inp);
        }
        await refresh();
    };

    const removeInterview = async (iv: Interview) => {
        if (iv.id > 0) await api.deleteInterview(iv.id);
        setInterviews((ivs) => ivs.filter((x) => x.id !== iv.id));
        await refresh();
    };

    const inputCls = 'w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-amber-500 focus:ring-2 focus:ring-amber-200 focus:outline-none';
    const labelCls = 'block text-xs font-medium text-slate-500 mb-1';

    return (
        <Modal
            title={`${app.company} · ${app.position}`}
            onClose={onClose}
            width="max-w-3xl"
            footer={
                <>
                    <button
                        type="button"
                        onClick={deleteRecord}
                        className="mr-auto rounded-lg px-3 py-2 text-sm text-rose-500 hover:bg-rose-50 transition-colors"
                    >
                        <Trash2 size={15} className="inline mr-1"/>
                        删除
                    </button>
                    <button
                        type="button"
                        onClick={onClose}
                        className="rounded-lg px-4 py-2 text-sm text-slate-500 hover:bg-slate-100 transition-colors"
                    >
                        关闭
                    </button>
                    <button
                        type="button"
                        onClick={saveRecord}
                        disabled={busy}
                        className="rounded-lg bg-amber-600 px-4 py-2 text-sm font-medium text-white hover:bg-amber-700 disabled:opacity-50 transition-colors"
                    >
                        {busy ? '保存中…' : '保存修改'}
                    </button>
                </>
            }
        >
            <div className="space-y-5 max-h-[72vh] overflow-y-auto pr-1">
                {error && <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600">{error}</div>}

                {/* 基础信息 */}
                <section className="grid grid-cols-2 gap-3">
                    <div>
                        <label className={labelCls}>公司 *</label>
                        <input className={inputCls} value={form.company} onChange={(e) => set('company', e.target.value)}/>
                    </div>
                    <div>
                        <label className={labelCls}>岗位名称 *</label>
                        <input className={inputCls} value={form.position} onChange={(e) => set('position', e.target.value)}/>
                    </div>
                    <div>
                        <label className={labelCls}>状态</label>
                        <select className={inputCls} value={form.status} onChange={(e) => changeStatus(e.target.value)}>
                            {STATUSES.map((s) => (
                                <option key={s} value={s}>{STATUS_META[s].label}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className={labelCls}>简历版本</label>
                        <input className={inputCls} value={form.resume_version} onChange={(e) => set('resume_version', e.target.value)}/>
                    </div>
                    <div>
                        <label className={labelCls}>投递日期</label>
                        <input type="date" className={inputCls} value={form.applied_at} onChange={(e) => set('applied_at', e.target.value)}/>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <label className={labelCls}>地点</label>
                            <input className={inputCls} value={form.location} onChange={(e) => set('location', e.target.value)}/>
                        </div>
                        <div>
                            <label className={labelCls}>薪资范围</label>
                            <input className={inputCls} value={form.salary_range} onChange={(e) => set('salary_range', e.target.value)}/>
                        </div>
                    </div>
                    <div className="col-span-2">
                        <label className={labelCls}>联系方式 / 内推备注</label>
                        <input className={inputCls} value={form.contact_info} onChange={(e) => set('contact_info', e.target.value)}/>
                    </div>
                </section>

                <section>
                    <label className={labelCls}>JD 原文</label>
                    <textarea rows={4} className={inputCls} value={form.jd_text} onChange={(e) => set('jd_text', e.target.value)}/>
                </section>

                <section>
                    <label className={labelCls}>备注（Markdown）</label>
                    <MarkdownEditor value={form.notes} onChange={(v) => set('notes', v)} rows={5}/>
                </section>

                {/* 面试记录 */}
                <section>
                    <div className="mb-2 flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-slate-700">面试记录</h3>
                        <button
                            type="button"
                            onClick={addInterview}
                            className="inline-flex items-center gap-1 rounded-lg bg-slate-100 px-2.5 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-200 transition-colors"
                        >
                            <Plus size={14}/>
                            添加面试
                        </button>
                    </div>

                    {interviews.length === 0 && (
                        <div className="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-xs text-slate-300">
                            还没有面试记录，点击右上角添加
                        </div>
                    )}

                    <div className="space-y-2">
                        {interviews.map((iv) => {
                            const isOpen = expanded === iv.id;
                            const isTemp = iv.id < 0;
                            return (
                                <div key={iv.id} className="rounded-lg border border-slate-200 bg-white">
                                    <button
                                        type="button"
                                        onClick={() => setExpanded(isOpen ? null : iv.id)}
                                        className="flex w-full items-center gap-2 px-3 py-2.5 text-left"
                                    >
                                        <span className="text-slate-400">
                                            {isOpen ? <ChevronUp size={16}/> : <ChevronDown size={16}/>}
                                        </span>
                                        <span className={`h-2 w-2 rounded-full ${STATUS_META[iv.outcome === 'PASSED' ? 'OFFERED' : iv.outcome === 'FAILED' ? 'INTERVIEW_FAILED' : 'APPLIED'].dot}`}/>
                                        <span className="text-sm font-medium text-slate-700 flex-1 truncate">
                                            {iv.round_name || (isTemp ? '新面试' : '(未命名)')}
                                        </span>
                                        <span className="inline-flex items-center gap-1 text-xs text-slate-400 font-mono">
                                            <CalendarClock size={12}/>
                                            {iv.scheduled_at || '—'}
                                        </span>
                                        <span className={`rounded-full px-2 py-0.5 text-[11px] ${iv.outcome === 'PASSED' ? 'bg-emerald-50 text-emerald-600' : iv.outcome === 'FAILED' ? 'bg-rose-50 text-rose-600' : 'bg-slate-100 text-slate-500'}`}>
                                            {iv.outcome === 'PASSED' ? '通过' : iv.outcome === 'FAILED' ? '未通过' : '待定'}
                                        </span>
                                    </button>

                                    {isOpen && (
                                        <InterviewForm
                                            key={iv.id}
                                            initial={iv}
                                            onSave={saveInterview}
                                            onDelete={removeInterview}
                                        />
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </section>
            </div>
        </Modal>
    );
}

function InterviewForm({
    initial,
    onSave,
    onDelete,
}: {
    initial: Interview;
    onSave: (iv: Interview) => Promise<void>;
    onDelete: (iv: Interview) => Promise<void>;
}) {
    const [draft, setDraft] = useState({
        round_name: initial.round_name,
        scheduled_at: initial.scheduled_at,
        outcome: initial.outcome,
        questions_and_notes: initial.questions_and_notes,
    });
    const [saving, setSaving] = useState(false);
    const [msg, setMsg] = useState('');

    const inputCls = 'w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-amber-500 focus:ring-2 focus:ring-amber-200 focus:outline-none';
    const labelCls = 'block text-xs font-medium text-slate-500 mb-1';

    const submit = async () => {
        setSaving(true);
        setMsg('');
        try {
            await onSave({...initial, ...draft, round_name: draft.round_name.trim() || '面试'});
            setMsg('已保存');
        } catch (e: any) {
            setMsg(String(e?.message ?? e));
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="space-y-3 border-t border-slate-100 px-3 py-3">
            <div className="grid grid-cols-3 gap-3">
                <div>
                    <label className={labelCls}>轮次名称</label>
                    <input className={inputCls} value={draft.round_name} onChange={(e) => setDraft({...draft, round_name: e.target.value})} placeholder="如 1st Technical"/>
                </div>
                <div>
                    <label className={labelCls}>面试时间</label>
                    <input type="datetime-local" className={inputCls} value={draft.scheduled_at} onChange={(e) => setDraft({...draft, scheduled_at: e.target.value})}/>
                </div>
                <div>
                    <label className={labelCls}>结果</label>
                    <select className={inputCls} value={draft.outcome} onChange={(e) => setDraft({...draft, outcome: e.target.value as Interview['outcome']})}>
                        <option value="PENDING">待定</option>
                        <option value="PASSED">通过</option>
                        <option value="FAILED">未通过</option>
                    </select>
                </div>
            </div>
            <div>
                <label className={labelCls}>问题 / 复盘笔记（Markdown）</label>
                <MarkdownEditor
                    value={draft.questions_and_notes}
                    onChange={(v) => setDraft({...draft, questions_and_notes: v})}
                    placeholder={'**问题：**\n- 提问内容\n\n**回答/改进：**\n- 下次注意…'}
                    rows={4}
                />
            </div>
            <div className="flex items-center justify-between">
                <button
                    type="button"
                    onClick={() => onDelete(initial)}
                    className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs text-rose-500 hover:bg-rose-50 transition-colors"
                >
                    <Trash2 size={13}/>
                    删除
                </button>
                <div className="flex items-center gap-2">
                    {msg && <span className="text-xs text-slate-400">{msg}</span>}
                    <button
                        type="button"
                        onClick={submit}
                        disabled={saving}
                        className="rounded-lg bg-slate-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700 disabled:opacity-50 transition-colors"
                    >
                        {saving ? '保存中…' : '保存面试'}
                    </button>
                </div>
            </div>
        </div>
    );
}
