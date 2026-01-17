"use client";

import { useEffect, useState, useCallback } from "react";
import { motion } from "framer-motion";
import { Activity, MessageSquare, Bot, Lightbulb, ChevronRight } from "lucide-react";
import { Header } from "@/components/Header";
import { Timeline } from "@/components/Timeline";
import { MessageInspector } from "@/components/MessageInspector";
import { InsightsPanel } from "@/components/InsightsPanel";
import { AgentList } from "@/components/AgentNode";
import { useWebSocket } from "@/hooks/useWebSocket";
import { useTraceStore } from "@/lib/store";
import type { Summary } from "@/lib/types";

type TabType = "timeline" | "agents" | "insights";

export default function Home() {
  const [activeTab, setActiveTab] = useState<TabType>("timeline");
  const [summary, setSummary] = useState<Summary | null>(null);

  const {
    trace,
    messages,
    agents,
    insights,
    selectedMessageId,
    setTrace,
    addMessage,
    setMessages,
    addAgent,
    setAgents,
    addInsight,
    setInsights,
    setConnected,
    clearAll,
    getTimelineItems,
  } = useTraceStore();

  // Determine WebSocket URL based on current location
  const wsUrl =
    typeof window !== "undefined"
      ? `ws://${window.location.host}/ws`
      : "ws://localhost:8080/ws";

  // Fetch initial data
  const fetchData = useCallback(async () => {
    try {
      const baseUrl =
        typeof window !== "undefined"
          ? `${window.location.protocol}//${window.location.host}`
          : "http://localhost:8080";

      const [traceRes, messagesRes, agentsRes, insightsRes, summaryRes] =
        await Promise.all([
          fetch(`${baseUrl}/api/trace`),
          fetch(`${baseUrl}/api/messages`),
          fetch(`${baseUrl}/api/agents`),
          fetch(`${baseUrl}/api/insights`),
          fetch(`${baseUrl}/api/summary`),
        ]);

      if (traceRes.ok) {
        const traceData = await traceRes.json();
        setTrace(traceData);
      }
      if (messagesRes.ok) {
        const messagesData = await messagesRes.json();
        setMessages(messagesData || []);
      }
      if (agentsRes.ok) {
        const agentsData = await agentsRes.json();
        setAgents(agentsData || []);
      }
      if (insightsRes.ok) {
        const insightsData = await insightsRes.json();
        setInsights(insightsData || []);
      }
      if (summaryRes.ok) {
        const summaryData = await summaryRes.json();
        setSummary(summaryData);
      }
    } catch (error) {
      console.error("Failed to fetch initial data:", error);
    }
  }, [setTrace, setMessages, setAgents, setInsights]);

  // Connect to WebSocket for real-time updates
  const { isConnected } = useWebSocket(wsUrl, {
    onConnect: () => {
      setConnected(true);
      fetchData();
    },
    onDisconnect: () => setConnected(false),
    onMessage: (message) => {
      addMessage(message);
      // Update summary
      setSummary((prev) =>
        prev
          ? {
              ...prev,
              total_messages: prev.total_messages + 1,
              error_count:
                message.error || message.status_code >= 400
                  ? prev.error_count + 1
                  : prev.error_count,
            }
          : null
      );
    },
    onAgent: (agent) => addAgent(agent),
    onInsight: (insight) => {
      addInsight(insight);
      setSummary((prev) =>
        prev ? { ...prev, total_insights: prev.total_insights + 1 } : null
      );
    },
    onTraceStatus: (trace) => setTrace(trace),
  });

  // Fetch data on mount
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Handle export
  const handleExport = async () => {
    try {
      const baseUrl =
        typeof window !== "undefined"
          ? `${window.location.protocol}//${window.location.host}`
          : "http://localhost:8080";
      
      window.open(`${baseUrl}/api/export`, "_blank");
    } catch (error) {
      console.error("Failed to export:", error);
    }
  };

  // Handle clear
  const handleClear = () => {
    clearAll();
    setSummary({
      total_messages: 0,
      total_insights: 0,
      error_count: 0,
      success_count: 0,
      avg_duration_ms: 0,
      method_counts: {},
      agent_error_counts: {},
    });
  };

  const timelineItems = getTimelineItems();

  const tabs = [
    {
      id: "timeline" as const,
      label: "Timeline",
      icon: MessageSquare,
      count: messages.length,
    },
    {
      id: "agents" as const,
      label: "Agents",
      icon: Bot,
      count: agents.length,
    },
    {
      id: "insights" as const,
      label: "Insights",
      icon: Lightbulb,
      count: insights.length,
    },
  ];

  return (
    <div className="min-h-screen bg-zinc-950">
      <Header
        trace={trace}
        summary={summary}
        isConnected={isConnected}
        onExport={handleExport}
        onClear={handleClear}
      />

      <main className="max-w-7xl mx-auto px-4 py-6">
        {/* Tabs */}
        <div className="flex items-center gap-1 mb-6 border-b border-zinc-800">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors
                border-b-2 -mb-[2px]
                ${
                  activeTab === tab.id
                    ? "text-blue-400 border-blue-400"
                    : "text-zinc-500 border-transparent hover:text-zinc-300"
                }
              `}
            >
              <tab.icon className="w-4 h-4" />
              {tab.label}
              {tab.count > 0 && (
                <span
                  className={`px-1.5 py-0.5 rounded text-xs ${
                    activeTab === tab.id
                      ? "bg-blue-500/20 text-blue-400"
                      : "bg-zinc-800 text-zinc-500"
                  }`}
                >
                  {tab.count}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className={selectedMessageId ? "mr-[520px]" : ""}>
          {activeTab === "timeline" && <Timeline items={timelineItems} />}
          {activeTab === "agents" && <AgentList agents={agents} />}
          {activeTab === "insights" && <InsightsPanel insights={insights} />}
        </div>
      </main>

      {/* Message Inspector Sidebar */}
      <MessageInspector />

      {/* Empty state for new users */}
      {messages.length === 0 && !trace && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="fixed inset-0 flex items-center justify-center pointer-events-none"
        >
          <div className="text-center">
            <Activity className="w-16 h-16 text-zinc-700 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-zinc-400 mb-2">
              Waiting for connections...
            </h2>
            <p className="text-zinc-600 max-w-md">
              Start your A2A agent with the trace command to begin debugging:
            </p>
            <div className="mt-4 p-4 rounded-lg bg-zinc-900 border border-zinc-800 inline-block">
              <code className="text-sm text-emerald-400">
                a2a-trace -- node my-agent.js
              </code>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
}
