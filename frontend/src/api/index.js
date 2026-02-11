import axios from 'axios'
import router from '../router'

// 获取 API 基础 URL
const getBaseURL = () => {
    // 优先使用环境变量，否则使用 Vite 代理
    const envUrl = import.meta.env.VITE_API_URL
    return envUrl ? `${envUrl}/api` : '/api'
}

// 创建带有基础URL的axios实例
const api = axios.create({
    baseURL: getBaseURL(),
    timeout: 30000, // 增加超时时间，代码执行可能需要更长时间
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

// 运行代码
export const runCode = async (language, code) => {
    const response = await api.post('/run', { language, code })
    return response.data
}

// 获取历史记录
export const fetchSubmissions = async () => {
    const response = await api.get('/submissions')
    return response.data.submissions || []
}

// 获取 WebSocket URL
export const getWebSocketURL = (taskId) => {
    const apiUrl = import.meta.env.VITE_API_URL || window.location.origin
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = apiUrl.replace(/^https?:\/\//, '')
    return `${wsProtocol}//${host}/api/ws?taskId=${taskId}`
}

// 智能代码分析 - 调用 LLM API
// AI 分析需要更长超时时间，大模型响应可能需要 30-60 秒
export const analyzeCode = async (code, language, type = 'bug') => {
    const response = await api.post('/ai/analyze', { code, language, type }, {
        timeout: 90000 // 90 秒超时，AI 分析需要更长时间
    })
    return response.data
}

// 报错分析 - 分析代码运行时的错误
// AI 分析需要更长超时时间
export const analyzeError = async (code, language, errorOutput) => {
    // 将错误输出作为代码的一部分传给 AI 分析
    const combinedCode = `=== 用户代码 ===\n${code}\n\n=== 终端输出/报错信息 ===\n${errorOutput}`
    const response = await api.post('/ai/analyze', {
        code: combinedCode,
        language,
        type: 'error'
    }, {
        timeout: 90000 // 90 秒超时
    })
    return response.data
}

export default api


