import { Component } from "react";
import type { ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  error: string | null;
}

// 渲染期异常兜底: 避免整棵 React 树卸载导致白屏
export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(e: unknown): State {
    return { error: String(e instanceof Error ? e.message : e) };
  }

  componentDidCatch(error: unknown, info: unknown) {
    // WebView2 devtools 与后端 error.log 之外的前端错误兜底记录
    console.error("[ErrorBoundary]", error, info);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-3 p-8 text-center">
          <div className="text-sm font-semibold text-slate-700">
            界面出现异常
          </div>
          <div className="max-w-md break-all text-xs text-slate-400">
            {this.state.error}
          </div>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="rounded-lg bg-amber-600 px-4 py-2 text-sm font-medium text-white hover:bg-amber-700 transition-colors"
          >
            重新加载
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
