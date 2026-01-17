// A2A Trace Types

export interface Trace {
  id: string;
  started_at: string;
  command: string;
  status: "running" | "completed" | "error";
}

export interface Message {
  id: string;
  trace_id: string;
  timestamp: string;
  direction: "request" | "response";
  from_agent: string;
  to_agent: string;
  method: string;
  url: string;
  headers: string;
  body: string;
  duration_ms: number;
  status_code: number;
  error: string;
  request_id: string;
  content_type: string;
  size: number;
}

export interface Agent {
  id: string;
  url: string;
  name: string;
  description: string;
  version: string;
  skills: string;
  first_seen: string;
}

export interface Insight {
  id: string;
  trace_id: string;
  message_id: string;
  type: "error" | "warning" | "info";
  category: string;
  title: string;
  details: string;
  timestamp: string;
}

export interface Summary {
  total_messages: number;
  total_insights: number;
  error_count: number;
  success_count: number;
  avg_duration_ms: number;
  method_counts: Record<string, number>;
  agent_error_counts: Record<string, number>;
}

export interface WebSocketMessage {
  type: "message" | "agent" | "insight" | "trace_status";
  payload: Message | Agent | Insight | Trace;
}

// Parsed versions of JSON fields
export interface ParsedMessage extends Omit<Message, "headers" | "body"> {
  headers: Record<string, string>;
  body: unknown;
  parsedBody?: A2ARequest | A2AResponse;
}

export interface A2ARequest {
  jsonrpc: string;
  method: string;
  id: string | number | null;
  params?: unknown;
}

export interface A2AResponse {
  jsonrpc: string;
  id: string | number | null;
  result?: unknown;
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
}

// UI State
export interface TimelineItem {
  id: string;
  timestamp: Date;
  request?: ParsedMessage;
  response?: ParsedMessage;
  fromAgent: string;
  toAgent: string;
  method: string;
  status: "pending" | "success" | "error" | "slow";
  duration?: number;
}

