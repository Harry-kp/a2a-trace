"use client";

import { motion } from "framer-motion";
import { ArrowRight, CheckCircle2, XCircle, Clock, AlertTriangle } from "lucide-react";
import type { TimelineItem } from "@/lib/types";
import { useTraceStore } from "@/lib/store";

interface TimelineProps {
  items: TimelineItem[];
}

export function Timeline({ items }: TimelineProps) {
  const { selectMessage, selectedMessageId } = useTraceStore();

  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-zinc-500">
        <Clock className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-lg">Waiting for A2A traffic...</p>
        <p className="text-sm mt-2">Messages will appear here in real-time</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {items.map((item, index) => (
        <TimelineRow
          key={item.id}
          item={item}
          index={index}
          isSelected={selectedMessageId === item.id}
          onSelect={() => selectMessage(item.id)}
        />
      ))}
    </div>
  );
}

interface TimelineRowProps {
  item: TimelineItem;
  index: number;
  isSelected: boolean;
  onSelect: () => void;
}

function TimelineRow({ item, index, isSelected, onSelect }: TimelineRowProps) {
  const statusConfig = {
    pending: {
      icon: Clock,
      color: "text-yellow-500",
      bg: "bg-yellow-500/10",
      border: "border-yellow-500/20",
    },
    success: {
      icon: CheckCircle2,
      color: "text-emerald-500",
      bg: "bg-emerald-500/10",
      border: "border-emerald-500/20",
    },
    error: {
      icon: XCircle,
      color: "text-red-500",
      bg: "bg-red-500/10",
      border: "border-red-500/20",
    },
    slow: {
      icon: AlertTriangle,
      color: "text-orange-500",
      bg: "bg-orange-500/10",
      border: "border-orange-500/20",
    },
  };

  const config = statusConfig[item.status];
  const StatusIcon = config.icon;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2, delay: index * 0.05 }}
      onClick={onSelect}
      className={`
        relative p-4 rounded-lg border cursor-pointer transition-all
        ${isSelected 
          ? "bg-blue-500/10 border-blue-500/30" 
          : `${config.bg} ${config.border} hover:border-zinc-600`
        }
      `}
    >
      {/* Timeline connector */}
      {index > 0 && (
        <div className="absolute -top-2 left-6 w-0.5 h-2 bg-zinc-700" />
      )}

      <div className="flex items-center gap-4">
        {/* Status Icon */}
        <div className={`p-2 rounded-full ${config.bg}`}>
          <StatusIcon className={`w-4 h-4 ${config.color}`} />
        </div>

        {/* Agents */}
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span className="text-sm font-medium text-zinc-300 truncate max-w-[120px]">
            {item.fromAgent}
          </span>
          <ArrowRight className="w-4 h-4 text-zinc-600 flex-shrink-0" />
          <span className="text-sm font-medium text-zinc-300 truncate max-w-[120px]">
            {item.toAgent}
          </span>
        </div>

        {/* Method */}
        <div className="px-2 py-1 rounded bg-zinc-800 text-xs font-mono text-zinc-400">
          {item.method || "HTTP"}
        </div>

        {/* Duration */}
        {item.duration !== undefined && (
          <div className={`text-sm font-mono ${item.duration > 1000 ? "text-orange-400" : "text-zinc-500"}`}>
            {item.duration}ms
          </div>
        )}

        {/* Timestamp */}
        <div className="text-xs text-zinc-600 font-mono">
          {formatTime(item.timestamp)}
        </div>
      </div>

      {/* Error message preview */}
      {item.status === "error" && item.response?.error && (
        <div className="mt-2 text-sm text-red-400 truncate">
          {typeof item.response.error === "string" ? item.response.error : "Error occurred"}
        </div>
      )}
    </motion.div>
  );
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

