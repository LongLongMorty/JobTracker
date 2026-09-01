import type {Status} from './types';

export const STATUSES: Status[] = ['APPLIED', 'INTERVIEWING', 'RESUME_REJECTED', 'INTERVIEW_FAILED', 'DECLINED', 'OFFERED'];

export const STATUS_META: Record<Status, {
    label: string;
    dot: string;
    railHex: string;
    chip: string;
}> = {
    APPLIED: {
        label: '已投递',
        dot: 'bg-blue-500',
        railHex: '#3b82f6',
        chip: 'bg-blue-50 text-blue-600',
    },
    INTERVIEWING: {
        label: '面试中',
        dot: 'bg-amber-500',
        railHex: '#f59e0b',
        chip: 'bg-amber-50 text-amber-700',
    },
    OFFERED: {
        label: 'Offer',
        dot: 'bg-emerald-500',
        railHex: '#10b981',
        chip: 'bg-emerald-50 text-emerald-700',
    },
    RESUME_REJECTED: {
        label: '简历挂',
        dot: 'bg-rose-500',
        railHex: '#f43f5e',
        chip: 'bg-rose-50 text-rose-600',
    },
    INTERVIEW_FAILED: {
        label: '面试挂',
        dot: 'bg-red-600',
        railHex: '#dc2626',
        chip: 'bg-red-50 text-red-600',
    },
    DECLINED: {
        label: '已拒绝',
        dot: 'bg-slate-400',
        railHex: '#94a3b8',
        chip: 'bg-slate-100 text-slate-600',
    },
};
