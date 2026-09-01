import {useState} from 'react';
import Modal from './Modal';
import MarkdownEditor from './MarkdownEditor';
import type {Application, ApplicationInput} from '../types';

interface Props {
    initial: Application | null; // null = 新建
    defaultStatus: string;
    onClose: () => void;
    onSave: (inp: ApplicationInput, id: number | null) => Promise<void>;
}

const EMPTY = {
    company: '',
    position: '',
    resume_version: '',
    applied_at: '',
    location: '',
    salary_range: '',
    contact_info: '',
    jd_text: '',
    notes: '',
};

export default function ApplicationModal({initial, defaultStatus, onClose, onSave}: Props) {
    const [form, setForm] = useState<Omit<ApplicationInput, 'company' | 'position'>>(() =>
        initial
            ? {
                  resume_version: initial.resume_version,
                  applied_at: initial.applied_at,
                  location: initial.location,
                  salary_range: initial.salary_range,
                  contact_info: initial.contact_info,
                  jd_text: initial.jd_text,
                  notes: initial.notes,
              }
            : {...EMPTY},
    );
    const [company, setCompany] = useState(initial?.company ?? '');
    const [position, setPosition] = useState(initial?.position ?? '');
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');

    const set = (k: keyof typeof form, v: string) => setForm((f) => ({...f, [k]: v}));

    const submit = async () => {
        if (!company.trim() || !position.trim()) {
            setError('公司和岗位名称为必填');
            return;
        }
        setError('');
        setSaving(true);
        try {
            await onSave({...form, company: company.trim(), position: position.trim()}, initial?.id ?? null);
        } catch (e: any) {
            setError(String(e?.message ?? e));
            setSaving(false);
        }
    };

    const inputCls = 'w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-amber-500 focus:ring-2 focus:ring-amber-200 focus:outline-none';
    const labelCls = 'block text-xs font-medium text-slate-500 mb-1';

    return (
        <Modal
            title={initial ? '编辑投递记录' : '新增投递记录'}
            onClose={onClose}
            footer={
                <>
                    <button
                        type="button"
                        onClick={onClose}
                        className="rounded-lg px-4 py-2 text-sm text-slate-500 hover:bg-slate-100 transition-colors"
                    >
                        取消
                    </button>
                    <button
                        type="button"
                        onClick={submit}
                        disabled={saving}
                        className="rounded-lg bg-amber-600 px-4 py-2 text-sm font-medium text-white hover:bg-amber-700 disabled:opacity-50 transition-colors"
                    >
                        {saving ? '保存中…' : initial ? '保存修改' : '创建'}
                    </button>
                </>
            }
        >
            <div className="space-y-4">
                {error && <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-600">{error}</div>}
                <div className="grid grid-cols-2 gap-3">
                    <div>
                        <label className={labelCls}>公司 *</label>
                        <input autoFocus className={inputCls} value={company} onChange={(e) => setCompany(e.target.value)} placeholder="如 Acme Inc."/>
                    </div>
                    <div>
                        <label className={labelCls}>岗位名称 *</label>
                        <input className={inputCls} value={position} onChange={(e) => setPosition(e.target.value)} placeholder="如 Backend Engineer"/>
                    </div>
                    <div>
                        <label className={labelCls}>简历版本</label>
                        <input className={inputCls} value={form.resume_version} onChange={(e) => set('resume_version', e.target.value)} placeholder="如 v2.0-Backend"/>
                    </div>
                    <div>
                        <label className={labelCls}>投递日期</label>
                        <input type="date" className={inputCls} value={form.applied_at} onChange={(e) => set('applied_at', e.target.value)}/>
                    </div>
                    <div>
                        <label className={labelCls}>地点</label>
                        <input className={inputCls} value={form.location} onChange={(e) => set('location', e.target.value)} placeholder="如 远程 / 上海"/>
                    </div>
                    <div>
                        <label className={labelCls}>薪资范围</label>
                        <input className={inputCls} value={form.salary_range} onChange={(e) => set('salary_range', e.target.value)} placeholder="如 30-40k"/>
                    </div>
                </div>
                <div>
                    <label className={labelCls}>联系方式 / 内推备注</label>
                    <input className={inputCls} value={form.contact_info} onChange={(e) => set('contact_info', e.target.value)} placeholder="Recruiter 姓名 / 邮箱 / 内推人"/>
                </div>
                <div>
                    <label className={labelCls}>JD 原文</label>
                    <textarea
                        rows={4}
                        className={inputCls}
                        value={form.jd_text}
                        onChange={(e) => set('jd_text', e.target.value)}
                        placeholder="粘贴岗位描述…"
                    />
                </div>
                <div>
                    <label className={labelCls}>备注（Markdown）</label>
                    <MarkdownEditor value={form.notes} onChange={(v) => set('notes', v)} placeholder="记录想法、进展…"/>
                </div>
            </div>
        </Modal>
    );
}
