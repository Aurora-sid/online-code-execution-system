import { createRouter, createWebHistory } from 'vue-router'
import { verifyToken } from '../api'

const routes = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('../views/LoginView.vue'),
        meta: { requiresGuest: true }
    },
    {
        path: '/',
        name: 'Editor',
        component: () => import('../views/EditorView.vue'),
        meta: { requiresAuth: true }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 带有后端令牌验证的导航守卫
router.beforeEach(async (to, from, next) => {
    const token = localStorage.getItem('token')

    if (to.meta.requiresAuth) {
        if (!token) {
            // No token, go to login
            next('/login')
        } else {
            // Verify token with backend
            const result = await verifyToken()
            if (result && result.valid) {
                // Token is valid, update user info
                localStorage.setItem('user', JSON.stringify(result.user))
                next()
            } else {
                // Token invalid or expired, clear and redirect
                localStorage.removeItem('token')
                localStorage.removeItem('user')
                next('/login')
            }
        }
    } else if (to.meta.requiresGuest && token) {
        // Already logged in, redirect to home
        next('/')
    } else {
        next()
    }
})

export default router
