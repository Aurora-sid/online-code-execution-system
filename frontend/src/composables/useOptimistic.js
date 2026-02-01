import { ref, watch } from 'vue'
import { cloneDeep } from 'lodash-es' // 你可能需要安装 lodash-es: npm i lodash-es

/**
 * 乐观更新 Composable
 * @param {Function} asyncOperation - 实际的异步操作函数 (api call)
 * @param {Object} options - 配置项
 * @param {Ref} options.state - 需要乐观更新的响应式状态 (Ref)
 * @param {Function} options.updater - (state, ...args) => void 更新状态的函数
 * @param {Function} [options.onSuccess] - 成功回调
 * @param {Function} [options.onError] - 失败回调
 */
export function useOptimistic(asyncOperation, { state, updater, onSuccess, onError }) {
    const isOptimistic = ref(false)
    const isError = ref(false)
    const error = ref(null)

    const execute = async (...args) => {
        isOptimistic.value = true
        isError.value = false
        error.value = null

        // 1. 创建状态快照 (Snapshot)
        // 根据 state 的类型选择合适的克隆方式
        const originalState = cloneDeep(state.value)

        // 2. 乐观更新 UI
        try {
            updater(state.value, ...args)
        } catch (e) {
            console.error('Optimistic updater failed:', e)
            // 如果更新器自己都挂了，就没必要继续了
            return
        }

        try {
            // 3. 执行实际的异步操作
            const result = await asyncOperation(...args)

            // 成功：保持乐观状态，或者利用由于服务器返回的数据做进一步修正
            // 如果 api 返回了最新的数据，这里可以选择再次更新 state
            isOptimistic.value = false
            if (onSuccess) onSuccess(result)
            return result

        } catch (err) {
            // 4. 失败：回滚 (Rollback)
            console.error('Async operation failed, rolling back.', err)
            state.value = originalState

            isError.value = true
            error.value = err
            isOptimistic.value = false

            if (onError) onError(err)

            throw err
        }
    }

    return {
        execute,
        isOptimistic,
        isError,
        error
    }
}
