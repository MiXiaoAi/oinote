import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            component: () => import('../layouts/MainLayout.vue'),
            children: [
                { path: '', name: 'home', component: () => import('../views/Home.vue') },
                { path: 'notes', name: 'notes', component: () => import('../views/NotesList.vue') },
                { path: 'channels', name: 'channels', component: () => import('../views/ChannelsList.vue') },
                { path: 'channel/:id', name: 'channel', component: () => import('../views/ChannelView.vue') },
                { path: 'note/:id?', name: 'note-editor', component: () => import('../views/NoteEditor.vue') },
                { path: 'approvals', name: 'approvals', component: () => import('../views/Approvals.vue'), meta: { requiresAuth: true } },
            ],
        },
        {
            path: '/login',
            name: 'login',
            component: () => import('../views/Auth/Login.vue'),
        },
        {
            path: '/register',
            name: 'register',
            component: () => import('../views/Auth/Register.vue'),
        },
    ],
});

// 添加路由守卫
router.beforeEach((to, from, next) => {
    const authStore = useAuthStore();

    // 检查路由是否需要认证
    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
        // 需要认证但用户未登录，重定向到登录页
        next({ name: 'login', query: { redirect: to.fullPath } });
    } else {
        // 不需要认证或用户已登录，继续导航
        next();
    }
});

export default router;
