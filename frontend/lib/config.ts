// Frontend configuration - API and WebSocket URLs
export const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
export const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';

// Remove trailing slashes
export const cleanApiUrl = API_URL.replace(/\/$/, '');
export const cleanWsUrl = WS_URL.replace(/\/$/, '');

