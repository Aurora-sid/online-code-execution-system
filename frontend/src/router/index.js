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
    },
    {
        path: '/admin',
        name: 'Admin',
        component: () => import('../views/AdminView.vue'),
        meta: { requiresAuth: true, requiresAdmin: true }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 带有后端令牌验证的导航守卫 (优化版：乐观跳转 + 后台验证)
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token')

    // 尝试从 localStorage 获取用户信息
    let user = null
    try {
        const userStr = localStorage.getItem('user')
        if (userStr) {
            user = JSON.parse(userStr)
        }
    } catch (e) {
        console.error('Failed to parse user info', e)
        localStorage.removeItem('user')
    }

    if (to.meta.requiresAuth) {
        if (!token) {
            // 没有 token，直接去登录
            next('/login')
        } else {
            // ✅ 只要本地有 token，先通过 (乐观策略)
            // 这样可以消除路由跳转时的等待时间

            // 简单的权限检查 (基于本地缓存的角色)
            if (to.meta.requiresAdmin && user && user.role !== 'admin') {
                next('/')
                return
            }

            // 放行
            next()

            // 🚀 在后台异步验证 Token 有效性
            // 如果 Token 失效，会自动跳转到登录页 (由 api/index.js 中的拦截器处理)
            verifyToken().then(result => {
                if (result && result.valid) {
                    // Token 有效，更新本地用户信息
                    localStorage.setItem('user', JSON.stringify(result.user))

                    // 如果刚才因为本地没有 user 信息而放行了 (虽然罕见)，
                    // 这里不需要做额外处理，因为已经 next() 过了。
                    // Vue 的响应式系统会自动更新界面。
                } else {
                    // Token 无效 (后端返回 200 但 valid=false，或者请求失败返回 null)
                    // 注意：401 错误已经被拦截器处理并跳转了，这里处理的是业务逻辑上的无效
                    // 清除失效数据
                    localStorage.removeItem('token')
                    localStorage.removeItem('user')

                    // 只有在当前还在受保护路由时才跳转 (避免用户已经手动跳到公开页)
                    if (router.currentRoute.value.meta.requiresAuth) {
                        router.push('/login')
                    }
                }
            }).catch(e => {
                // 网络错误等，暂不强制登出，以免断网时无法使用
                console.error('Background token verification failed:', e)
            })
        }
    } else if (to.meta.requiresGuest && token) {
        // 已登录用户访问登录页 -> 跳转首页
        next('/')
    } else {
        next()
    }
})

export default router

