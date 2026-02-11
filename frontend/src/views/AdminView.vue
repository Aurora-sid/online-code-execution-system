<template>
  <div class="min-h-screen bg-gray-50 text-gray-900 font-sans">
    <!-- 顶部导航栏 -->
    <header class="fixed top-4 left-4 right-4 z-50 bg-white/80 backdrop-blur-xl border border-gray-200 rounded-2xl shadow-sm">
      <div class="flex items-center justify-between px-6 py-3">
        <div class="flex items-center gap-4">
          <router-link 
            to="/" 
            class="hidden md:flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-gray-500 hover:text-gray-900 hover:bg-gray-100/50 transition-colors"
            title="返回代码编辑器"
          >
            <i class="ph ph-arrow-left text-lg"></i>
          </router-link>
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center shadow-md">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                <path d="M2 17l10 5 10-5"/>
                <path d="M2 12l10 5 10-5"/>
              </svg>
            </div>
            <div>
              <h1 class="text-lg font-bold tracking-wide">Aurora Admin</h1>
              <p class="text-xs text-gray-500">System Dashboard</p>
            </div>
          </div>
        </div>
        
        <!-- 导航标签 -->
        <nav class="flex items-center gap-1 bg-gray-100 rounded-xl p-1">
          <button 
            v-for="tab in tabs" 
            :key="tab.id"
            @click="activeTab = tab.id"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 cursor-pointer"
            :class="activeTab === tab.id 
              ? 'bg-white text-cyan-600 shadow-sm' 
              : 'text-gray-500 hover:text-gray-900 hover:bg-gray-200/50'"
          >
            <i :class="tab.icon" class="mr-2"></i>
            {{ tab.label }}
          </button>
        </nav>
        
        <!-- 用户菜单 -->
        <div class="flex items-center gap-4">
          <div class="text-right">
            <p class="text-sm font-medium text-gray-900">{{ user?.username }}</p>
            <p class="text-xs text-cyan-600">管理员</p>
          </div>
          <button 
            @click="handleLogout"
            class="p-2 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 transition-colors cursor-pointer"
            title="退出登录"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16 17 21 12 16 7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
          </button>
        </div>
      </div>
    </header>
    
    <!-- 主内容区 -->
    <main class="pt-24 px-6 pb-8">
      <!-- 仪表盘 -->
      <div v-if="activeTab === 'dashboard'" class="space-y-6">
        <!-- 统计卡片 -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div 
            v-for="stat in stats" 
            :key="stat.label"
            @click="handleStatClick(stat)"
            class="bg-white border border-gray-200 rounded-2xl p-6 hover:border-gray-300 hover:shadow-lg transition-all cursor-pointer group"
          >
            <div class="flex items-center justify-between mb-4">
              <div class="p-3 rounded-xl group-hover:scale-110 transition-transform" :class="stat.bgColor">
                <i :class="[stat.icon, stat.iconColor]" class="text-xl"></i>
              </div>
              <span class="text-xs px-2 py-1 rounded-full" :class="stat.changeClass">
                {{ stat.change }}
              </span>
            </div>
            <p class="text-3xl font-bold mb-1 text-gray-900">{{ stat.value }}</p>
            <p class="text-sm text-gray-500 group-hover:text-gray-700 transition-colors">{{ stat.label }}</p>
          </div>
        </div>
        
        <!-- 最近活动 -->
        <div class="bg-white border border-gray-200 rounded-2xl p-6 shadow-sm">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2 text-gray-800">
            <i class="ph ph-activity text-cyan-500"></i>
            最近提交
          </h3>
          <div class="overflow-x-auto max-h-[400px] overflow-y-auto">
            <table class="w-full text-sm">
              <thead class="sticky top-0 bg-white">
                <tr class="text-left text-gray-500 border-b border-gray-200">
                  <th class="pb-3 font-medium">用户</th>
                  <th class="pb-3 font-medium">语言</th>
                  <th class="pb-3 font-medium">状态</th>
                  <th class="pb-3 font-medium">时间</th>
                  <th class="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr 
                  v-for="sub in recentSubmissions" 
                  :key="sub.id"
                  class="hover:bg-gray-50 transition-colors"
                >
                  <td class="py-3">
                    <span class="text-gray-700">{{ sub.username }}</span>
                  </td>
                  <td class="py-3">
                    <span class="px-2 py-1 bg-gray-100 text-gray-600 rounded text-xs">{{ sub.language }}</span>
                  </td>
                  <td class="py-3">
                    <span 
                      class="px-2 py-1 rounded-full text-xs font-medium"
                      :class="getStatusClass(sub.status)"
                    >
                      {{ sub.status }}
                    </span>
                  </td>
                  <td class="py-3 text-gray-500">{{ formatTime(sub.createdAt) }}</td>
                  <td class="py-3">
                    <button 
                      @click="viewSubmission(sub)"
                      class="text-cyan-600 hover:text-cyan-500 transition-colors cursor-pointer font-medium"
                    >
                      查看详情
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      
      <!-- 用户管理 -->
      <div v-if="activeTab === 'users'" class="space-y-6">
        <div class="flex items-center justify-between">
          <h2 class="text-2xl font-bold text-gray-800">用户管理</h2>
          <button 
            @click="showCreateUserModal = true"
            class="px-4 py-2 bg-gradient-to-r from-cyan-500 to-blue-600 text-white rounded-xl font-medium hover:brightness-110 hover:shadow-lg transition-all cursor-pointer flex items-center gap-2"
          >
            <i class="ph ph-plus"></i>
            添加用户
          </button>
        </div>
        
        <div class="bg-white border border-gray-200 rounded-2xl overflow-hidden shadow-sm">
          <div class="max-h-[75vh] overflow-y-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 bg-gray-50 border-b border-gray-200">
                <th class="px-6 py-4 font-medium">ID</th>
                <th class="px-6 py-4 font-medium">用户名</th>
                <th class="px-6 py-4 font-medium">角色</th>
                <th class="px-6 py-4 font-medium">状态</th>
                <th class="px-6 py-4 font-medium">注册时间</th>
                <th class="px-6 py-4 font-medium">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              <tr v-for="u in users" 
                :key="u.id"
                class="hover:bg-gray-50 transition-colors"
              >
                <td class="px-6 py-4 text-gray-500">#{{ u.id }}</td>
                <td class="px-6 py-4 text-gray-900 font-medium">{{ u.username }}</td>
                <td class="px-6 py-4">
                  <span 
                    class="px-2 py-1 rounded-full text-xs font-medium"
                    :class="u.role === 'admin' ? 'bg-purple-100 text-purple-600' : 'bg-gray-100 text-gray-500'"
                  >
                    {{ u.role === 'admin' ? '管理员' : '用户' }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span 
                    class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium"
                    :class="u.isOnline ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'"
                  >
                    <span 
                      class="w-2 h-2 rounded-full"
                      :class="u.isOnline ? 'bg-green-500 animate-pulse' : 'bg-red-500'"
                    ></span>
                    {{ u.isOnline ? '在线' : '离线' }}
                  </span>
                </td>
                <td class="px-6 py-4 text-gray-500">{{ formatDate(u.createdAt) }}</td>
                <td class="px-6 py-4">
                  <button 
                    v-if="u.username !== 'admin'"
                    @click="confirmDeleteUser(u)"
                    class="text-red-500 hover:text-red-700 transition-colors cursor-pointer font-medium"
                  >
                    删除
                  </button>
                  <span v-else class="text-gray-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </div>
      </div>
      
      <!-- 运行日志 -->
      <div v-if="activeTab === 'logs'" class="space-y-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <h2 class="text-2xl font-bold text-gray-800">运行日志</h2>
            <!-- 今日筛选标签 -->
            <span 
              v-if="logFilter.dateFilter === 'today'"
              class="px-3 py-1 bg-purple-100 text-purple-600 rounded-full text-sm font-medium flex items-center gap-1"
            >
              <i class="ph ph-calendar"></i>
              仅今日
              <button 
                @click="logFilter.dateFilter = ''"
                class="ml-1 hover:text-white transition-colors cursor-pointer"
              >
                <i class="ph ph-x text-xs"></i>
              </button>
            </span>
          </div>
          <div class="flex items-center gap-2">
            <!-- 代码提交量按钮 -->
            <button 
              @click="toggleWeeklyChart"
              class="px-3 py-2 rounded-lg text-sm font-medium transition-all cursor-pointer border flex items-center gap-1"
              :class="showWeeklyChart 
                ? 'bg-cyan-50 text-cyan-600 border-cyan-200' 
                : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50 hover:text-gray-900'"
            >
              <i class="ph ph-chart-bar mr-1"></i>
              代码提交量
            </button>
            <!-- 日期筛选按钮 -->
            <button 
              @click="logFilter.dateFilter = logFilter.dateFilter === 'today' ? '' : 'today'"
              class="px-3 py-2 rounded-lg text-sm font-medium transition-all cursor-pointer border"
              :class="logFilter.dateFilter === 'today' 
                ? 'bg-purple-50 text-purple-600 border-purple-200' 
                : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50 hover:text-gray-900'"
            >
              <i class="ph ph-calendar mr-1"></i>
              今日
            </button>
            <select 
              v-model="logFilter.status"
              class="w-36 bg-white border border-gray-200 rounded-lg px-3 py-2 text-sm cursor-pointer focus:outline-none focus:border-cyan-500 text-gray-700"
            >
              <option value="">全部状态</option>
              <option value="Success">成功</option>
              <option value="Failed">失败</option>
              <option value="Timeout">超时</option>
            </select>
          </div>
        </div>
        
        <!-- 周提交量图表 -->
        <div v-if="showWeeklyChart" class="bg-white border border-gray-200 rounded-2xl p-6 shadow-sm mb-4 transition-all">
          <h3 class="text-base font-semibold text-gray-700 mb-4 flex items-center gap-2">
            <i class="ph ph-chart-bar text-cyan-500"></i>
            本周代码提交量
          </h3>
          <div ref="weeklyChartRef" style="width: 100%; height: 320px;"></div>
        </div>

        <div class="bg-white border border-gray-200 rounded-2xl overflow-hidden shadow-sm">
          <div class="max-h-[75vh] overflow-y-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 bg-gray-50 border-b border-gray-200 h-10">
                <th class="px-6 py-4 font-medium">ID</th>
                <th class="px-6 py-4 font-medium">用户</th>
                <th class="px-6 py-4 font-medium">语言</th>
                <th class="px-6 py-4 font-medium">状态</th>
                <th class="px-6 py-4 font-medium">时间</th>
                <th class="px-6 py-4 font-medium">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              <tr 
                v-for="log in filteredLogs" 
                :key="log.id"
                class="hover:bg-gray-50 transition-colors"
              >
                <td class="px-6 py-4 text-gray-500">#{{ log.id }}</td>
                <td class="px-6 py-4 text-gray-900">{{ log.username }}</td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 bg-gray-100 text-gray-600 rounded text-xs">{{ log.language }}</span>
                </td>
                <td class="px-6 py-4">
                  <span 
                    class="px-2 py-1 rounded-full text-xs font-medium"
                    :class="getStatusClass(log.status)"
                  >
                    {{ log.status }}
                  </span>
                </td>
                <td class="px-6 py-4 text-gray-500">{{ formatTime(log.createdAt) }}</td>
                <td class="px-6 py-4">
                  <button 
                    @click="viewSubmission(log)"
                    class="text-cyan-600 hover:text-cyan-500 transition-colors cursor-pointer font-medium"
                  >
                    查看详情
                  </button>
                </td>
              </tr>
            </tbody>
            </table>
          </div>
        </div>
      </div>
      
      <!-- 容器池管理 (Aurora UI Pro Max) -->
      <div v-if="activeTab === 'pool'" class="space-y-6 animate-fade-in">
        <h2 class="text-2xl font-bold flex items-center justify-between text-gray-800">
          <span>容器池监控</span>
          <div class="text-sm font-normal text-gray-500 flex gap-4">
             <span class="flex items-center gap-2 px-3 py-1 rounded-full bg-white border border-gray-200 shadow-sm text-green-600 font-medium">
               <div class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div> 
               实时监控中
             </span>
          </div>
        </h2>
        
        
        <!-- 资源概览卡片 -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
           <!-- 卡片 1: 总容量 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-purple-300 transition-all duration-300 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-purple-500/10 rounded-full blur-xl group-hover:bg-purple-500/20 transition-all duration-500"></div>
              <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">总预热容器</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.total || 0 }}</span>
                    <span class="text-gray-400 mb-1.5 font-medium">/ {{ (statsData.poolStats?.maxPerLang || 3) * Object.keys(statsData.poolStats?.details || {}).length }} (Max)</span>
                 </div>
                 <div class="mt-4 h-1.5 w-full bg-gray-100 rounded-full overflow-hidden">
                    <div class="h-full bg-gradient-to-r from-purple-500 to-indigo-500 rounded-full transition-all duration-500 ease-out"
                         :style="{ width: `${(statsData.poolStats?.total || 0) / ((statsData.poolStats?.maxPerLang || 3) * Object.keys(statsData.poolStats?.details || {}).length || 1) * 100}%` }"></div>
                 </div>
              </div>
           </div>

           <!-- 卡片 2: CPU 限制 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-cyan-300 transition-all duration-300 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-cyan-500/10 rounded-full blur-xl group-hover:bg-cyan-500/20 transition-all duration-500"></div>
               <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">单容器 CPU 限制</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.cpuCores || '2.0' }}</span>
                    <span class="text-cyan-500 mb-1.5 font-medium">Cores</span>
                 </div>
                 <div class="mt-4 flex items-center gap-2 text-xs text-gray-500">
                    <i class="ph ph-cpu text-cyan-500/70"></i>
                    <span>硬限制 (NanoCPUs)</span>
                 </div>
              </div>
           </div>

           <!-- 卡片 3: 内存限制 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-emerald-300 transition-all duration-300 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-emerald-500/10 rounded-full blur-xl group-hover:bg-emerald-500/20 transition-all duration-500"></div>
               <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">单容器内存限制</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.memoryMB || '2048' }}</span>
                    <span class="text-emerald-500 mb-1.5 font-medium">MB</span>
                 </div>
                  <div class="mt-4 flex items-center gap-2 text-xs text-gray-500">
                    <i class="ph ph-memory text-emerald-500/70"></i>
                    <span>物理内存上限</span>
                 </div>
              </div>
           </div>
        </div>

        <!-- 语言池详情 -->
        <div class="mt-8">
           <h3 class="text-lg font-bold text-gray-700 mb-4 flex items-center gap-2">
              <i class="ph ph-squares-four text-gray-400"></i>
              语言池状态
           </h3>
           <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="(detail, lang) in (statsData.poolStats?.details || {})" :key="lang" 
                   class="bg-white border border-gray-200 rounded-xl p-5 hover:bg-gray-50 transition-all duration-300 group shadow-sm">
                 <div class="flex justify-between items-start mb-4">
                    <div class="flex items-center gap-3">
                       <div class="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center text-xl text-gray-700 group-hover:scale-110 transition-transform duration-300 group-hover:bg-gray-200">
                          <i class="ph" :class="getLangIcon(lang)"></i>
                       </div>
                       <div>
                          <h4 class="font-bold text-gray-800 capitalize tracking-wide">{{ lang }}</h4>
                          <p class="text-xs text-gray-500">
                            Pool Capacity: {{ statsData.poolStats?.maxPerLang }}
                          </p>
                       </div>
                    </div>
                    <!-- 状态指示 -->
                    <div class="flex flex-col items-end gap-1.5">
                       <span class="px-2 py-0.5 rounded text-xs font-bold tracking-wide bg-green-100 text-green-600 border border-green-200">
                          {{ detail.idle }} IDLE
                       </span>
                       <span v-if="detail.active > 0" class="px-2 py-0.5 rounded text-xs font-bold tracking-wide bg-blue-100 text-blue-600 border border-blue-200">
                          {{ detail.active }} ACTIVE
                       </span>
                    </div>
                 </div>
                 
                 <!-- 填充率 -->
                 <div class="space-y-1.5">
                    <div class="flex justify-between text-xs font-medium text-gray-500">
                       <span>Pool Usage</span>
                       <span>{{ Math.round((detail.idle + detail.active) / (statsData.poolStats?.maxPerLang || 1) * 100) }}%</span>
                    </div>
                    <div class="h-2 w-full bg-gray-100 rounded-full overflow-hidden">
                       <div class="h-full rounded-full transition-all duration-500 relative"
                            :class="(detail.idle + detail.active) >= (statsData.poolStats?.maxPerLang || 1) ? 'bg-gradient-to-r from-emerald-500 to-green-400' : 'bg-gradient-to-r from-yellow-500 to-orange-400'"
                            :style="{ width: `${(detail.idle + detail.active) / (statsData.poolStats?.maxPerLang || 1) * 100}%` }">
                            <div class="absolute inset-0 bg-white/20 animate-[shimmer_2s_infinite]"></div>
                       </div>
                    </div>
                 </div>
              </div>
           </div>
        </div>
        
        <!-- 管理操作 -->
        <div class="mt-8 border-t border-gray-200 pt-8">
           <div class="bg-red-50 border border-red-100 rounded-xl p-6 flex flex-col md:flex-row items-center justify-between gap-6 hover:bg-red-50/80 transition-colors duration-300">
              <div class="flex gap-4">
                 <div class="p-3 rounded-full bg-red-100 text-red-500 h-fit">
                    <i class="ph ph-warning-octagon text-2xl"></i>
                 </div>
                 <div>
                    <h3 class="text-red-700 font-bold mb-1">危险区域</h3>
                    <p class="text-gray-600 text-sm max-w-lg">
                       强制重置容器池将立即销毁所有 "Idle" 状态的容器并重新进行预热。
                       <br>
                       注意：正在进行的 Active 任务可能不会受到直接影响，但系统负载会短暂升高。
                    </p>
                 </div>
              </div>
              <button 
                @click="handleResetPool"
                :disabled="isResettingPool"
                class="whitespace-nowrap px-6 py-3 bg-red-500/10 text-red-400 border border-red-500/20 rounded-xl font-medium hover:bg-red-500/20 hover:text-red-300 hover:shadow-lg hover:shadow-red-900/10 transition-all cursor-pointer disabled:opacity-50 flex items-center gap-2 active:scale-95"
              >
                <i class="ph ph-arrow-counter-clockwise" :class="{'animate-spin': isResettingPool}"></i>
                {{ isResettingPool ? '重置中...' : '重置容器池' }}
              </button>
           </div>
        </div>
      </div>
    </main>
    
    <!-- 创建用户弹窗 -->
    <div 
      v-if="showCreateUserModal" 
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
      @click.self="showCreateUserModal = false"
    >
      <div class="bg-white border border-gray-200 rounded-2xl p-8 w-full max-w-md shadow-2xl">
        <h3 class="text-xl font-bold mb-6">创建新用户</h3>
        <form @submit.prevent="createUser" class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-2">用户名</label>
            <input 
              v-model="newUser.username"
              type="text"
              required
              class="w-full bg-gray-50 border border-gray-200 rounded-xl px-4 py-3 text-gray-900 focus:outline-none focus:border-cyan-500 transition-colors"
              placeholder="输入用户名"
            />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-2">密码</label>
            <input 
              v-model="newUser.password"
              type="password"
              required
              class="w-full bg-gray-50 border border-gray-200 rounded-xl px-4 py-3 text-gray-900 focus:outline-none focus:border-cyan-500 transition-colors"
              placeholder="输入密码"
            />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-2">角色</label>
            <select 
              v-model="newUser.role"
              class="w-full bg-gray-50 border border-gray-200 rounded-xl px-4 py-3 text-gray-900 focus:outline-none focus:border-cyan-500 transition-colors cursor-pointer"
            >
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div class="flex gap-3 pt-4">
            <button 
              type="button"
              @click="showCreateUserModal = false"
              class="flex-1 px-4 py-3 bg-gray-100 text-gray-600 rounded-xl font-medium hover:bg-gray-200 transition-colors cursor-pointer"
            >
              取消
            </button>
            <button 
              type="submit"
              :disabled="isCreatingUser"
              class="flex-1 px-4 py-3 bg-gradient-to-r from-cyan-500 to-blue-500 text-white rounded-xl font-medium hover:brightness-110 transition-all cursor-pointer disabled:opacity-50"
            >
              {{ isCreatingUser ? '创建中...' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- 提交详情弹窗 -->
    <div 
      v-if="selectedSubmission" 
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-8"
      @click.self="selectedSubmission = null"
    >
      <div class="bg-white border border-gray-200 rounded-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
        <div class="flex items-center justify-between p-6 border-b border-gray-200">
          <div>
            <h3 class="text-xl font-bold">提交详情 #{{ selectedSubmission.id }}</h3>
            <p class="text-sm text-gray-500 mt-1">
              {{ selectedSubmission.username }} · {{ selectedSubmission.language }} · {{ formatTime(selectedSubmission.createdAt) }}
            </p>
          </div>
          <span 
            class="px-3 py-1 rounded-full text-sm font-medium"
            :class="getStatusClass(selectedSubmission.status)"
          >
            {{ selectedSubmission.status }}
          </span>
        </div>
        
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          <div>
            <h4 class="text-sm font-medium text-gray-400 mb-2">源代码</h4>
            <pre class="bg-gray-50 rounded-xl p-4 overflow-x-auto text-sm font-mono text-gray-800 border border-gray-200">{{ selectedSubmission.code }}</pre>
          </div>
          
          <div v-if="selectedSubmission.input">
            <h4 class="text-sm font-medium text-gray-400 mb-2">标准输入</h4>
            <pre class="bg-gray-50 rounded-xl p-4 overflow-x-auto text-sm font-mono text-gray-800 border border-gray-200">{{ selectedSubmission.input }}</pre>
          </div>
          
          <div>
            <h4 class="text-sm font-medium text-gray-400 mb-2">输出结果</h4>
            <pre 
              class="rounded-xl p-4 overflow-x-auto text-sm font-mono whitespace-pre-wrap"
              :class="selectedSubmission.status === 'Failed' || selectedSubmission.status === 'Timeout' 
                ? 'bg-red-50 text-red-600 border border-red-200' 
                : 'bg-gray-50 text-gray-800 border border-gray-200'"
            >{{ selectedSubmission.output || '(无输出)' }}</pre>
          </div>
        </div>
        
        <div class="p-6 border-t border-gray-200">
          <button 
            @click="selectedSubmission = null"
            class="w-full px-4 py-3 bg-gray-100 text-gray-600 rounded-xl font-medium hover:bg-gray-200 transition-colors cursor-pointer"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
    
    <!-- 在线用户弹窗 -->
    <div 
      v-if="showOnlineModal" 
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
      @click.self="showOnlineModal = false"
    >
      <div class="bg-gray-900 border border-gray-800 rounded-2xl p-8 w-full max-w-md">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-xl font-bold flex items-center gap-2">
            <i class="ph ph-users text-green-400"></i>
            在线用户
          </h3>
          <span class="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-sm font-medium">
            {{ statsData.onlineUsers }} 人
          </span>
        </div>
        
        <div class="space-y-3 max-h-[300px] overflow-y-auto">
          <div v-if="statsData.onlineUsers === 0" class="text-center py-8 text-gray-500">
            <i class="ph ph-user-circle text-4xl mb-2 block"></i>
            <p>暂无在线用户</p>
          </div>
          <div v-else class="text-center py-8 text-gray-600">
            <i class="ph ph-pulse text-4xl mb-2 block text-green-500 animate-pulse"></i>
            <p class="mb-2">当前 {{ statsData.onlineUsers }} 位用户在线</p>
            <p class="text-xs text-gray-500">在线状态每 30 秒自动刷新</p>
          </div>
        </div>
        
        <button 
          @click="showOnlineModal = false"
          class="w-full mt-6 px-4 py-3 bg-gray-100 text-gray-600 rounded-xl font-medium hover:bg-gray-200 transition-colors cursor-pointer"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../stores/auth'
import { adminAPI } from '../api/admin'

// ECharts 按需引入
import * as echarts from 'echarts/core'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { BarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
echarts.use([GridComponent, TooltipComponent, BarChart, CanvasRenderer])

const router = useRouter()
const { user, logout } = useAuth()

// 导航标签
const tabs = [
  { id: 'dashboard', label: '仪表盘', icon: 'ph ph-chart-pie-slice' },
  { id: 'users', label: '用户管理', icon: 'ph ph-users' },
  { id: 'logs', label: '运行日志', icon: 'ph ph-terminal' },
  { id: 'pool', label: '容器池', icon: 'ph ph-cube' }
]
const activeTab = ref('dashboard')

// 统计数据
const statsData = ref({
  onlineUsers: 0,
  totalUsers: 0,
  todaySubmissions: 0,
  failedSubmissions: 0,
  successRate: 0,
  poolStats: null
})

const stats = computed(() => [
  {
    id: 'online',
    label: '在线用户',
    value: statsData.value.onlineUsers,
    icon: 'ph ph-users',
    iconColor: 'text-green-500',
    bgColor: 'bg-green-100',
    change: '+实时',
    changeClass: 'bg-green-100 text-green-600',
    action: 'showOnline'
  },
  {
    id: 'users',
    label: '总用户数',
    value: statsData.value.totalUsers,
    icon: 'ph ph-user-circle',
    iconColor: 'text-blue-500',
    bgColor: 'bg-blue-100',
    change: '累计',
    changeClass: 'bg-blue-100 text-blue-600',
    action: 'goUsers'
  },
  {
    id: 'submissions',
    label: '今日提交',
    value: statsData.value.todaySubmissions,
    icon: 'ph ph-code',
    iconColor: 'text-purple-500',
    bgColor: 'bg-purple-100',
    change: '今日',
    changeClass: 'bg-purple-100 text-purple-600',
    action: 'goLogs'
  },
  {
    id: 'success',
    label: '成功率',
    value: statsData.value.successRate.toFixed(1) + '%',
    icon: 'ph ph-check-circle',
    iconColor: 'text-cyan-500',
    bgColor: 'bg-cyan-100',
    change: statsData.value.failedSubmissions + ' 失败',
    changeClass: statsData.value.failedSubmissions > 0 ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-500',
    action: 'goFailed'
  }
])

// 用户数据
const users = ref([])
const showCreateUserModal = ref(false)
const newUser = reactive({ username: '', password: '', role: 'user' })
const isCreatingUser = ref(false)

// 日志数据
const submissions = ref([])
const logFilter = reactive({ status: '', dateFilter: '' }) // dateFilter: '' | 'today'
const selectedSubmission = ref(null)

const recentSubmissions = computed(() => submissions.value.slice(0, 20))

// 判断是否是今天
const isToday = (dateStr) => {
  const date = new Date(dateStr)
  const today = new Date()
  return date.getFullYear() === today.getFullYear() &&
         date.getMonth() === today.getMonth() &&
         date.getDate() === today.getDate()
}

const filteredLogs = computed(() => {
  let result = submissions.value
  
  // 日期筛选
  if (logFilter.dateFilter === 'today') {
    result = result.filter(s => isToday(s.createdAt))
  }
  
  // 状态筛选
  if (logFilter.status) {
    result = result.filter(s => s.status === logFilter.status)
  }
  
  return result
})

// 方法
const formatTime = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

const getStatusClass = (status) => {
  const classes = {
    'Success': 'bg-green-100 text-green-700 border border-green-200',
    'Failed': 'bg-red-100 text-red-700 border border-red-200',
    'Timeout': 'bg-orange-100 text-orange-700 border border-orange-200',
    'Pending': 'bg-gray-100 text-gray-600 border border-gray-200',
    'Running': 'bg-blue-100 text-blue-700 border border-blue-200'
  }
  return classes[status] || 'bg-gray-100 text-gray-600'
}

const viewSubmission = (sub) => {
  selectedSubmission.value = sub
}

const handleLogout = () => {
  logout()
  router.push('/login')
}

// 点击统计卡片的处理函数
const showOnlineModal = ref(false)

const handleStatClick = (stat) => {
  switch (stat.action) {
    case 'showOnline':
      showOnlineModal.value = true
      break
    case 'goUsers':
      activeTab.value = 'users'
      break
    case 'goLogs':
      // 今日提交：筛选今天的记录
      logFilter.status = ''
      logFilter.dateFilter = 'today'
      activeTab.value = 'logs'
      break
    case 'goFailed':
      // 失败记录：显示所有失败的
      logFilter.status = 'Failed'
      logFilter.dateFilter = ''
      activeTab.value = 'logs'
      break
  }
}

const getLangIcon = (lang) => {
  const map = {
    'python': 'ph-file-code',
    'go': 'ph-file-code',
    'cpp': 'ph-file-c',
    'java': 'ph-coffee',
    'javascript': 'ph-file-js'
  }
  return map[lang.toLowerCase()] || 'ph-code'
}

const loadStats = async () => {
  try {
    const data = await adminAPI.getStats()
    statsData.value = data
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
}

const loadUsers = async () => {
  try {
    const data = await adminAPI.getUsers()
    users.value = data.users || []
  } catch (e) {
    console.error('Failed to load users:', e)
  }
}

const loadSubmissions = async () => {
  try {
    const data = await adminAPI.getSubmissions()
    submissions.value = data.submissions || []
  } catch (e) {
    console.error('Failed to load submissions:', e)
  }
}

const createUser = async () => {
  isCreatingUser.value = true
  try {
    await adminAPI.createUser(newUser)
    showCreateUserModal.value = false
    newUser.username = ''
    newUser.password = ''
    newUser.role = 'user'
    await loadUsers()
  } catch (e) {
    alert('创建用户失败: ' + (e.message || '未知错误'))
  } finally {
    isCreatingUser.value = false
  }
}

const confirmDeleteUser = async (u) => {
  if (!confirm(`确定要删除用户 "${u.username}" 吗？此操作将同时删除该用户的所有提交记录。`)) return
  try {
    await adminAPI.deleteUser(u.id)
    await loadUsers()
  } catch (e) {
    alert('删除用户失败: ' + (e.message || '未知错误'))
  }
}

const isResettingPool = ref(false)

const handleResetPool = async () => {
  if (!confirm('确定要重置容器池吗？这将销毁所有闲置容器。')) return
  
  isResettingPool.value = true
  try {
    const res = await adminAPI.resetPool()
    alert(`成功: ${res.message} (清理了 ${res.removed} 个容器)`)
  } catch (e) {
    alert('重置失败: ' + (e.message || '未知错误'))
  } finally {
    isResettingPool.value = false
  }
}

// 周提交量图表
const showWeeklyChart = ref(false)
const weeklyChartRef = ref(null)
let weeklyChartInstance = null

const toggleWeeklyChart = async () => {
  showWeeklyChart.value = !showWeeklyChart.value
  if (showWeeklyChart.value) {
    await nextTick()
    await initWeeklyChart()
  } else {
    disposeWeeklyChart()
  }
}

const initWeeklyChart = async () => {
  if (!weeklyChartRef.value) return
  
  // 销毁旧实例
  disposeWeeklyChart()
  
  weeklyChartInstance = echarts.init(weeklyChartRef.value)

  try {
    const data = await adminAPI.getWeeklyStats()
    const option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#e5e7eb',
        textStyle: { color: '#374151' }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '8%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: data.labels,
        axisLine: { lineStyle: { color: '#e5e7eb' } },
        axisLabel: { color: '#6b7280', fontSize: 13 }
      },
      yAxis: {
        type: 'value',
        minInterval: 1,
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { lineStyle: { color: '#f3f4f6' } },
        axisLabel: { color: '#9ca3af' }
      },
      series: [{
        data: data.data,
        type: 'bar',
        barWidth: '45%',
        itemStyle: {
          borderRadius: [6, 6, 0, 0],
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#06b6d4' },
            { offset: 1, color: '#3b82f6' }
          ])
        },
        emphasis: {
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: '#22d3ee' },
              { offset: 1, color: '#60a5fa' }
            ])
          }
        }
      }]
    }
    weeklyChartInstance.setOption(option)
  } catch (e) {
    console.error('Failed to load weekly stats:', e)
  }
}

const disposeWeeklyChart = () => {
  if (weeklyChartInstance) {
    weeklyChartInstance.dispose()
    weeklyChartInstance = null
  }
}

// 窗口 resize 自适应
const handleResize = () => {
  weeklyChartInstance?.resize()
}

onMounted(() => {
  loadStats()
  loadUsers()
  loadSubmissions()
  
  // 定时刷新统计数据
  setInterval(loadStats, 30000)

  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  disposeWeeklyChart()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap');

* {
  font-family: 'Fira Sans', sans-serif;
}

pre, code {
  font-family: 'Fira Code', monospace;
}
</style>
