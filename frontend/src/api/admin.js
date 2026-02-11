// Admin API 模块
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

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
        const res = await fetch(`${API_BASE}/api/admin/stats`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch stats')
        return res.json()
    },

    // 获取用户列表
    async getUsers(page = 1, pageSize = 50) {
        const res = await fetch(`${API_BASE}/api/admin/users?page=${page}&pageSize=${pageSize}`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch users')
        return res.json()
    },

    // 创建用户
    async createUser(userData) {
        const res = await fetch(`${API_BASE}/api/admin/users`, {
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
        const res = await fetch(`${API_BASE}/api/admin/users/${userId}`, {
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
        let url = `${API_BASE}/api/admin/submissions?page=${page}&pageSize=${pageSize}`
        if (status) url += `&status=${status}`
        const res = await fetch(url, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch submissions')
        return res.json()
    },

    // 获取本周提交统计
    async getWeeklyStats() {
        const res = await fetch(`${API_BASE}/api/admin/stats/weekly`, {
            headers: getAuthHeaders()
        })
        if (!res.ok) throw new Error('Failed to fetch weekly stats')
        return res.json()
    },

    // 重置容器池
    async resetPool() {
        const res = await fetch(`${API_BASE}/api/admin/pool/reset`, {
            method: 'POST',
            headers: getAuthHeaders()
        })
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.error || 'Failed to reset pool')
        }
        return res.json()
    }
}
