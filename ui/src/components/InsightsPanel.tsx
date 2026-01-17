"use client";

import { motion } from "framer-motion";
import { AlertTriangle, XCircle, Info, Lightbulb } from "lucide-react";
import type { Insight } from "@/lib/types";
import { useTraceStore } from "@/lib/store";

interface InsightsPanelProps {
  insights: Insight[];
}

export function InsightsPanel({ insights }: InsightsPanelProps) {
  const { selectMessage } = useTraceStore();

  if (insights.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-32 text-zinc-500">
        <Lightbulb className="w-8 h-8 mb-2 opacity-50" />
        <p className="text-sm">No issues detected</p>
      </div>
    );
  }

  const sortedInsights = [...insights].sort((a, b) => {
    const priority = { error: 0, warning: 1, info: 2 };
    return priority[a.type] - priority[b.type];
  });

  return (
    <div className="space-y-2">
      {sortedInsights.map((insight, index) => (
        <InsightCard
          key={insight.id}
          insight={insight}
          index={index}
          onClick={() => insight.message_id && selectMessage(insight.message_id)}
        />
      ))}
    </div>
  );
}

interface InsightCardProps {
  insight: Insight;
  index: number;
  onClick: () => void;
}

function InsightCard({ insight, index, onClick }: InsightCardProps) {
  const config = {
    error: {
      icon: XCircle,
      color: "text-red-400",
      bg: "bg-red-500/10",
      border: "border-red-500/20",
    },
    warning: {
      icon: AlertTriangle,
      color: "text-orange-400",
      bg: "bg-orange-500/10",
      border: "border-orange-500/20",
    },
    info: {
      icon: Info,
      color: "text-blue-400",
      bg: "bg-blue-500/10",
      border: "border-blue-500/20",
    },
  };

  const { icon: Icon, color, bg, border } = config[insight.type];

  let parsedDetails: Record<string, unknown> | null = null;
  try {
    parsedDetails = JSON.parse(insight.details);
  } catch {
    // Details is not JSON
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.2, delay: index * 0.05 }}
      onClick={onClick}
      className={`p-3 rounded-lg border ${bg} ${border} cursor-pointer hover:border-zinc-600 transition-colors`}
    >
      <div className="flex items-start gap-3">
        <Icon className={`w-5 h-5 ${color} flex-shrink-0 mt-0.5`} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium text-zinc-200">{insight.title}</span>
            <span className="text-xs text-zinc-500 px-1.5 py-0.5 rounded bg-zinc-800">
              {insight.category.replace("_", " ")}
            </span>
          </div>
          
          {parsedDetails ? (
            <div className="space-y-1 text-sm text-zinc-400">
              {parsedDetails.duration_ms !== undefined && (
                <p>Duration: <span className="text-zinc-300">{String(parsedDetails.duration_ms)}ms</span></p>
              )}
              {parsedDetails.method !== undefined && (
                <p>Method: <span className="font-mono text-zinc-300">{String(parsedDetails.method)}</span></p>
              )}
              {parsedDetails.error !== undefined && (
                <p>Error: <span className="text-red-400">{String(parsedDetails.error)}</span></p>
              )}
              {parsedDetails.suggestion !== undefined && (
                <p className="text-blue-400 mt-2">
                  💡 {String(parsedDetails.suggestion)}
                </p>
              )}
            </div>
          ) : (
            <p className="text-sm text-zinc-400">{insight.details}</p>
          )}
        </div>
      </div>
    </motion.div>
  );
}

