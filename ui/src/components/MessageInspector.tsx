"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  X, 
  Copy, 
  Check, 
  ChevronDown, 
  ChevronRight,
  ArrowUpRight,
  ArrowDownLeft,
  Clock,
  AlertTriangle
} from "lucide-react";
import type { ParsedMessage } from "@/lib/types";
import { useTraceStore } from "@/lib/store";

export function MessageInspector() {
  const { getSelectedMessage, selectMessage, messages } = useTraceStore();
  const selectedMessage = getSelectedMessage();

  if (!selectedMessage) {
    return null;
  }

  // Find the paired request/response
  const pairedMessage = messages.find((m) => {
    if (selectedMessage.direction === "request") {
      return m.direction === "response" && m.request_id === selectedMessage.id;
    }
    return m.direction === "request" && m.id === selectedMessage.request_id;
  });

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        exit={{ opacity: 0, x: 20 }}
        className="fixed right-0 top-0 bottom-0 w-[500px] bg-zinc-900 border-l border-zinc-800 shadow-2xl overflow-hidden flex flex-col"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-zinc-800">
          <div className="flex items-center gap-2">
            {selectedMessage.direction === "request" ? (
              <ArrowUpRight className="w-5 h-5 text-blue-400" />
            ) : (
              <ArrowDownLeft className="w-5 h-5 text-green-400" />
            )}
            <span className="font-semibold text-zinc-200">
              {selectedMessage.direction === "request" ? "Request" : "Response"}
            </span>
            {selectedMessage.method && (
              <span className="px-2 py-0.5 rounded bg-zinc-800 text-xs font-mono text-zinc-400">
                {selectedMessage.method}
              </span>
            )}
          </div>
          <button
            onClick={() => selectMessage(null)}
            className="p-1 rounded hover:bg-zinc-800 transition-colors"
          >
            <X className="w-5 h-5 text-zinc-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Meta info */}
          <div className="grid grid-cols-2 gap-4">
            <MetaItem label="URL" value={selectedMessage.url} mono />
            <MetaItem label="Status" value={selectedMessage.status_code?.toString() || "-"} />
            <MetaItem label="Duration" value={`${selectedMessage.duration_ms}ms`} />
            <MetaItem label="Size" value={formatBytes(selectedMessage.size)} />
          </div>

          {/* Error */}
          {selectedMessage.error && (
            <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20">
              <div className="flex items-center gap-2 text-red-400 mb-1">
                <AlertTriangle className="w-4 h-4" />
                <span className="font-medium">Error</span>
              </div>
              <p className="text-sm text-red-300">{selectedMessage.error}</p>
            </div>
          )}

          {/* Headers */}
          <CollapsibleSection title="Headers" defaultOpen={false}>
            <JsonViewer data={selectedMessage.headers} />
          </CollapsibleSection>

          {/* Body */}
          <CollapsibleSection title="Body" defaultOpen={true}>
            <JsonViewer data={selectedMessage.body} />
          </CollapsibleSection>

          {/* Paired message preview */}
          {pairedMessage && (
            <div className="mt-4 pt-4 border-t border-zinc-800">
              <button
                onClick={() => selectMessage(pairedMessage.id)}
                className="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300 transition-colors"
              >
                {selectedMessage.direction === "request" ? (
                  <>
                    <ArrowDownLeft className="w-4 h-4" />
                    View Response
                  </>
                ) : (
                  <>
                    <ArrowUpRight className="w-4 h-4" />
                    View Request
                  </>
                )}
              </button>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-zinc-800 flex gap-2">
          <CopyButton label="Copy JSON" data={selectedMessage.body} />
          <CopyButton label="Copy cURL" data={generateCurl(selectedMessage)} />
        </div>
      </motion.div>
    </AnimatePresence>
  );
}

function MetaItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <div className="text-xs text-zinc-500 mb-1">{label}</div>
      <div className={`text-sm text-zinc-300 truncate ${mono ? "font-mono" : ""}`}>
        {value}
      </div>
    </div>
  );
}

function CollapsibleSection({
  title,
  defaultOpen,
  children,
}: {
  title: string;
  defaultOpen: boolean;
  children: React.ReactNode;
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border border-zinc-800 rounded-lg overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center gap-2 p-3 bg-zinc-800/50 hover:bg-zinc-800 transition-colors"
      >
        {isOpen ? (
          <ChevronDown className="w-4 h-4 text-zinc-500" />
        ) : (
          <ChevronRight className="w-4 h-4 text-zinc-500" />
        )}
        <span className="text-sm font-medium text-zinc-300">{title}</span>
      </button>
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ height: 0 }}
            animate={{ height: "auto" }}
            exit={{ height: 0 }}
            className="overflow-hidden"
          >
            <div className="p-3 bg-zinc-950">{children}</div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

function JsonViewer({ data }: { data: unknown }) {
  const [copied, setCopied] = useState(false);

  const jsonString =
    typeof data === "string" ? data : JSON.stringify(data, null, 2);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(jsonString);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group">
      <pre className="text-xs font-mono text-zinc-400 overflow-x-auto whitespace-pre-wrap break-all max-h-[400px] overflow-y-auto">
        {jsonString || "(empty)"}
      </pre>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-1.5 rounded bg-zinc-800 opacity-0 group-hover:opacity-100 transition-opacity"
      >
        {copied ? (
          <Check className="w-3 h-3 text-emerald-400" />
        ) : (
          <Copy className="w-3 h-3 text-zinc-400" />
        )}
      </button>
    </div>
  );
}

function CopyButton({ label, data }: { label: string; data: unknown }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const text = typeof data === "string" ? data : JSON.stringify(data, null, 2);
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-2 px-3 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-sm text-zinc-300 transition-colors"
    >
      {copied ? (
        <Check className="w-4 h-4 text-emerald-400" />
      ) : (
        <Copy className="w-4 h-4" />
      )}
      {label}
    </button>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

function generateCurl(msg: ParsedMessage): string {
  const headers = Object.entries(msg.headers || {})
    .map(([key, value]) => `-H '${key}: ${value}'`)
    .join(" \\\n  ");

  const body =
    typeof msg.body === "string"
      ? msg.body
      : JSON.stringify(msg.body);

  return `curl -X POST '${msg.url}' \\
  ${headers} \\
  -d '${body}'`;
}

