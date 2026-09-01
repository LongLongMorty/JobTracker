import {useCallback, useEffect, useMemo, useState} from 'react';
import {Briefcase, Download, PieChart, Plus, Search, Loader2} from 'lucide-react';
import type {Application, ApplicationInput, Status} from './types';
import {STATUS_META, STATUSES} from './constants';
import {api} from './api';
import KanbanColumn from './components/KanbanColumn';
import ApplicationModal from './components/ApplicationModal';
import DetailModal from './components/DetailModal';
import StatsModal from './components/StatsModal';
import type {Stats} from './types';

export default function App() {
    const [apps, setApps] = useState<Application[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [stats, setStats] = useState<Stats | null>(null);
    const [statsOpen, setStatsOpen] = useState(false);

    // null = 关闭; {app: null, status} = 新建; {app} = 编辑
    const [modal, setModal] = useState<{app: Application | null; status: Status} | null>(null);
    const [detail, setDetail] = useState<Application | null>(null);
    const [toast, setToast] = useState('');

    const notify = (msg: string) => {
        setToast(msg);
        setTimeout(() => setToast(''), 2200);
    };

    const refresh = useCallback(async () => {
        const list = await api.listApplications();
        setApps(list);
        setStats(await api.getStats());
    }, []);

    useEffect(() => {
        (async () => {
            await refresh();
            setLoading(false);
        })();
    }, [refresh]);

    const openStats = async () => {
        setStatsOpen(true);
        setStats(await api.getStats());
    };

    const grouped = useMemo(() => {
        const q = search.trim().toLowerCase();
        const filtered = q
            ? apps.filter((a) =>
                  [a.company, a.position, a.resume_version, a.location].some((f) => f.toLowerCase().includes(q)),
              )
            : apps;
        const map = new Map<Status, Application[]>();
        for (const s of STATUSES) map.set(s, []);
        for (const a of filtered) {
            const arr = map.get(a.status) ?? [];
            arr.push(a);
            map.set(a.status, arr);
        }
        return map;
    }, [apps, search]);

    const handleDrop = async (status: Status, appId: number) => {
        const current = apps.find((a) => a.id === appId);
        if (!current || current.status === status) return;
        // 乐观更新
        setApps((prev) => prev.map((a) => (a.id === appId ? {...a, status} : a)));
        const updated = await api.updateStatus(appId, status);
        setApps((prev) => prev.map((a) => (a.id === appId ? updated : a)));
        notify(`已移至「${STATUS_META[status].label}」`);
    };

    const handleSave = async (inp: ApplicationInput, id: number | null) => {
        if (id == null) {
            await api.createApplication(inp);
            notify('已创建');
        } else {
            await api.updateApplication(id, inp);
            notify('已保存修改');
        }
        await refresh();
        setModal(null);
    };

    const handleExport = async () => {
        const dir = await api.exportData();
        if (dir) notify(`已导出到 ${dir}`);
    };

    const saveResult = (app: Application) => {
        setApps((prev) => prev.map((a) => (a.id === app.id ? app : a)));
    };

    if (loading) {
        return (
            <div className="flex h-full items-center justify-center text-slate-400">
                <Loader2 size={20} className="animate-spin mr-2"/>
                加载中…
            </div>
        );
    }

    return (
        <div className="flex h-full flex-col">
            {/* 顶部栏 */}
            <header className="flex items-center gap-4 border-b border-slate-200 bg-white px-5 py-3">
                <div className="flex items-center gap-2.5">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-800 text-amber-400">
                        <Briefcase size={17}/>
                    </div>
                    <div className="leading-tight">
                        <div className="font-mono text-sm font-bold tracking-tight text-slate-800">JobTracker</div>
                        <div className="text-[10px] text-slate-400">求职流水线</div>
                    </div>
                </div>

                <div className="relative ml-2 min-w-0 flex-1 max-w-md">
                    <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"/>
                    <input
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        placeholder="搜索公司 / 岗位 / 简历版本…"
                        className="w-full rounded-lg border border-slate-200 bg-slate-50 py-2 pl-9 pr-3 text-sm focus:border-amber-500 focus:ring-2 focus:ring-amber-200 focus:bg-white focus:outline-none transition-colors"
                    />
                </div>

                {/* 快速指标 (窄窗口隐藏, 避免头部溢出) */}
                <div className="ml-auto hidden shrink-0 items-center gap-2 font-mono text-xs text-slate-500 xl:flex">
                    <span>已投 <b className="text-slate-800">{stats ? stats.total_applications : '–'}</b></span>
                    <span className="text-slate-200">|</span>
                    <span>面试率 <b className="text-amber-600">{stats ? Math.round(stats.interview_rate * 100) + '%' : '–'}</b></span>
                    <span className="text-slate-200">|</span>
                    <span>Offer <b className="text-emerald-600">{stats ? Math.round(stats.offer_rate * 100) + '%' : '–'}</b></span>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                    <button
                        type="button"
                        onClick={openStats}
                        title="数据统计"
                        className="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-700 transition-colors"
                    >
                        <PieChart size={18}/>
                    </button>
                    <button
                        type="button"
                        onClick={handleExport}
                        title="导出数据"
                        className="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-700 transition-colors"
                    >
                        <Download size={18}/>
                    </button>
                    <button
                        type="button"
                        onClick={() => setModal({app: null, status: 'APPLIED'})}
                        className="inline-flex items-center gap-1.5 rounded-lg bg-amber-600 px-3.5 py-2 text-sm font-medium text-white hover:bg-amber-700 transition-colors"
                    >
                        <Plus size={16}/>
                        新增投递
                    </button>
                </div>
            </header>

            {/* 看板 */}
            <main className="flex-1 overflow-x-auto overflow-y-hidden p-4">
                <div className="flex h-full items-stretch gap-3">
                    {STATUSES.map((s) => (
                        <KanbanColumn
                            key={s}
                            status={s}
                            apps={grouped.get(s) ?? []}
                            onOpen={setDetail}
                            onDropOnColumn={handleDrop}
                            onQuickAdd={(status) => setModal({app: null, status})}
                        />
                    ))}
                </div>
            </main>

            {/* 弹窗 */}
            {modal && (
                <ApplicationModal
                    initial={modal.app}
                    defaultStatus={modal.status}
                    onClose={() => setModal(null)}
                    onSave={handleSave}
                />
            )}
            {detail && (
                <DetailModal
                    app={detail}
                    onClose={() => setDetail(null)}
                    onChanged={async () => {
                        await refresh();
                        if (detail) {
                            const fresh = await api.getApplication(detail.id);
                            setDetail(fresh);
                            saveResult(fresh);
                        }
                    }}
                />
            )}
            {statsOpen && <StatsModal stats={stats} onClose={() => setStatsOpen(false)}/>}

            {/* 轻提示 */}
            {toast && (
                <div className="fixed bottom-5 left-1/2 -translate-x-1/2 rounded-lg bg-slate-800 px-4 py-2 text-sm text-white shadow-lg">
                    {toast}
                </div>
            )}
        </div>
    );
}
