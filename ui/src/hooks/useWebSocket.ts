"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import type { Message, Agent, Insight, Trace, WebSocketMessage } from "@/lib/types";

interface UseWebSocketOptions {
  onMessage?: (message: Message) => void;
  onAgent?: (agent: Agent) => void;
  onInsight?: (insight: Insight) => void;
  onTraceStatus?: (trace: Trace) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
}

export function useWebSocket(url: string, options: UseWebSocketOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const optionsRef = useRef(options);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectAttemptsRef = useRef(0);

  // Keep options ref updated
  useEffect(() => {
    optionsRef.current = options;
  }, [options]);

  const connect = useCallback(() => {
    // Prevent multiple connections
    if (wsRef.current?.readyState === WebSocket.OPEN || 
        wsRef.current?.readyState === WebSocket.CONNECTING) {
      return;
    }

    try {
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
        optionsRef.current.onConnect?.();
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;
        optionsRef.current.onDisconnect?.();

        // Don't reconnect if it was a clean close
        if (event.code === 1000) {
          return;
        }

        // Reconnect with exponential backoff (max 30s)
        const attempts = reconnectAttemptsRef.current;
        const delay = Math.min(1000 * Math.pow(2, attempts), 30000);
        
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
        }
        
        reconnectTimeoutRef.current = setTimeout(() => {
          reconnectAttemptsRef.current += 1;
          connect();
        }, delay);
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      ws.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);

          switch (data.type) {
            case "message":
              optionsRef.current.onMessage?.(data.payload as Message);
              break;
            case "agent":
              optionsRef.current.onAgent?.(data.payload as Agent);
              break;
            case "insight":
              optionsRef.current.onInsight?.(data.payload as Insight);
              break;
            case "trace_status":
              optionsRef.current.onTraceStatus?.(data.payload as Trace);
              break;
            case "pong":
            case "connected":
              // Heartbeat/connection confirmation
              break;
          }
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error("Failed to connect WebSocket:", error);
    }
  }, [url]); // Only depend on url

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current !== null) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close(1000, "Client disconnect");
      wsRef.current = null;
    }
  }, []);

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    }
  }, []);

  // Connect on mount, disconnect on unmount
  useEffect(() => {
    connect();
    return () => disconnect();
  }, [url]); // Only reconnect if url changes

  return {
    isConnected,
    send,
    reconnect: connect,
  };
}

