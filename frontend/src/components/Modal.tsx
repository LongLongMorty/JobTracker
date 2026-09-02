import {X} from 'lucide-react';
import type {ReactNode} from 'react';

interface Props {
    title: string;
    onClose: () => void;
    children: ReactNode;
    footer?: ReactNode;
    width?: string;
    fullscreen?: boolean;
}

export default function Modal({title, onClose, children, footer, width = 'max-w-xl', fullscreen = false}: Props) {
    if (fullscreen) {
        return (
            <div className="fixed inset-0 z-50 flex flex-col bg-white">
                <header className="flex items-center justify-between border-b border-slate-100 px-5 py-3">
                    <h2 className="text-base font-semibold text-slate-800">{title}</h2>
                    <button
                        type="button"
                        onClick={onClose}
                        className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
                    >
                        <X size={18}/>
                    </button>
                </header>
                <div className="flex-1 overflow-y-auto px-6 py-5">{children}</div>
                {footer && (
                    <footer className="flex justify-end gap-2 border-t border-slate-100 px-5 py-3">
                        {footer}
                    </footer>
                )}
            </div>
        );
    }

    return (
        <div
            className="fixed inset-0 z-50 flex items-start justify-center bg-slate-900/40 backdrop-blur-[2px] p-6 overflow-y-auto"
            onMouseDown={(e) => {
                if (e.target === e.currentTarget) onClose();
            }}
        >
            <div className={`mt-4 w-full ${width} rounded-xl bg-white shadow-2xl ring-1 ring-slate-900/5`}>
                <header className="flex items-center justify-between border-b border-slate-100 px-5 py-3.5">
                    <h2 className="text-base font-semibold text-slate-800">{title}</h2>
                    <button
                        type="button"
                        onClick={onClose}
                        className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
                    >
                        <X size={18}/>
                    </button>
                </header>
                <div className="px-5 py-4">{children}</div>
                {footer && (
                    <footer className="flex justify-end gap-2 border-t border-slate-100 px-5 py-3.5">
                        {footer}
                    </footer>
                )}
            </div>
        </div>
    );
}