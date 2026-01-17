"use client";

import { motion } from "framer-motion";
import { Bot, Server, Globe } from "lucide-react";
import type { Agent } from "@/lib/types";

interface AgentNodeProps {
  agent: Agent;
  index: number;
}

export function AgentNode({ agent, index }: AgentNodeProps) {
  const skills = agent.skills ? JSON.parse(agent.skills) : [];

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.2, delay: index * 0.1 }}
      className="p-4 rounded-lg border border-zinc-800 bg-zinc-900/50 hover:border-zinc-700 transition-colors"
    >
      <div className="flex items-start gap-3">
        <div className="p-2 rounded-lg bg-blue-500/10">
          <Bot className="w-5 h-5 text-blue-400" />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-zinc-200 truncate">
            {agent.name || "Unknown Agent"}
          </h3>
          <p className="text-sm text-zinc-500 truncate">{agent.url}</p>
          
          {agent.description && (
            <p className="text-sm text-zinc-400 mt-2 line-clamp-2">
              {agent.description}
            </p>
          )}

          {agent.version && (
            <div className="flex items-center gap-2 mt-2">
              <span className="text-xs text-zinc-600">Version:</span>
              <span className="text-xs font-mono text-zinc-400">{agent.version}</span>
            </div>
          )}

          {skills.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-3">
              {skills.slice(0, 3).map((skill: { name: string }, i: number) => (
                <span
                  key={i}
                  className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-400"
                >
                  {skill.name}
                </span>
              ))}
              {skills.length > 3 && (
                <span className="text-xs text-zinc-600">
                  +{skills.length - 3} more
                </span>
              )}
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
}

export function AgentList({ agents }: { agents: Agent[] }) {
  if (agents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-32 text-zinc-500">
        <Globe className="w-8 h-8 mb-2 opacity-50" />
        <p className="text-sm">No agents discovered yet</p>
        <p className="text-xs mt-1">Agents are detected from /.well-known/agent.json</p>
      </div>
    );
  }

  return (
    <div className="grid gap-3 grid-cols-1 md:grid-cols-2">
      {agents.map((agent, index) => (
        <AgentNode key={agent.id} agent={agent} index={index} />
      ))}
    </div>
  );
}

