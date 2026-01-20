import axios from 'axios'
import router from '../router'

// 创建带有基础URL的axios实例
const api = axios.create({
    baseURL: '/api',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// 请求拦截器 - 为所有请求添加令牌
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => Promise.reject(error)
)

// 响应拦截器 - 全局处理401错误
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response && error.response.status === 401) {
            // Clear invalid token
            localStorage.removeItem('token')
            localStorage.removeItem('user')
            // Redirect to login
            router.push('/login')
        }
        return Promise.reject(error)
    }
)

// 与后端验证令牌
export const verifyToken = async () => {
    try {
        const response = await api.get('/verify')
        return response.data
    } catch (error) {
        return null
    }
}

// 获取支持的编程语言列表
export const fetchLanguages = async () => {
    try {
        const response = await api.get('/languages')
        return response.data.languages || []
    } catch (error) {
        console.error('Failed to fetch languages:', error)
        return []
    }
}

export default api

