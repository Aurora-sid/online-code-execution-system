// Admin API 模块
// 负责与后端管理员相关的接口交互，如获取统计数据、管理用户、查看日志等
// 所有接口请求都要加上 /api 这个前缀
const API_BASE = '/api'

// 获取认证头部，包含 JWT token
const getAuthHeaders = () => {
    const token = localStorage.getItem('token')
    return {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
    }
}

export const adminAPI = {
    // 获取系统统计
    async getStats() {
        const res = await fetch(`${API_BASE}/admin/stats`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch stats')
        return res.json()
    },

    // 获取用户列表
    async getUsers(page = 1, pageSize = 50) {
        const res = await fetch(`${API_BASE}/admin/users?page=${page}&pageSize=${pageSize}`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch users')
        return res.json()
    },

    // 创建用户
    async createUser(userData) {
        const res = await fetch(`${API_BASE}/admin/users`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(userData)
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to create user')
        }
        return res.json()
    },

    // 删除用户
    async deleteUser(userId) {
        const res = await fetch(`${API_BASE}/admin/users/${userId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to delete user')
        }
        return res.json()
    },

    // 获取提交记录
    async getSubmissions(page = 1, pageSize = 50, status = '') {
        let url = `${API_BASE}/admin/submissions?page=${page}&pageSize=${pageSize}`
        if (status) url += `&status=${status}`
        const res = await fetch(url, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch submissions')
        return res.json()
    },

    // 获取本周提交统计
    async getWeeklyStats() {
        const res = await fetch(`${API_BASE}/admin/stats/weekly`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch weekly stats')
        return res.json()
    },

    // 重置容器池
    async resetPool() {
        const res = await fetch(`${API_BASE}/admin/pool/reset`, {
            method: 'POST',
            headers: getAuthHeaders()
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to reset pool')
        }
        return res.json()
    },

    // ==================== LLM 大模型管理 ====================

    // 获取 LLM 状态和使用统计
    async getLLMStatus() {
        const res = await fetch(`${API_BASE}/admin/llm/status`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch LLM status')
        return res.json()
    },

    // 启用/禁用 LLM
    async toggleLLM(enabled) {
        const res = await fetch(`${API_BASE}/admin/llm/toggle`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ enabled })
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to toggle LLM')
        }
        return res.json()
    },

    // 切换 LLM 模型
    async setLLMModel(model) {
        const res = await fetch(`${API_BASE}/admin/llm/model`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify({ model })
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to set LLM model')
        }
        return res.json()
    },

    // 重置 LLM 使用统计
    async resetLLMStats() {
        const res = await fetch(`${API_BASE}/admin/llm/stats/reset`, {
            method: 'POST',
            headers: getAuthHeaders()
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to reset LLM stats')
        }
        return res.json()
    },

    // ==================== 后端日志查看 ====================

    // 获取后端运行日志
    async getLogs(count = 200) {
        const res = await fetch(`${API_BASE}/admin/logs?count=${count}`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch logs')
        return res.json()
    },

    // 清空日志缓冲区
    async clearLogs() {
        const res = await fetch(`${API_BASE}/admin/logs/clear`, {
            method: 'POST',
            headers: getAuthHeaders()
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to clear logs')
        }
        return res.json()
    }
}
