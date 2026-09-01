import type {Status} from './types';

export const STATUSES: Status[] = ['WISHLIST', 'APPLIED', 'INTERVIEWING', 'OFFERED', 'REJECTED', 'ARCHIVED'];

export const STATUS_META: Record<Status, {
    label: string;
    short: string;
    dot: string;
    railHex: string;
    chip: string;
}> = {
    WISHLIST: {
        label: '准备投递',
        short: '备投',
        dot: 'bg-slate-400',
        railHex: '#94a3b8',
        chip: 'bg-slate-100 text-slate-600',
    },
    APPLIED: {
        label: '已投递',
        short: '已投',
        dot: 'bg-blue-500',
        railHex: '#3b82f6',
        chip: 'bg-blue-50 text-blue-600',
    },
    INTERVIEWING: {
        label: '面试中',
        short: '面试',
        dot: 'bg-amber-500',
        railHex: '#f59e0b',
        chip: 'bg-amber-50 text-amber-700',
    },
    OFFERED: {
        label: 'Offer',
        short: 'Offer',
        dot: 'bg-emerald-500',
        railHex: '#10b981',
        chip: 'bg-emerald-50 text-emerald-700',
    },
    REJECTED: {
        label: '未通过',
        short: '未过',
        dot: 'bg-rose-500',
        railHex: '#f43f5e',
        chip: 'bg-rose-50 text-rose-600',
    },
    ARCHIVED: {
        label: '已归档',
        short: '归档',
        dot: 'bg-violet-400',
        railHex: '#a78bfa',
        chip: 'bg-violet-50 text-violet-600',
    },
};

export function fmtDate(s: string): string {
    if (!s) return '—';
    return s;
}
