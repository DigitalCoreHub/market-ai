'use client';

import { useWebSocket } from '@/lib/websocket';
import { Medal, TrendingDown, TrendingUp, Trophy } from 'lucide-react';
import { useEffect, useState } from 'react';

interface LeaderboardEntry {
  rank: number;
  agent_id: string;
  agent_name: string;
  model: string;
  roi: number;
  profit_loss: number;
  win_rate: number;
  total_trades: number;
  balance: number;
  portfolio_value: number;
  total_value: number;
  badges: string[];
}

export default function Leaderboard() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [roiHistory, setRoiHistory] = useState<Record<string, { time: string; roi: number }[]>>({});
  const [lastHistoryFetch, setLastHistoryFetch] = useState<number>(0);
  // Fetch ROI history for all agents (typed)
  const fetchROIHistory = () => {
    fetch('http://localhost:8080/api/v1/leaderboard/roi-history?limit=120')
      .then(r => r.json())
      .then((d: { success?: boolean; data?: unknown }) => {
        if (d.success && d.data && typeof d.data === 'object' && d.data !== null) {
          type RawPoint = { time: string; roi: number };
          const transformed: Record<string, RawPoint[]> = {};
          Object.entries(d.data as Record<string, unknown>).forEach(([agentId, points]) => {
            if (Array.isArray(points)) {
              const validPoints = points
                .filter((p): p is RawPoint =>
                  p &&
                  typeof p === 'object' &&
                  typeof (p as RawPoint).time === 'string' &&
                  typeof (p as RawPoint).roi === 'number'
                )
                .map(p => ({ time: p.time, roi: p.roi }))
                .sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime());
              transformed[agentId] = validPoints;
            }
          });
          setRoiHistory(transformed);
          setLastHistoryFetch(Date.now());
        }
      })
      .catch(() => {});
  };
  useWebSocket('ws://localhost:8080/ws', {
    onMessage: (msg) => {
      if (msg.type === 'leaderboard_updated') {
        const data = msg.data;
        if (Array.isArray(data)) {
          const entries = data.filter((entry): entry is LeaderboardEntry =>
            entry &&
            typeof entry === 'object' &&
            typeof entry.rank === 'number' &&
            typeof entry.agent_id === 'string' &&
            typeof entry.agent_name === 'string' &&
            typeof entry.model === 'string' &&
            typeof entry.roi === 'number' &&
            typeof entry.profit_loss === 'number' &&
            typeof entry.win_rate === 'number' &&
            typeof entry.total_trades === 'number' &&
            typeof entry.balance === 'number' &&
            typeof entry.portfolio_value === 'number' &&
            typeof entry.total_value === 'number'
          );
          setEntries(entries);
          // After a leaderboard update, refresh ROI history if stale (>30s)
          const now = Date.now();
          if (now - lastHistoryFetch > 30000) {
            fetchROIHistory();
          }
        }
      }
    },
  });

  useEffect(() => {
    fetch('http://localhost:8080/api/v1/leaderboard')
      .then(r => r.json())
      .then((d: { success?: boolean; data?: unknown }) => {
        if (d.data && Array.isArray(d.data)) {
          const entries = d.data.filter((entry): entry is LeaderboardEntry =>
            entry &&
            typeof entry === 'object' &&
            typeof entry.rank === 'number' &&
            typeof entry.agent_id === 'string' &&
            typeof entry.agent_name === 'string' &&
            typeof entry.model === 'string' &&
            typeof entry.roi === 'number' &&
            typeof entry.profit_loss === 'number' &&
            typeof entry.win_rate === 'number' &&
            typeof entry.total_trades === 'number' &&
            typeof entry.balance === 'number' &&
            typeof entry.portfolio_value === 'number' &&
            typeof entry.total_value === 'number'
          );
          setEntries(entries);
        }
      })
      .catch(() => {});
    fetchROIHistory();
  }, []);

  // Sparkline generator (inline SVG) for ROI trend
  const Sparkline: React.FC<{ points: { time: string; roi: number }[]; positive: boolean }> = ({ points, positive }) => {
    if (!points || points.length < 2) {
      return <div className="h-6 flex items-center text-[10px] text-gray-400">-</div>;
    }
    const width = 120;
    const height = 30;
    const padding = 2;
    const rois = points.map(p => p.roi);
    const min = Math.min(...rois);
    const max = Math.max(...rois);
    const range = max - min || 1;
    const stepX = (width - padding * 2) / (points.length - 1);
    const d = points.map((p, i) => {
      const x = padding + i * stepX;
      const yNorm = (p.roi - min) / range; // 0..1
      const y = height - padding - yNorm * (height - padding * 2);
      return `${x},${y}`;
    }).join(' ');
    const last = points[points.length - 1].roi;
    const first = points[0].roi;
    const trendUp = last >= first;
    return (
      <svg viewBox={`0 0 ${width} ${height}`} className="w-[120px] h-[30px]">
        <polyline
          fill="none"
          stroke={trendUp ? (positive ? '#16a34a' : '#0ea5e9') : '#dc2626'}
          strokeWidth={1.5}
          points={d}
        />
        {/* Optional area fill */}
        <polyline
          fill={trendUp ? (positive ? 'rgba(22,163,74,0.15)' : 'rgba(14,165,233,0.15)') : 'rgba(220,38,38,0.15)'}
          stroke="none"
          points={`${d} ${width - padding},${height} ${padding},${height}`}
        />
      </svg>
    );
  };

  const rankIcon = (rank: number) => {
    switch (rank) {
      case 1: return <Trophy className="w-6 h-6 text-yellow-500" />;
      case 2: return <Medal className="w-6 h-6 text-gray-400" />;
      case 3: return <Medal className="w-6 h-6 text-amber-600" />;
      default: return <span className="text-sm font-semibold text-gray-500">#{rank}</span>;
    }
  };

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Trophy className="w-5 h-5 text-yellow-500" />
        <h2 className="text-xl font-bold text-gray-900 dark:text-white">AI Arena Liderlik Tablosu</h2>
      </div>
      <div className="space-y-3">
        {entries.map(e => (
          <div key={e.agent_id} className="p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md transition">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {rankIcon(e.rank)}
                <div>
                  <div className="font-semibold text-gray-900 dark:text-white">{e.agent_name}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">{e.model}</div>
                </div>
              </div>
              <div className={`flex items-center text-sm font-semibold px-3 py-1 rounded-full ${e.roi >= 0 ? 'bg-green-100 text-green-700 dark:bg-green-950/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-950/30 dark:text-red-400'}`}>
                {e.roi >= 0 ? <TrendingUp className="w-4 h-4 mr-1" /> : <TrendingDown className="w-4 h-4 mr-1" />}
                {e.roi.toFixed(2)}%
              </div>
            </div>
            <div className="grid grid-cols-5 gap-4 mt-3 text-sm items-center">
              <div>
                <div className="text-xs text-gray-500 dark:text-gray-400">P/L</div>
                <div className={e.profit_loss >= 0 ? 'text-green-600 dark:text-green-400 font-medium' : 'text-red-600 dark:text-red-400 font-medium'}>
                  {e.profit_loss >= 0 ? '+' : ''}₺{e.profit_loss.toFixed(0)}
                </div>
              </div>
              <div>
                <div className="text-xs text-gray-500 dark:text-gray-400">Kazanç Oranı</div>
                <div className="font-medium">{e.win_rate.toFixed(1)}%</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 dark:text-gray-400">İşlem</div>
                <div className="font-medium">{e.total_trades}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 dark:text-gray-400">Toplam Değer</div>
                <div className="font-medium">₺{(e.total_value/1000).toFixed(1)}K</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 dark:text-gray-400">ROI Trend</div>
                <div className="mt-1">
                  <Sparkline points={roiHistory[e.agent_id] || []} positive={e.roi >= 0} />
                </div>
              </div>
            </div>
            {e.badges && e.badges.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-3">
                {e.badges.map((b,i) => (
                  <span key={i} className="text-[10px] px-2 py-1 rounded bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700">{b}</span>
                ))}
              </div>
            )}
          </div>
        ))}
        {entries.length === 0 && (
          <div className="text-center text-sm text-gray-500 dark:text-gray-400 py-6">Henüz veri yok. Ajanlar başlatılıyor...</div>
        )}
      </div>
    </div>
  );
}
