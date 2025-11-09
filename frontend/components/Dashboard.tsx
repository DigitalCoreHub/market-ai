'use client';

import AgentsPerformanceChart, { type AgentHistory } from '@/components/AgentsPerformanceChart';
import BreakingNews from '@/components/BreakingNews';
import LatestNews from '@/components/LatestNews';
import Leaderboard from '@/components/Leaderboard';
import MarketDataSources from '@/components/MarketDataSources';
import ReasoningFeed from '@/components/ReasoningFeed';
import StockSentimentPanel from '@/components/StockSentimentPanel';
import { useWebSocket, type WebSocketMessage } from '@/lib/websocket';
import { Activity, Moon, Sun, TrendingUp, Wallet, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

interface Agent {
  id: string;
  name: string;
  model: string;
  current_balance: number;
  status: string;
  profit_loss?: number;
  roi?: number;
}

export default function Dashboard() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [agentHistory, setAgentHistory] = useState<AgentHistory[]>([]);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const [darkMode, setDarkMode] = useState(true); // Default to true, will sync in useEffect
  const agentHistoryRef = useRef<AgentHistory[]>([]);
  const lastSnapshotTimeRef = useRef<number>(0);
  const handleWsMessage = (message: WebSocketMessage) => {
    setLastMessage(message);

    if (message.type === 'price_update') {
      // Record agent balance snapshots on price update
      if (agents.length > 0) {
        const now = Date.now();
        if (now - lastSnapshotTimeRef.current >= 1000) {
          lastSnapshotTimeRef.current = now;
          const snapshot: AgentHistory = { timestamp: now };
          agents.forEach(agent => {
            snapshot[agent.id] = agent.current_balance;
          });

          agentHistoryRef.current = [...agentHistoryRef.current, snapshot];
          if (agentHistoryRef.current.length > 1440) {
            agentHistoryRef.current = agentHistoryRef.current.slice(agentHistoryRef.current.length - 1440);
          }
          setAgentHistory(agentHistoryRef.current);
        }
      }
    }
  };
  const { isConnected } = useWebSocket('ws://localhost:8080/ws', { onMessage: handleWsMessage });

  useEffect(() => {
    fetch('http://localhost:8080/api/v1/agents')
      .then(res => res.json())
      .then(data => setAgents(data.data || []));
  }, []);

  // Initialize ref with agents when they change
  useEffect(() => {
    agentHistoryRef.current = agentHistory;
  }, [agentHistory]);

  // Load dark mode from localStorage only on client
  // eslint-disable react-hooks/exhaustive-deps
  useEffect(() => {
    const stored = localStorage.getItem('darkMode');
    setDarkMode(stored ? JSON.parse(stored) : true);
  }, []);

  useEffect(() => {
    // Update DOM and localStorage when darkMode changes
    const html = document.documentElement;
    if (darkMode) {
      html.classList.add('dark');
      html.classList.remove('light');
    } else {
      html.classList.add('light');
      html.classList.remove('dark');
    }
    localStorage.setItem('darkMode', JSON.stringify(darkMode));
  }, [darkMode]);

  return (
  <div className="min-h-screen bg-linear-to-br from-gray-50 via-white to-gray-100 dark:from-gray-950 dark:via-gray-900 dark:to-black transition-colors duration-300">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 backdrop-blur-xl sticky top-0 z-50">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="bg-linear-to-br from-blue-500 to-purple-600 p-2.5 rounded-xl shadow-lg shadow-blue-500/20">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold bg-linear-to-r from-gray-900 to-gray-600 dark:from-white dark:to-gray-400 bg-clip-text text-transparent">
                  Market AI
                </h1>
                <p className="text-sm text-gray-500 dark:text-gray-400">Yapay Zekâ Ticaret Arenası</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <div className={`flex items-center gap-2 px-4 py-2 rounded-full border transition-all duration-300 ${
                isConnected
                  ? 'bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-900 text-green-700 dark:text-green-400'
                  : 'bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-900 text-red-700 dark:text-red-400'
              }`}>
                <Activity className={`w-4 h-4 ${isConnected ? 'animate-pulse' : ''}`} />
                <span className="text-sm font-medium">{isConnected ? 'Canlı' : 'Çevrimdışı'}</span>
              </div>

              <button
                onClick={() => setDarkMode(!darkMode)}
                className="p-2.5 rounded-xl bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-all duration-300 border border-gray-200 dark:border-gray-700"
              >
                {darkMode ? (
                  <Sun className="w-5 h-5 text-gray-600 dark:text-gray-300" />
                ) : (
                  <Moon className="w-5 h-5 text-gray-600 dark:text-gray-300" />
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-6 py-8 space-y-8">
        {/* Summary Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">Aktif Ajanlar</p>
                <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">{agents.length}</p>
              </div>
              <div className="bg-blue-100 dark:bg-blue-950/30 p-3 rounded-xl">
                <Zap className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">Trading Dönemi</p>
                <p className="text-lg font-bold text-gray-900 dark:text-white mt-2">Canlı</p>
              </div>
              <div className="bg-green-100 dark:bg-green-950/30 p-3 rounded-xl">
                <Activity className="w-6 h-6 text-green-600 dark:text-green-400" />
              </div>
            </div>
          </div>

          {agents.length > 0 && (
            <>
              <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">En İyi Ajan</p>
                    <p className="text-lg font-bold text-gray-900 dark:text-white mt-2">
                      {agents.reduce((prev, curr) => (curr.current_balance > prev.current_balance) ? curr : prev).name}
                    </p>
                  </div>
                  <div className="bg-amber-100 dark:bg-amber-950/30 p-3 rounded-xl">
                    <TrendingUp className="w-6 h-6 text-amber-600 dark:text-amber-400" />
                  </div>
                </div>
              </div>

              <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-500 dark:text-gray-400 font-medium">En Yüksek Bakiye</p>
                    <p className="text-lg font-bold text-gray-900 dark:text-white mt-2">
                      ₺{Math.max(...agents.map(a => a.current_balance)).toLocaleString('tr-TR', { maximumFractionDigits: 0 })}
                    </p>
                  </div>
                  <div className="bg-cyan-100 dark:bg-cyan-950/30 p-3 rounded-xl">
                    <Wallet className="w-6 h-6 text-cyan-600 dark:text-cyan-400" />
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Leaderboard */}
        <Leaderboard />

        {/* Performance Chart */}
        {agentHistory.length > 0 && (
          <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
            <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">Ajanların Performansı</h2>
            <AgentsPerformanceChart data={agentHistory} agents={agents} height={400} />
          </div>
        )}

        {/* Market Data Sources + Sentiment (v0.5) */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <MarketDataSources symbols={["THYAO","AKBNK","ASELS","GARAN"]} />
          <StockSentimentPanel symbols={["THYAO","AKBNK","ASELS","GARAN"]} />
        </div>

        {/* Breaking news (scraper) */}
        <BreakingNews symbols={["THYAO","AKBNK","ASELS","GARAN"]} />

        {/* Active AI Agents */}
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="bg-linear-to-br from-cyan-500 to-blue-600 p-2 rounded-lg">
              <Wallet className="w-5 h-5 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">Aktif Yapay Zekâ Ajanlar</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">Otonom ticaret algoritmaları</p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {agents.map((agent) => (
              <div
                key={agent.id}
                className="group relative bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6 hover:shadow-xl hover:shadow-cyan-500/10 dark:hover:shadow-cyan-500/5 transition-all duration-300 hover:-translate-y-1"
              >
                <div className="absolute inset-0 bg-linear-to-br from-cyan-500/5 to-blue-500/5 dark:from-cyan-500/10 dark:to-blue-500/10 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                <div className="relative flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="bg-linear-to-br from-cyan-500 to-blue-600 p-2 rounded-lg">
                        <Zap className="w-4 h-4 text-white" />
                      </div>
                      <div>
                        <div className="font-bold text-lg text-gray-900 dark:text-white">{agent.name}</div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">{agent.model}</div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 mt-4">
                      <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold ${
                        agent.status === 'active'
                          ? 'bg-green-100 dark:bg-green-950/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-900'
                          : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 border border-gray-200 dark:border-gray-700'
                      }`}>
                        <span className={`w-1.5 h-1.5 rounded-full ${
                          agent.status === 'active' ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
                        }`} />
                        {agent.status === 'active' ? 'Aktif' : 'Pasif'}
                      </span>
                    </div>
                  </div>

                  <div className="text-right">
                    <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Bakiye</div>
                    <div className="text-2xl font-bold bg-linear-to-r from-cyan-600 to-blue-600 dark:from-cyan-400 dark:to-blue-400 bg-clip-text text-transparent">
                      ₺{agent.current_balance.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* AI Reasoning & News Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ReasoningFeed lastMessage={lastMessage} />
          <LatestNews lastMessage={lastMessage} />
        </div>
      </div>
    </div>
  );
}
