"use client";

import { motion } from "framer-motion";
import { Activity, Wifi, WifiOff, Download, RotateCcw, Terminal } from "lucide-react";
import type { Trace, Summary } from "@/lib/types";

interface HeaderProps {
  trace: Trace | null;
  summary: Summary | null;
  isConnected: boolean;
  onExport: () => void;
  onClear: () => void;
}

export function Header({ trace, summary, isConnected, onExport, onClear }: HeaderProps) {
  return (
    <header className="sticky top-0 z-50 bg-zinc-950/80 backdrop-blur-lg border-b border-zinc-800">
      <div className="max-w-7xl mx-auto px-4 py-3">
        <div className="flex items-center justify-between">
          {/* Logo and Status */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Activity className="w-6 h-6 text-blue-400" />
              <span className="text-xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                A2A Trace
              </span>
            </div>

            {/* Connection Status */}
            <div className="flex items-center gap-2">
              {isConnected ? (
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  className="flex items-center gap-1.5 px-2 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20"
                >
                  <Wifi className="w-3.5 h-3.5 text-emerald-400" />
                  <span className="text-xs text-emerald-400">Live</span>
                </motion.div>
              ) : (
                <div className="flex items-center gap-1.5 px-2 py-1 rounded-full bg-zinc-800 border border-zinc-700">
                  <WifiOff className="w-3.5 h-3.5 text-zinc-500" />
                  <span className="text-xs text-zinc-500">Disconnected</span>
                </div>
              )}
            </div>
          </div>

          {/* Stats */}
          {summary && (
            <div className="hidden md:flex items-center gap-6">
              <Stat label="Messages" value={summary.total_messages} />
              <Stat label="Errors" value={summary.error_count} isError={summary.error_count > 0} />
              <Stat label="Avg Latency" value={`${summary.avg_duration_ms}ms`} />
              <Stat label="Insights" value={summary.total_insights} />
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-2">
            {trace && (
              <div className="hidden sm:flex items-center gap-2 mr-4 px-3 py-1.5 rounded-lg bg-zinc-800/50">
                <Terminal className="w-4 h-4 text-zinc-500" />
                <span className="text-sm font-mono text-zinc-400 truncate max-w-[200px]">
                  {trace.command}
                </span>
              </div>
            )}

            <button
              onClick={onClear}
              className="p-2 rounded-lg hover:bg-zinc-800 text-zinc-500 hover:text-zinc-300 transition-colors"
              title="Clear traces"
            >
              <RotateCcw className="w-4 h-4" />
            </button>

            <button
              onClick={onExport}
              className="flex items-center gap-2 px-3 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium transition-colors"
            >
              <Download className="w-4 h-4" />
              Export
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}

function Stat({
  label,
  value,
  isError,
}: {
  label: string;
  value: string | number;
  isError?: boolean;
}) {
  return (
    <div className="text-center">
      <div
        className={`text-lg font-bold ${
          isError ? "text-red-400" : "text-zinc-200"
        }`}
      >
        {value}
      </div>
      <div className="text-xs text-zinc-500">{label}</div>
    </div>
  );
}

