import { useEffect, useRef } from "react";
import { X } from "lucide-react";
import type { ReactNode } from "react";

interface Props {
  title: string;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
  width?: string;
  fullscreen?: boolean;
}

// Esc 关闭: 多层弹窗叠加时只关闭最上层 (data-modal-root 层叠顺序)
function useEscClose(
  onClose: () => void,
  rootRef: React.RefObject<HTMLDivElement | null>,
) {
  // onClose 每次渲染都是新引用, 用 ref 保存最新值, 避免 effect 反复重挂 (会抢输入焦点)
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    // 打开时把焦点移入弹窗 (若已有内部元素被 autoFocus 则不抢)
    const root = rootRef.current;
    if (root && !root.contains(document.activeElement)) {
      root.focus();
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      const modals = document.querySelectorAll("[data-modal-root]");
      const last = modals[modals.length - 1];
      if (rootRef.current && last === rootRef.current) onCloseRef.current();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [rootRef]);
}

export default function Modal({
  title,
  onClose,
  children,
  footer,
  width = "max-w-xl",
  fullscreen = false,
}: Props) {
  const rootRef = useRef<HTMLDivElement | null>(null);
  useEscClose(onClose, rootRef);

  if (fullscreen) {
    return (
      <div
        ref={rootRef}
        data-modal-root
        role="dialog"
        aria-modal="true"
        aria-label={title}
        tabIndex={-1}
        className="fixed inset-0 z-50 flex flex-col bg-white outline-none"
      >
        <header className="flex items-center justify-between border-b border-slate-100 px-5 py-3">
          <h2 className="text-base font-semibold text-slate-800">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="关闭"
            className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
          >
            <X size={18} />
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
      ref={rootRef}
      data-modal-root
      role="dialog"
      aria-modal="true"
      aria-label={title}
      tabIndex={-1}
      className="fixed inset-0 z-50 flex items-start justify-center bg-slate-900/40 backdrop-blur-[2px] p-6 overflow-y-auto outline-none"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className={`mt-4 w-full ${width} rounded-xl bg-white shadow-2xl ring-1 ring-slate-900/5`}
      >
        <header className="flex items-center justify-between border-b border-slate-100 px-5 py-3.5">
          <h2 className="text-base font-semibold text-slate-800">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="关闭"
            className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
          >
            <X size={18} />
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
