import { ref, computed } from 'vue'

// 简单的响应式状态管理 (不使用 Pinia 以保持简洁)
const token = ref(localStorage.getItem('token') || '')
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

export function useAuth() {
    const isAuthenticated = computed(() => !!token.value)

    const login = (tokenValue, userData) => {
        token.value = tokenValue
        user.value = userData
        localStorage.setItem('token', tokenValue)
        localStorage.setItem('user', JSON.stringify(userData))
    }

    const logout = () => {
        token.value = ''
        user.value = null
        localStorage.removeItem('token')
        localStorage.removeItem('user')
    }

    const getToken = () => token.value

    const getUser = () => user.value

    return {
        token,
        user,
        isAuthenticated,
        login,
        logout,
        getToken,
        getUser
    }
}
