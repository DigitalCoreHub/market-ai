'use client';

import { formatDistanceToNow } from 'date-fns';
import { tr } from 'date-fns/locale';
import { AlertCircle, ExternalLink, Newspaper, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';

interface NewsArticle {
  id: string;
  title: string;
  description: string;
  source: string;
  url: string;
  related_stocks: string[];
  published_at: string;
  impact_level?: string;
}

interface WebSocketMessage {
  type: string;
  data: NewsArticle[] | Record<string, unknown>;
  timestamp: number;
}

export default function LatestNews({ lastMessage }: { lastMessage: WebSocketMessage | null }) {
  const [news, setNews] = useState<NewsArticle[]>([]);

  // eslint-disable react-hooks/exhaustive-deps
  useEffect(() => {
    if (!lastMessage || lastMessage.type !== 'news_update') return;

    const articles = (lastMessage.data as NewsArticle[]) || [];
    setNews(articles);
  }, [lastMessage]);

  const getImpactColor = (level?: string) => {
    switch (level?.toLowerCase()) {
      case 'high':
        return 'bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-900';
      case 'medium':
        return 'bg-yellow-50 dark:bg-yellow-950/30 border-yellow-200 dark:border-yellow-900';
      case 'low':
        return 'bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-900';
      default:
        return 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700';
    }
  };

  const getImpactIcon = (level?: string) => {
    switch (level?.toLowerCase()) {
      case 'high':
        return <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400" />;
      case 'medium':
        return <TrendingUp className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />;
      default:
        return <Newspaper className="w-4 h-4 text-gray-600 dark:text-gray-400" />;
    }
  };

  return (
    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="bg-linear-to-br from-cyan-500 to-blue-600 p-2 rounded-lg">
          <Newspaper className="w-5 h-5 text-white" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">En Son Pazar Haberleri</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">News API ve RSS beslemeleri</p>
        </div>
      </div>

      <div className="space-y-3 max-h-[600px] overflow-y-auto">
        {news.length > 0 ? (
          news.map((article) => (
            <div
              key={article.id}
              className={`p-4 border rounded-lg hover:shadow-md dark:hover:shadow-black/50 transition-all ${getImpactColor(article.impact_level)}`}
            >
              {/* Header with icon and title */}
              <div className="flex items-start gap-3 mb-2">
                {getImpactIcon(article.impact_level)}
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-sm text-gray-900 dark:text-white line-clamp-2">
                    {article.title}
                  </h3>
                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                    {article.source} • {formatDistanceToNow(new Date(article.published_at), { addSuffix: true, locale: tr })}
                  </p>
                </div>
              </div>

              {/* Description */}
              {article.description && (
                <p className="text-sm text-gray-700 dark:text-gray-300 mb-3 line-clamp-2">
                  {article.description}
                </p>
              )}

              {/* Related stocks */}
              {article.related_stocks && article.related_stocks.length > 0 && (
                <div className="flex flex-wrap gap-2 mb-3">
                  {article.related_stocks.map((stock) => (
                    <span
                      key={stock}
                      className="inline-block px-2 py-1 bg-blue-100 dark:bg-blue-900/50 text-blue-800 dark:text-blue-200 rounded text-xs font-medium"
                    >
                      {stock}
                    </span>
                  ))}
                </div>
              )}

              {/* Read more link */}
              {article.url && (
                <a
                  href={article.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
                >
                  Devamını oku
                  <ExternalLink className="w-3 h-3" />
                </a>
              )}
            </div>
          ))
        ) : (
          <div className="text-center text-gray-500 dark:text-gray-400 py-8">
            <Newspaper className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>Henüz haber mevcut değil</p>
          </div>
        )}
      </div>
    </div>
  );
}
