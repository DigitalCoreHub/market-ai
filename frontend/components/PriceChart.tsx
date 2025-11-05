'use client';

import { format } from 'date-fns';
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

export interface PricePoint {
  time: number; // epoch millis
  price: number;
}

interface PriceChartProps {
  data: PricePoint[];
  color?: string; // tailwind color hex or css color
  height?: number;
}

export default function PriceChart({ data, color = '#8b5cf6', height = 260 }: PriceChartProps) {
  const gridColor = 'rgba(148, 163, 184, 0.2)';
  const gradientId = 'priceGradient';

  return (
    <div className="w-full h-full">
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={data} margin={{ top: 10, right: 10, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.35} />
              <stop offset="95%" stopColor={color} stopOpacity={0.02} />
            </linearGradient>
          </defs>
          <CartesianGrid stroke={gridColor} strokeDasharray="3 3" />
          <XAxis
            dataKey="time"
            tick={{ fontSize: 11 }}
            tickFormatter={(ts) => format(new Date(ts), 'HH:mm')}
          />
          <YAxis
            tick={{ fontSize: 11 }}
            width={45}
            tickFormatter={(v) => `₺${Number(v).toFixed(0)}`}
            domain={["auto", "auto"]}
          />
          <Tooltip
            contentStyle={{ background: 'rgba(17, 24, 39, 0.8)', border: '1px solid #374151', borderRadius: 12 }}
            labelClassName="text-gray-100"
            itemStyle={{ color: '#e5e7eb' }}
            labelFormatter={(ts) => format(new Date(Number(ts)), 'HH:mm:ss')}
            formatter={(value: number | string) => [`₺${Number(value).toFixed(2)}`, 'Price']}
          />
          <Area type="monotone" dataKey="price" stroke={color} strokeWidth={2} fill={`url(#${gradientId})`} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
