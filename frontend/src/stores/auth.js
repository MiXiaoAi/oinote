import { defineStore } from 'pinia';
import api from '../api/axios';
import { connectWebSocket, disconnectWebSocket } from '../utils/websocket';

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: JSON.parse(localStorage.getItem('user')) || null,
        token: localStorage.getItem('token') || null,
    }),
    getters: {
        isAuthenticated: (state) => !!state.token,
    },
    actions: {
        async login(username, password) {
            const res = await api.post('/login', { username, password });
            this.token = res.data.token;
            this.user = res.data.user;
            localStorage.setItem('token', this.token);
            localStorage.setItem('user', JSON.stringify(this.user));
            
            // 连接 WebSocket
            if (this.user.id) {
                localStorage.setItem('userId', this.user.id);
                connectWebSocket(this.user.id);
            }
        },
        async refreshMe() {
            const res = await api.get('/me');
            this.user = res.data;
            localStorage.setItem('user', JSON.stringify(this.user));
            
            // 连接 WebSocket（如果还没连接）
            if (this.user.id) {
                localStorage.setItem('userId', this.user.id);
                connectWebSocket(this.user.id);
            }
        },
        async updateMe(payload) {
            const res = await api.put('/me', payload);
            this.user = res.data;
            localStorage.setItem('user', JSON.stringify(this.user));
        },
        async register(username, password) {
            await api.post('/register', { username, password });
        },
        logout() {
            // 断开 WebSocket
            disconnectWebSocket();
            localStorage.removeItem('userId');
            
            this.token = null;
            this.user = null;
            localStorage.removeItem('token');
            localStorage.removeItem('user');
        }
    },
});
