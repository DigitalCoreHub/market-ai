'use client';

import { formatDistanceToNow } from 'date-fns';
import { tr } from 'date-fns/locale';
import { Brain, Minus, TrendingDown, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';

interface Decision {
  agent_id: string;
  agent_name: string;
  decision_id: string;
  action: 'BUY' | 'SELL' | 'HOLD';
  stock_symbol: string;
  quantity: number;
  reasoning_summary: string;
  confidence: number;
  risk_level: 'low' | 'medium' | 'high';
  thinking_steps: Array<{ step: string; observation: string }>;
  timestamp: number;
}

interface WebSocketMessage {
  type: string;
  data: Decision | Record<string, unknown>;
  timestamp: number;
}

export default function ReasoningFeed({ lastMessage }: { lastMessage: WebSocketMessage | null }) {
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [thinkingAgents, setThinkingAgents] = useState<Set<string>>(new Set());

  // eslint-disable react-hooks/exhaustive-deps
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.type === 'agent_thinking') {
      const agentId = lastMessage.data.agent_id;
      setThinkingAgents(prev => new Set(prev).add(agentId));
    }

    if (lastMessage.type === 'agent_decision') {
      const decision: Decision = {
        agent_id: lastMessage.data.agent_id,
        agent_name: lastMessage.data.agent_name,
        decision_id: lastMessage.data.decision_id,
        action: lastMessage.data.action,
        stock_symbol: lastMessage.data.stock_symbol,
        quantity: lastMessage.data.quantity,
        reasoning_summary: lastMessage.data.reasoning_summary,
        confidence: lastMessage.data.confidence,
        risk_level: lastMessage.data.risk_level,
        thinking_steps: lastMessage.data.thinking_steps || [],
        timestamp: lastMessage.data.timestamp || lastMessage.timestamp,
      };

      setDecisions(prev => [decision, ...prev].slice(0, 10));

      // Remove from thinking agents
      setThinkingAgents(prev => {
        const next = new Set(prev);
        next.delete(decision.agent_id);
        return next;
      });
    }
  }, [lastMessage]);

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'BUY':
        return <TrendingUp className="w-4 h-4 text-green-600 dark:text-green-400" />;
      case 'SELL':
        return <TrendingDown className="w-4 h-4 text-red-600 dark:text-red-400" />;
      default:
        return <Minus className="w-4 h-4 text-gray-600 dark:text-gray-400" />;
    }
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case 'BUY':
        return 'bg-green-50 dark:bg-green-950/30 text-green-900 dark:text-green-100 border-green-200 dark:border-green-900';
      case 'SELL':
        return 'bg-red-50 dark:bg-red-950/30 text-red-900 dark:text-red-100 border-red-200 dark:border-red-900';
      default:
        return 'bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 border-gray-200 dark:border-gray-700';
    }
  };

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case 'low':
        return 'bg-green-100 dark:bg-green-950/50 text-green-800 dark:text-green-200';
      case 'medium':
        return 'bg-yellow-100 dark:bg-yellow-950/50 text-yellow-800 dark:text-yellow-200';
      case 'high':
        return 'bg-red-100 dark:bg-red-950/50 text-red-800 dark:text-red-200';
      default:
        return 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200';
    }
  };

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="bg-linear-to-br from-purple-500 to-pink-600 p-2 rounded-lg">
          <Brain className="w-5 h-5 text-white" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">AI Akıl Yürütme Beslemesi</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">Canlı ajan kararları ve düşünce süreci</p>
        </div>
      </div>

      <div className="space-y-4 max-h-[600px] overflow-y-auto">
        {/* Thinking agents */}
        {Array.from(thinkingAgents).map(agentId => (
          <div key={agentId} className="p-4 border border-blue-200 dark:border-blue-900 rounded-lg bg-blue-50 dark:bg-blue-950/30 animate-pulse">
            <div className="flex items-center gap-2">
              <Brain className="w-5 h-5 text-blue-600 dark:text-blue-400 animate-spin" />
              <span className="font-semibold text-blue-900 dark:text-blue-100">Düşünüyor...</span>
            </div>
          </div>
        ))}

        {/* Decisions */}
        {decisions.map((decision) => (
          <div
            key={decision.decision_id}
            className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:shadow-md dark:hover:shadow-black/50 transition-shadow"
          >
            {/* Header */}
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-3">
                <span className="font-semibold text-lg text-gray-900 dark:text-white">{decision.agent_name}</span>
                <div className={`inline-flex items-center gap-2 px-3 py-1 rounded-full border text-sm font-semibold ${getActionColor(decision.action)}`}>
                  {getActionIcon(decision.action)}
                  {decision.action}
                </div>
              </div>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {formatDistanceToNow(new Date(decision.timestamp * 1000), { addSuffix: true, locale: tr })}
              </span>
            </div>

            {/* Decision details */}
            {decision.action !== 'HOLD' && (
              <div className="mb-3 text-sm text-gray-700 dark:text-gray-300 font-medium">
                <strong>{decision.stock_symbol}</strong> × {decision.quantity} lots
              </div>
            )}

            {/* Reasoning */}
            <p className="text-sm text-gray-800 dark:text-gray-200 mb-3 italic border-l-2 border-purple-300 dark:border-purple-700 pl-3">
              &quot;{decision.reasoning_summary}&quot;
            </p>

            {/* Thinking steps */}
            {decision.thinking_steps && decision.thinking_steps.length > 0 && (
              <details className="text-sm mb-3">
                <summary className="cursor-pointer text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 font-medium">
                  📊 Detaylı analiz göster ({decision.thinking_steps.length} adım)
                </summary>
                <div className="mt-3 space-y-2 pl-4 border-l-2 border-purple-200 dark:border-purple-800 py-2">
                  {decision.thinking_steps.map((step, idx) => (
                    <div key={idx}>
                      <div className="font-semibold text-gray-700 dark:text-gray-300">{step.step}</div>
                      <div className="text-gray-600 dark:text-gray-400 text-xs">{step.observation}</div>
                    </div>
                  ))}
                </div>
              </details>
            )}

            {/* Metrics */}
            <div className="flex flex-wrap gap-3 pt-3 border-t border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-600 dark:text-gray-400">Güven:</span>
                <div className="flex items-center gap-1">
                  <div className="w-20 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-linear-to-r from-blue-500 to-purple-600"
                      style={{ width: `${decision.confidence}%` }}
                    />
                  </div>
                  <span className="font-semibold text-xs text-gray-900 dark:text-white">{decision.confidence.toFixed(0)}%</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-600 dark:text-gray-400">Risk:</span>
                <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${getRiskColor(decision.risk_level)}`}>
                  {decision.risk_level === 'low' ? 'Düşük' : decision.risk_level === 'medium' ? 'Orta' : 'Yüksek'}
                </span>
              </div>
            </div>
          </div>
        ))}

        {decisions.length === 0 && thinkingAgents.size === 0 && (
          <div className="text-center text-gray-500 dark:text-gray-400 py-8">
            <Brain className="w-8 h-8 mx-auto mb-2 opacity-50" />
            Henüz AI kararı yok. Ajanlar bekleniyor...
          </div>
        )}
      </div>
    </div>
  );
}
