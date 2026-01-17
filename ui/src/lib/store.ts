"use client";

import { create } from "zustand";
import type { Message, Agent, Insight, Trace, ParsedMessage, TimelineItem } from "./types";

interface TraceStore {
  // Data
  trace: Trace | null;
  messages: Message[];
  agents: Agent[];
  insights: Insight[];
  
  // UI State
  selectedMessageId: string | null;
  isConnected: boolean;
  
  // Actions
  setTrace: (trace: Trace | null) => void;
  addMessage: (message: Message) => void;
  setMessages: (messages: Message[]) => void;
  addAgent: (agent: Agent) => void;
  setAgents: (agents: Agent[]) => void;
  addInsight: (insight: Insight) => void;
  setInsights: (insights: Insight[]) => void;
  selectMessage: (id: string | null) => void;
  setConnected: (connected: boolean) => void;
  clearAll: () => void;
  
  // Computed
  getTimelineItems: () => TimelineItem[];
  getSelectedMessage: () => ParsedMessage | null;
  getAgentByUrl: (url: string) => Agent | undefined;
}

function parseMessage(msg: Message): ParsedMessage {
  let headers: Record<string, string> = {};
  let body: unknown = msg.body;
  let parsedBody: ParsedMessage["parsedBody"] = undefined;

  try {
    if (msg.headers) {
      headers = JSON.parse(msg.headers);
    }
  } catch {
    headers = {};
  }

  try {
    if (msg.body) {
      body = JSON.parse(msg.body);
      parsedBody = body as ParsedMessage["parsedBody"];
    }
  } catch {
    body = msg.body;
  }

  return {
    ...msg,
    headers,
    body,
    parsedBody,
  };
}

export const useTraceStore = create<TraceStore>((set, get) => ({
  // Initial state
  trace: null,
  messages: [],
  agents: [],
  insights: [],
  selectedMessageId: null,
  isConnected: false,

  // Actions
  setTrace: (trace) => set({ trace }),
  
  addMessage: (message) =>
    set((state) => ({
      messages: [...state.messages, message],
    })),
    
  setMessages: (messages) => set({ messages }),
  
  addAgent: (agent) =>
    set((state) => {
      const exists = state.agents.some((a) => a.url === agent.url);
      if (exists) {
        return {
          agents: state.agents.map((a) =>
            a.url === agent.url ? agent : a
          ),
        };
      }
      return { agents: [...state.agents, agent] };
    }),
    
  setAgents: (agents) => set({ agents }),
  
  addInsight: (insight) =>
    set((state) => ({
      insights: [...state.insights, insight],
    })),
    
  setInsights: (insights) => set({ insights }),
  
  selectMessage: (id) => set({ selectedMessageId: id }),
  
  setConnected: (connected) => set({ isConnected: connected }),
  
  clearAll: () =>
    set({
      trace: null,
      messages: [],
      agents: [],
      insights: [],
      selectedMessageId: null,
    }),

  // Computed
  getTimelineItems: () => {
    const { messages, agents } = get();
    const items: TimelineItem[] = [];
    const requestMap = new Map<string, Message>();

    // Group requests and responses
    for (const msg of messages) {
      if (msg.direction === "request") {
        requestMap.set(msg.id, msg);
      }
    }

    for (const msg of messages) {
      if (msg.direction === "request") {
        const response = messages.find(
          (m) => m.direction === "response" && m.request_id === msg.id
        );

        const fromAgent = agents.find((a) => a.url.includes(msg.from_agent));
        const toAgent = agents.find((a) => a.url.includes(msg.to_agent));

        let status: TimelineItem["status"] = "pending";
        if (response) {
          if (response.error || response.status_code >= 400) {
            status = "error";
          } else if (response.duration_ms > 1000) {
            status = "slow";
          } else {
            status = "success";
          }
        }

        items.push({
          id: msg.id,
          timestamp: new Date(msg.timestamp),
          request: parseMessage(msg),
          response: response ? parseMessage(response) : undefined,
          fromAgent: fromAgent?.name || msg.from_agent || "Client",
          toAgent: toAgent?.name || msg.to_agent || "Agent",
          method: msg.method || "HTTP",
          status,
          duration: response?.duration_ms,
        });
      }
    }

    return items.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());
  },

  getSelectedMessage: () => {
    const { messages, selectedMessageId } = get();
    const msg = messages.find((m) => m.id === selectedMessageId);
    return msg ? parseMessage(msg) : null;
  },

  getAgentByUrl: (url) => {
    const { agents } = get();
    return agents.find((a) => url.includes(a.url) || a.url.includes(url));
  },
}));

