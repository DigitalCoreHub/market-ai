'use client';

import { format } from 'date-fns';
import { tr } from 'date-fns/locale';
import { useState } from 'react';
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

export interface AgentHistory {
  timestamp: number;
  [agentId: string]: number; // agent id -> balance
}

interface AgentsPerformanceChartProps {
  data: AgentHistory[];
  agents: { id: string; name: string }[];
  height?: number;
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4'];

export default function AgentsPerformanceChart({
  data,
  agents,
  height = 400,
}: AgentsPerformanceChartProps) {
  const [selectedDate, setSelectedDate] = useState<string | null>(null);

  // Group data by date
  const dataByDate = new Map<string, AgentHistory[]>();
  data.forEach((item) => {
    const dateStr = format(new Date(item.timestamp), 'yyyy-MM-dd');
    if (!dataByDate.has(dateStr)) {
      dataByDate.set(dateStr, []);
    }
    dataByDate.get(dateStr)!.push(item);
  });

  const uniqueDates = Array.from(dataByDate.keys()).sort();

  // Get selected day's snapshots or latest
  const selectedDaySnapshots = selectedDate
    ? dataByDate.get(selectedDate) || []
    : data; // Show all if no selection

  const selectedSnapshot = selectedDaySnapshots[selectedDaySnapshots.length - 1];
  if (!data.length) {
    return (
      <div className="flex items-center justify-center h-96 bg-gray-50 dark:bg-gray-800 rounded-2xl border border-gray-200 dark:border-gray-700">
        <p className="text-gray-500 dark:text-gray-400">Veri yükleniyor...</p>
      </div>
    );
  }

  return (
    <div className="w-full space-y-6">
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={data}
          margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
        >
          <CartesianGrid stroke="rgba(148, 163, 184, 0.2)" strokeDasharray="3 3" />
          <XAxis
            dataKey="timestamp"
            tick={{ fontSize: 11 }}
            tickFormatter={(ts) => format(new Date(ts), 'HH:mm')}
          />
          <YAxis
            tick={{ fontSize: 11 }}
            tickFormatter={(v) => `₺${(Number(v) / 1000).toFixed(0)}K`}
            domain={['dataMin - 5000', 'dataMax + 5000']}
          />
          <Tooltip
            contentStyle={{ background: 'rgba(17, 24, 39, 0.8)', border: '1px solid #374151', borderRadius: 12 }}
            labelClassName="text-gray-100"
            itemStyle={{ color: '#e5e7eb' }}
            labelFormatter={(ts) => format(new Date(Number(ts)), 'HH:mm:ss')}
            formatter={(value: number | string) => `₺${Number(value).toLocaleString('tr-TR', { maximumFractionDigits: 0 })}`}
          />
          <Legend wrapperStyle={{ paddingTop: '20px' }} />
          {agents.map((agent, idx) => (
            <Line
              key={agent.id}
              type="monotone"
              dataKey={agent.id}
              stroke={COLORS[idx % COLORS.length]}
              dot={false}
              strokeWidth={2}
              isAnimationActive={false}
              name={agent.name}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>

      {/* Date Selection Cards */}
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Günler</h3>
        <div className="grid grid-cols-7 gap-2">
          {uniqueDates.slice(-7).map((date) => {
            const isSelected = selectedDate === date;
            const dateObj = new Date(date + 'T00:00:00');
            return (
              <button
                key={date}
                onClick={() => setSelectedDate(isSelected ? null : date)}
                className={`p-3 rounded-lg text-center transition-all ${
                  isSelected
                    ? 'bg-blue-500 text-white shadow-lg'
                    : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600'
                }`}
              >
                <div className="text-xs font-semibold">{format(dateObj, 'EEE', { locale: tr })}</div>
                <div className="text-sm font-bold">{format(dateObj, 'd')}</div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Snapshot Info Card */}
      {selectedSnapshot && (
        <div className="bg-linear-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 border border-blue-200 dark:border-blue-800 rounded-2xl p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
              {selectedDate
                ? format(new Date(selectedDate + 'T00:00:00'), 'dd MMMM yyyy')
                : 'Son Güncelleme'
              } tarihinde bakiyeler
            </h3>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {agents.map((agent, idx) => (
              <div key={agent.id} className="bg-white dark:bg-gray-800 rounded-lg p-3 border border-gray-100 dark:border-gray-700">
                <div className="flex items-center gap-2 mb-1">
                  <div
                    className="w-2 h-2 rounded-full"
                    style={{ backgroundColor: COLORS[idx % COLORS.length] }}
                  />
                  <p className="text-xs font-medium text-gray-600 dark:text-gray-300">{agent.name}</p>
                </div>
                <p className="text-lg font-bold text-gray-900 dark:text-white">
                  ₺{(selectedSnapshot[agent.id] || 0).toLocaleString('tr-TR', { maximumFractionDigits: 0 })}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
