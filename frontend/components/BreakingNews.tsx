"use client";

import { useMarketContext } from '@/lib/marketContext';
import { ExternalLink, Newspaper } from 'lucide-react';

export default function BreakingNews({ symbols }: { symbols: string[] }) {
  const { data, loading, error, refresh } = useMarketContext(symbols, 0); // manual refresh
  const news = data?.news || [];

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Newspaper className="w-5 h-5 text-blue-500" />
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">Son Gelişmeler (Scraper)</h2>
        </div>
        <button onClick={refresh} className="text-xs px-3 py-1 rounded-full border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800">Yenile</button>
      </div>
      {error && <div className="text-sm text-red-600 dark:text-red-400 mb-2">{error}</div>}
      {loading && !data && <div className="text-sm text-gray-500 dark:text-gray-400">Yükleniyor...</div>}
      <div className="space-y-3">
        {news.slice(0, 8).map((n, idx) => (
          <div key={idx} className="p-3 border rounded-lg border-gray-200 dark:border-gray-800">
            <div className="text-sm font-semibold text-gray-900 dark:text-white line-clamp-2">{n.title}</div>
            <div className="text-[11px] text-gray-500 dark:text-gray-400 mt-1">{n.source} • {new Date(n.scraped_at).toLocaleString('tr-TR')}</div>
            <div className="flex items-center justify-between mt-2">
              <div className="flex gap-1 flex-wrap">
                {(n.related_stocks || []).map(s => (
                  <span key={s} className="px-2 py-0.5 rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 text-[10px]">{s}</span>
                ))}
              </div>
              {n.url && (
                <a href={n.url} target="_blank" rel="noreferrer" className="text-[11px] inline-flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:underline">
                  Oku <ExternalLink className="w-3 h-3" />
                </a>
              )}
            </div>
          </div>
        ))}
        {news.length === 0 && !loading && (
          <div className="text-sm text-gray-500 dark:text-gray-400">Haber bulunamadı.</div>
        )}
      </div>
    </div>
  );
}
