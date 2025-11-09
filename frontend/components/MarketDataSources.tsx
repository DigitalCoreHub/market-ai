"use client";

import { useMarketContext } from '@/lib/marketContext';
import { Activity, Clock, Globe, Twitter, Waves } from 'lucide-react';
import { useMemo } from 'react';

export default function MarketDataSources({ symbols }: { symbols: string[] }) {
  const { data, loading, error, refresh } = useMarketContext(symbols, 120000);

  const durations = data?.fetch_durations || {};
  const updatedAt = data?.updated_at ? new Date(data.updated_at) : null;

  const cards = useMemo(() => ([
    { key: 'yahoo', label: 'Yahoo Fiyat', icon: Waves, color: 'text-emerald-500' },
    { key: 'scraper', label: 'Web Scraper', icon: Globe, color: 'text-blue-500' },
    { key: 'twitter', label: 'Twitter', icon: Twitter, color: 'text-sky-500' },
  ]), []);

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-cyan-500" />
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">Veri Kaynakları</h2>
        </div>
        <button
          onClick={refresh}
          className="text-xs px-3 py-1 rounded-full border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
        >Yenile</button>
      </div>

      {error && (
        <div className="text-sm text-red-600 dark:text-red-400 mb-3">{error}</div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {cards.map(c => {
          const Icon = c.icon;
          const rawVal = (durations as Record<string, number | string | undefined>)[c.key] ?? 0;
          // Backend sends time.Duration as nanoseconds (number). Convert to ms.
          const num = typeof rawVal === 'number' ? rawVal : Number(rawVal);
          const ms = Number.isFinite(num) ? num / 1e6 : 0;
          const msLabel = `${Math.round(ms)}ms`;
          return (
            <div key={c.key} className="p-4 border rounded-xl border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Icon className={`w-4 h-4 ${c.color}`} />
                  <div className="text-sm font-semibold text-gray-900 dark:text-white">{c.label}</div>
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400">{msLabel}</div>
              </div>
              <div className="text-[11px] text-gray-500 dark:text-gray-400 mt-1">{symbols.join(', ')}</div>
            </div>
          );
        })}
      </div>

      <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mt-3">
        <Clock className="w-3.5 h-3.5" />
        <span>Güncellendi: {updatedAt ? updatedAt.toLocaleTimeString('tr-TR') : (loading ? 'Yükleniyor...' : '-')}</span>
      </div>
    </div>
  );
}
