"use client";

import SentimentGauge from '@/components/SentimentGauge';
import { useMarketContext } from '@/lib/marketContext';

export default function StockSentimentPanel({ symbols }: { symbols: string[] }) {
  const { data, loading, error, refresh } = useMarketContext(symbols, 60000);
  const sentiments = data?.stock_sentiments || {};

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-gray-900 dark:text-white">Piyasa Duyarlılığı</h2>
        <button onClick={refresh} className="text-xs px-3 py-1 rounded-full border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800">Yenile</button>
      </div>
      {error && <div className="text-sm text-red-600 dark:text-red-400 mb-2">{error}</div>}
      {loading && !data && <div className="text-sm text-gray-500 dark:text-gray-400">Yükleniyor...</div>}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {symbols.map(sym => {
          const s = sentiments[sym];
          if (!s) return (
            <div key={sym} className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 text-sm text-gray-500 dark:text-gray-400">{sym}: Veri yok</div>
          );
          return (
            <SentimentGauge
              key={sym}
              symbol={sym}
              positive={s.positive_count}
              negative={s.negative_count}
              neutral={s.neutral_count}
              avg={s.avg_sentiment}
            />
          );
        })}
      </div>
    </div>
  );
}
