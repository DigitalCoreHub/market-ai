"use client";

import { useEffect, useState } from 'react';

export interface StockPrice {
  symbol: string;
  price: number;
  open: number;
  high: number;
  low: number;
  volume: number;
  source: string;
  timestamp: string;
  delay_minutes: number;
}

export interface ScrapedArticle {
  source: string;
  title: string;
  url: string;
  related_stocks?: string[];
  scraped_at: string;
}

export interface TweetSentiment {
  id: string;
  text: string;
  author: string;
  author_followers: number;
  sentiment_score: number;
  sentiment_label: string;
  impact_score: number;
  stock_symbols: string[];
}

export interface StockSentimentAgg {
  symbol: string;
  tweet_count: number;
  avg_sentiment: number;
  positive_count: number;
  negative_count: number;
  neutral_count: number;
  top_tweet?: TweetSentiment;
}

export interface MarketContextData {
  prices: StockPrice[];
  news: ScrapedArticle[];
  tweets: TweetSentiment[];
  stock_sentiments: Record<string, StockSentimentAgg>;
  updated_at: string;
  fetch_durations: Record<string, number>; // ms durations serialized
}

interface ResponseWrapper {
  success: boolean;
  data?: MarketContextData;
  message?: string;
}

export function useMarketContext(symbols: string[], pollIntervalMs = 60000) {
  const [data, setData] = useState<MarketContextData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchContext = () => {
    if (!symbols || symbols.length === 0) return;
    const param = symbols.join(',');
    fetch(`http://localhost:8080/api/v1/market/context?symbols=${param}`)
      .then(r => r.json())
      .then((res: ResponseWrapper) => {
        if (res.success && res.data) {
          setData(res.data);
          setError(null);
        } else {
          setError(res.message || 'Sunucu hatası');
        }
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchContext();
    if (pollIntervalMs > 0) {
      const id = setInterval(fetchContext, pollIntervalMs);
      return () => clearInterval(id);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [symbols.join(',')]);

  return { data, loading, error, refresh: fetchContext };
}
