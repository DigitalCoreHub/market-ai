"use client";


type Props = {
  symbol: string;
  positive: number;
  negative: number;
  neutral: number;
  avg: number; // -1..+1
};

export default function SentimentGauge({ symbol, positive, negative, neutral, avg }: Props) {
  const total = Math.max(positive + negative + neutral, 1);
  const posPct = Math.round((positive / total) * 100);
  const negPct = Math.round((negative / total) * 100);
  const neuPct = 100 - posPct - negPct;
  const avgPct = Math.round(((avg + 1) / 2) * 100); // map -1..+1 -> 0..100

  return (
    <div className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
      <div className="flex items-center justify-between mb-2">
        <div className="text-sm font-semibold text-gray-900 dark:text-white">{symbol}</div>
        <div className="text-xs text-gray-500 dark:text-gray-400">Ortalama: {avg.toFixed(2)}</div>
      </div>
      <div className="h-2 rounded-full overflow-hidden bg-gray-100 dark:bg-gray-800">
        <div className="h-full bg-green-500" style={{ width: `${posPct}%` }} />
        <div className="h-full bg-gray-400" style={{ width: `${neuPct}%` }} />
        <div className="h-full bg-red-500" style={{ width: `${negPct}%` }} />
      </div>
      <div className="flex items-center justify-between text-[10px] text-gray-500 dark:text-gray-400 mt-1">
        <span>Pozitif {posPct}%</span>
        <span>Nötr {neuPct}%</span>
        <span>Negatif {negPct}%</span>
      </div>
      <div className="mt-3">
  <div className="w-full h-1 bg-linear-to-r from-red-500 via-gray-400 to-green-500 rounded-full relative">
          <div
            className="absolute -top-1 w-3 h-3 rounded-full bg-blue-500 border-2 border-white dark:border-gray-900"
            style={{ left: `calc(${avgPct}% - 6px)` }}
          />
        </div>
        <div className="flex justify-between text-[10px] text-gray-500 dark:text-gray-400 mt-1">
          <span>-1</span>
          <span>0</span>
          <span>+1</span>
        </div>
      </div>
    </div>
  );
}
