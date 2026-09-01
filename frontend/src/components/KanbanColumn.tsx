import {useState} from 'react';
import {Plus} from 'lucide-react';
import type {Application, Status} from '../types';
import {STATUS_META} from '../constants';
import ApplicationCard from './ApplicationCard';

interface Props {
    status: Status;
    apps: Application[];
    onOpen: (app: Application) => void;
    onDropOnColumn: (status: Status, appId: number) => void;
    onQuickAdd: (status: Status) => void;
}

export default function KanbanColumn({status, apps, onOpen, onDropOnColumn, onQuickAdd}: Props) {
    const [over, setOver] = useState(false);
    const meta = STATUS_META[status];

    return (
        <div
            className={`flex w-64 shrink-0 flex-col rounded-xl border bg-slate-100/60 ${
                over ? 'border-amber-400 ring-2 ring-amber-200/60' : 'border-slate-200'
            } transition-colors`}
            onDragOver={(e) => {
                e.preventDefault();
                setOver(true);
            }}
            onDragLeave={() => setOver(false)}
            onDrop={(e) => {
                e.preventDefault();
                setOver(false);
                const id = Number(e.dataTransfer.getData('text/plain'));
                if (id) onDropOnColumn(status, id);
            }}
        >
            <header className="flex items-center gap-2 px-3 pt-3 pb-2">
                <span className={`h-2 w-2 rounded-full ${meta.dot}`}/>
                <h2 className="text-sm font-semibold text-slate-700">{meta.label}</h2>
                <span className="ml-auto rounded-full bg-white px-2 py-0.5 text-xs font-mono text-slate-500 ring-1 ring-slate-200">
                    {apps.length}
                </span>
                <button
                    type="button"
                    onClick={() => onQuickAdd(status)}
                    title={`添加到${meta.label}`}
                    className="rounded-md p-1 text-slate-400 hover:bg-white hover:text-slate-700 transition-colors"
                >
                    <Plus size={16}/>
                </button>
            </header>

            <div className="flex flex-col gap-2 px-2 pb-2 overflow-y-auto max-h-[calc(100vh-11rem)]">
                {apps.map((app) => (
                    <ApplicationCard key={app.id} app={app} onOpen={onOpen}/>
                ))}
                {apps.length === 0 && (
                    <div className="rounded-lg border border-dashed border-slate-200 px-3 py-6 text-center text-xs text-slate-300">
                        暂无记录
                    </div>
                )}
            </div>
        </div>
    );
}
