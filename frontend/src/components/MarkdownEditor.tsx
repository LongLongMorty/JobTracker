import {useState} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface Props {
    value: string;
    onChange: (v: string) => void;
    placeholder?: string;
    rows?: number;
}

export default function MarkdownEditor({value, onChange, placeholder, rows = 6}: Props) {
    const [mode, setMode] = useState<'write' | 'preview'>('write');

    return (
        <div className="rounded-lg border border-slate-200 bg-white overflow-hidden">
            <div className="flex items-center justify-between border-b border-slate-100 px-2">
                <div className="flex gap-1">
                    {(['write', 'preview'] as const).map((m) => (
                        <button
                            key={m}
                            type="button"
                            onClick={() => setMode(m)}
                            className={`px-2.5 py-1.5 text-xs font-medium rounded-md transition-colors ${
                                mode === m ? 'bg-slate-100 text-slate-800' : 'text-slate-400 hover:text-slate-600'
                            }`}
                        >
                            {m === 'write' ? '编辑' : '预览'}
                        </button>
                    ))}
                </div>
                <span className="text-[10px] text-slate-300 font-mono">Markdown</span>
            </div>
            {mode === 'write' ? (
                <textarea
                    value={value}
                    rows={rows}
                    onChange={(e) => onChange(e.target.value)}
                    placeholder={placeholder}
                    className="w-full p-3 text-sm bg-transparent resize-y focus:outline-none placeholder:text-slate-300"
                />
            ) : (
                <div className="p-3 text-sm md-body max-h-[40vh] overflow-auto">
                    {value ? (
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>{value}</ReactMarkdown>
                    ) : (
                        <span className="text-slate-300">暂无内容</span>
                    )}
                </div>
            )}
        </div>
    );
}
