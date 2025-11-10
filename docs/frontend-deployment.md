# 🎨 Frontend Deployment Guide - Vercel

Bu dokümantasyon frontend'in Vercel'de deploy edilmesi ve backend'e bağlanması için gerekli adımları içerir.

---

## 📋 Önkoşullar

1. **Vercel hesabı** → [https://vercel.com](https://vercel.com)
2. **Vercel CLI** (opsiyonel):
   ```bash
   npm i -g vercel
   ```
3. **GitHub repository** (Vercel otomatik deploy için)

---

## 🔧 Frontend Yapılandırması

### Environment Variables

Frontend artık environment variables kullanıyor. Hardcoded `localhost:8080` URL'leri kaldırıldı.

**Güncellenen dosyalar:**

- `frontend/components/Dashboard.tsx`
- `frontend/components/Leaderboard.tsx`
- `frontend/lib/marketContext.ts`
- `frontend/lib/config.ts` (yeni - config helper)

---

## 🚀 Vercel'de Deploy

### Yöntem 1: Vercel Dashboard (Önerilen)

1. **Vercel Dashboard'a git:**

   - [https://vercel.com/dashboard](https://vercel.com/dashboard)
   - "Add New..." → "Project"

2. **GitHub repository'yi bağla:**

   - Repository'yi seç
   - "Import" butonuna tıkla

3. **Project ayarları:**

   - **Framework Preset:** Next.js
   - **Root Directory:** `frontend`
   - **Build Command:** `npm run build` (otomatik)
   - **Output Directory:** `.next` (otomatik)

4. **Environment Variables ekle:**

   - Project Settings → Environment Variables
   - Şu değişkenleri ekle:

   | Key                   | Value                                  | Environment                      |
   | --------------------- | -------------------------------------- | -------------------------------- |
   | `NEXT_PUBLIC_API_URL` | `https://marketai-backend.fly.dev/api` | Production, Preview, Development |
   | `NEXT_PUBLIC_WS_URL`  | `wss://marketai-backend.fly.dev/ws`    | Production, Preview, Development |

5. **Deploy:**
   - "Deploy" butonuna tıkla
   - Vercel otomatik olarak build edip deploy eder

### Yöntem 2: Vercel CLI

```bash
cd frontend
vercel
```

**Sorular:**

- Set up and deploy? → **Yes**
- Which scope? → Hesabını seç
- Link to existing project? → **No** (ilk deploy)
- Project name → `market-ai-frontend` (veya istediğin isim)
- Directory → `./`
- Override settings? → **No**

**Environment Variables ekle:**

```bash
vercel env add NEXT_PUBLIC_API_URL
# Value: https://marketai-backend.fly.dev/api

vercel env add NEXT_PUBLIC_WS_URL
# Value: wss://marketai-backend.fly.dev/ws
```

**Deploy:**

```bash
vercel --prod
```

---

## 🔗 Backend Bağlantısı

### Backend URL'leri

**Production:**

- API URL: `https://marketai-backend.fly.dev/api`
- WebSocket URL: `wss://marketai-backend.fly.dev/ws`

**Local Development:**

- API URL: `http://localhost:8080/api`
- WebSocket URL: `ws://localhost:8080/ws`

### Environment Variables

Frontend şu environment variables'ı kullanıyor:

- `NEXT_PUBLIC_API_URL`: Backend API base URL
- `NEXT_PUBLIC_WS_URL`: WebSocket URL

**Not:** `NEXT_PUBLIC_` prefix'i Next.js'te client-side'da kullanılabilir değişkenler için gereklidir.

---

## 🧪 Test

### 1. Backend Test

```bash
# Health check
curl https://marketai-backend.fly.dev/health

# Ping test
curl https://marketai-backend.fly.dev/api/v1/ping
```

### 2. Frontend Test

Deploy sonrası:

1. Vercel'den verilen URL'e git
2. Dashboard açılmalı
3. WebSocket bağlantısı kurulmalı
4. API çağrıları çalışmalı

### 3. Browser Console Kontrolü

Browser Developer Tools → Console:

- WebSocket bağlantı mesajları görünmeli
- API çağrıları başarılı olmalı
- Hata mesajı olmamalı

---

## 🔄 Güncelleme

### Frontend'i güncelleme:

```bash
# Git'e push et
git push origin main

# Vercel otomatik olarak deploy eder
```

### Environment Variables güncelleme:

1. Vercel Dashboard → Project Settings → Environment Variables
2. Değişkeni güncelle
3. "Redeploy" butonuna tıkla

---

## 🐛 Sorun Giderme

### Backend'e bağlanamıyor:

1. **Backend çalışıyor mu?**

   ```bash
   curl https://marketai-backend.fly.dev/health
   ```

2. **Environment Variables doğru mu?**

   - Vercel Dashboard → Environment Variables kontrol et
   - `NEXT_PUBLIC_API_URL` ve `NEXT_PUBLIC_WS_URL` doğru mu?

3. **CORS hatası var mı?**
   - Backend'de CORS ayarları kontrol et
   - Frontend URL'i backend'de allowed origins'de mi?

### WebSocket bağlanamıyor:

1. **WebSocket URL doğru mu?**

   - `wss://` (secure) kullanılıyor mu?
   - URL'de `/ws` var mı?

2. **Backend WebSocket çalışıyor mu?**
   ```bash
   fly logs -a marketai-backend | grep -i websocket
   ```

### Build hatası:

1. **TypeScript hataları:**

   ```bash
   cd frontend
   npm run build
   ```

2. **Dependency hataları:**
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   ```

---

## ✅ Deployment Checklist

- [ ] Backend deploy edildi ve çalışıyor
- [ ] Backend health check başarılı
- [ ] Frontend environment variables ayarlandı
- [ ] Frontend Vercel'de deploy edildi
- [ ] Frontend backend'e bağlanabiliyor
- [ ] WebSocket bağlantısı kuruluyor
- [ ] API çağrıları çalışıyor
- [ ] Dashboard görüntüleniyor

---

**Son Güncelleme:** 2025-01-XX
