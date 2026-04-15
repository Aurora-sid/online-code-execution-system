<template>
  <div class="h-screen overflow-y-auto bg-gray-50 text-gray-900 font-sans">
    <!-- 顶部导航栏 -->
    <header class="fixed top-4 left-4 right-4 z-50 bg-white/90 backdrop-blur-md border border-gray-200 rounded-2xl shadow-sm">
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
            class="px-4 py-2 rounded-lg text-sm font-medium transition-colors duration-200 cursor-pointer"
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
      <div v-show="activeTab === 'dashboard'" class="space-y-6 tab-content">
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
      <div v-show="activeTab === 'users'" class="space-y-6 tab-content">
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
      <div v-show="activeTab === 'logs'" class="space-y-6 tab-content">
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

        <!-- 后端服务器日志终端 -->
        <div class="mt-6 bg-gray-900 border border-gray-700 rounded-2xl overflow-hidden shadow-lg">
           <!-- 终端标题栏 -->
           <div class="flex items-center justify-between px-5 py-3 bg-gray-800 border-b border-gray-700">
              <div class="flex items-center gap-3">
                 <div class="flex gap-1.5">
                    <div class="w-3 h-3 rounded-full bg-red-500"></div>
                    <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
                    <div class="w-3 h-3 rounded-full bg-green-500"></div>
                 </div>
                 <span class="text-gray-400 text-sm font-mono font-medium">server-logs — code-exec backend</span>
                 <span class="text-gray-600 text-xs">({{ serverLogs.length }} 条)</span>
              </div>
              <div class="flex items-center gap-2">
                 <button 
                    @click="loadServerLogs"
                    :disabled="serverLogsLoading"
                    class="px-3 py-1.5 bg-gray-700 text-gray-300 text-xs rounded-lg hover:bg-gray-600 transition-colors cursor-pointer flex items-center gap-1 disabled:opacity-50"
                 >
                    <i class="ph ph-arrows-clockwise" :class="serverLogsLoading ? 'animate-spin' : ''"></i>
                    刷新
                 </button>
                 <button 
                    @click="clearServerLogs"
                    class="px-3 py-1.5 bg-gray-700 text-gray-300 text-xs rounded-lg hover:bg-red-600/80 transition-colors cursor-pointer flex items-center gap-1"
                 >
                    <i class="ph ph-trash"></i>
                    清空
                 </button>
              </div>
           </div>
           <!-- 日志内容区 -->
           <div ref="serverLogContainer" class="p-4 max-h-[50vh] overflow-y-auto scrollbar-thin font-mono text-sm leading-relaxed">
              <div v-if="serverLogs.length === 0" class="text-gray-600 text-center py-8">
                 <i class="ph ph-terminal text-3xl block mb-2"></i>
                 <p>暂无日志</p>
              </div>
              <div 
                 v-for="(log, index) in serverLogs" 
                 :key="index"
                 class="flex gap-3 py-0.5 hover:bg-gray-800/50 rounded px-2 -mx-2 transition-colors"
              >
                 <span class="text-gray-600 shrink-0 select-none text-xs leading-6">{{ log.time }}</span>
                 <span 
                    class="break-all"
                    :class="getLogColor(log.message)"
                 >{{ log.message }}</span>
              </div>
           </div>
        </div>
      </div>
      
      <!-- 容器池管理 (Aurora UI Pro Max) -->
      <div v-show="activeTab === 'pool'" class="space-y-6 tab-content">
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
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
           <!-- 卡片 1: 总容器 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-purple-300 transition-colors duration-200 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-purple-500/10 rounded-full blur-xl group-hover:bg-purple-500/20 transition-colors duration-300"></div>
              <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">活跃容器总数</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.total || 0 }}</span>
                 </div>
                 <div class="mt-3 flex items-center gap-2 text-xs text-gray-400">
                    <i class="ph ph-stack text-purple-400"></i>
                    <span>所有语言容器合计</span>
                 </div>
              </div>
           </div>

           <!-- 卡片 2: 动态水位线 (EMA Predictor) -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-amber-300 transition-colors duration-200 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-amber-500/10 rounded-full blur-xl group-hover:bg-amber-500/20 transition-colors duration-300"></div>
              <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">动态水位线</p>
                 <div class="flex items-end gap-3">
                    <div class="flex flex-col items-center">
                       <span class="text-xs text-blue-500 font-medium">动态 LWM</span>
                       <span class="text-3xl font-bold text-gray-900">{{ statsData.poolStats?.dynamicLWM ?? statsData.poolStats?.lwm ?? 2 }}</span>
                    </div>
                    <span class="text-gray-300 mb-1 text-xl">/</span>
                    <div class="flex flex-col items-center">
                       <span class="text-xs text-red-500 font-medium">HWM</span>
                       <span class="text-3xl font-bold text-gray-900">{{ statsData.poolStats?.hwm ?? 8 }}</span>
                    </div>
                 </div>
                 <div class="mt-3 flex items-center gap-2 text-xs text-gray-400">
                    <i class="ph ph-waves text-amber-400"></i>
                    <span>EMA 自适应 · 实时弹性调度</span>
                 </div>
              </div>
           </div>

           <!-- 卡片 3: CPU 限制 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-cyan-300 transition-colors duration-200 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-cyan-500/10 rounded-full blur-xl group-hover:bg-cyan-500/20 transition-colors duration-300"></div>
               <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">单容器 CPU 限制</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.cpuCores || '2.0' }}</span>
                    <span class="text-cyan-500 mb-1.5 font-medium">Cores</span>
                 </div>
                 <div class="mt-3 flex items-center gap-2 text-xs text-gray-400">
                    <i class="ph ph-cpu text-cyan-400"></i>
                    <span>硬限制 (NanoCPUs)</span>
                 </div>
              </div>
           </div>

           <!-- 卡片 4: 内存限制 -->
           <div class="bg-white border border-gray-200 rounded-2xl p-6 relative overflow-hidden group hover:border-emerald-300 transition-colors duration-200 shadow-sm">
              <div class="absolute -right-4 -top-4 w-24 h-24 bg-emerald-500/10 rounded-full blur-xl group-hover:bg-emerald-500/20 transition-colors duration-300"></div>
               <div class="relative z-10">
                 <p class="text-gray-500 text-sm mb-1 font-medium">单容器内存限制</p>
                 <div class="flex items-end gap-2">
                    <span class="text-4xl font-bold text-gray-900 tracking-tight">{{ statsData.poolStats?.memoryMB || '2048' }}</span>
                    <span class="text-emerald-500 mb-1.5 font-medium">MB</span>
                 </div>
                  <div class="mt-3 flex items-center gap-2 text-xs text-gray-400">
                    <i class="ph ph-memory text-emerald-400"></i>
                    <span>物理内存上限</span>
                 </div>
              </div>
           </div>
        </div>

        <!-- 流量预测与动态水位线 -->
        <div class="mt-6 bg-white border border-gray-200 rounded-2xl p-6 shadow-sm">
           <h3 class="text-lg font-bold text-gray-700 mb-5 flex items-center gap-2">
              <i class="ph ph-chart-line-up text-indigo-500"></i>
              流量预测与动态水位线
              <span class="text-xs font-normal text-gray-400 ml-2">· EMA 指数移动平均算法</span>
           </h3>
           <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <!-- 当前 RPS -->
              <div class="bg-gradient-to-br from-indigo-50 to-blue-50 border border-indigo-100 rounded-xl p-4">
                 <p class="text-xs text-indigo-500 font-semibold mb-1 tracking-wide uppercase">当前 RPS</p>
                 <p class="text-3xl font-bold text-gray-900">{{ statsData.poolStats?.predictorStats?.lastRPS ?? 0 }}</p>
                 <p class="text-xs text-gray-400 mt-1">本周期请求数</p>
              </div>
              <!-- EMA 平滑值 -->
              <div class="bg-gradient-to-br from-purple-50 to-fuchsia-50 border border-purple-100 rounded-xl p-4">
                 <p class="text-xs text-purple-500 font-semibold mb-1 tracking-wide uppercase">EMA 平滑值</p>
                 <p class="text-3xl font-bold text-gray-900">{{ (statsData.poolStats?.predictorStats?.currentEMA ?? 0).toFixed(2) }}</p>
                 <p class="text-xs text-gray-400 mt-1">指数移动平均</p>
              </div>
              <!-- 加速度 -->
              <div class="bg-gradient-to-br rounded-xl p-4 border"
                   :class="predictorAcceleration > 2 ? 'from-red-50 to-orange-50 border-red-100' : predictorAcceleration < -2 ? 'from-teal-50 to-cyan-50 border-teal-100' : 'from-gray-50 to-slate-50 border-gray-100'">
                 <p class="text-xs font-semibold mb-1 tracking-wide uppercase"
                    :class="predictorAcceleration > 2 ? 'text-red-500' : predictorAcceleration < -2 ? 'text-teal-500' : 'text-gray-500'">流量加速度</p>
                 <p class="text-3xl font-bold text-gray-900 flex items-center gap-1">
                    <i class="ph text-lg" :class="predictorAcceleration > 2 ? 'ph-trend-up text-red-500' : predictorAcceleration < -2 ? 'ph-trend-down text-teal-500' : 'ph-minus text-gray-400'"></i>
                    {{ predictorAcceleration.toFixed(2) }}
                 </p>
                 <p class="text-xs text-gray-400 mt-1">{{ predictorAcceleration > 5 ? '⚡ 流量暴增' : predictorAcceleration > 2 ? '📈 快速上升' : predictorAcceleration < -2 ? '📉 快速回落' : '🟢 平稳' }}</p>
              </div>
              <!-- 动态 LWM 刻度 -->
              <div class="bg-gradient-to-br from-amber-50 to-yellow-50 border border-amber-100 rounded-xl p-4">
                 <p class="text-xs text-amber-600 font-semibold mb-1 tracking-wide uppercase">动态 LWM</p>
                 <p class="text-3xl font-bold text-gray-900">{{ statsData.poolStats?.dynamicLWM ?? 0 }}</p>
                 <div class="mt-2 relative h-2 w-full bg-gray-200 rounded-full overflow-hidden">
                    <div class="h-full bg-gradient-to-r from-amber-400 to-orange-400 rounded-full transition-all duration-500"
                         :style="{ width: `${dynamicLWMPercent}%` }"></div>
                 </div>
                 <p class="text-xs text-gray-400 mt-1">范围 {{ predictorLWMMin }} ~ {{ predictorLWMMax }}</p>
              </div>
           </div>
        </div>

        <!-- 语言池详情 -->
        <div class="mt-8">
           <h3 class="text-lg font-bold text-gray-700 mb-4 flex items-center gap-2">
              <i class="ph ph-squares-four text-gray-400"></i>
              语言池状态
              <span class="text-xs font-normal text-gray-400 ml-2">· 监控间隔 {{ statsData.poolStats?.monitorInterval || 3 }}s</span>
           </h3>
           <div class="max-h-[400px] overflow-y-auto rounded-xl pr-1 scrollbar-thin">
           <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="(detail, lang) in (statsData.poolStats?.details || {})" :key="lang" 
                   class="bg-white border rounded-xl p-5 hover:bg-gray-50 transition-colors duration-200 group shadow-sm"
                   :class="getPoolCardBorderClass(detail.state)">
                 <div class="flex justify-between items-start mb-4">
                    <div class="flex items-center gap-3">
                       <div class="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center overflow-hidden group-hover:scale-110 transition-transform duration-300 group-hover:bg-gray-200">
                          <img :src="getLangIconSrc(lang)" :alt="lang" class="w-7 h-7 object-contain" />
                       </div>
                       <div>
                          <h4 class="font-bold text-gray-800 capitalize tracking-wide">{{ lang }}</h4>
                          <p class="text-xs text-gray-500">
                            LWM {{ statsData.poolStats?.lwm }} · HWM {{ statsData.poolStats?.hwm }}
                          </p>
                       </div>
                    </div>
                    <!-- 水位状态 -->
                    <div class="flex flex-col items-end gap-1.5">
                       <span class="px-2.5 py-1 rounded-full text-xs font-bold tracking-wide"
                             :class="getStateClass(detail.state)">
                          {{ getStateLabel(detail.state) }}
                       </span>
                       <div class="flex gap-1.5">
                         <span class="px-2 py-0.5 rounded text-xs font-bold tracking-wide bg-green-100 text-green-600 border border-green-200">
                            {{ detail.idle }} IDLE
                         </span>
                         <span v-if="detail.active > 0" class="px-2 py-0.5 rounded text-xs font-bold tracking-wide bg-blue-100 text-blue-600 border border-blue-200">
                            {{ detail.active }} ACTIVE
                         </span>
                       </div>
                    </div>
                 </div>
                 
                 <!-- 水位刻度条 -->
                 <div class="space-y-1.5">
                    <div class="flex justify-between text-xs font-medium text-gray-500">
                       <span>水位 (Idle: {{ detail.idle }})</span>
                       <span>{{ detail.idle }} / {{ statsData.poolStats?.hwm || 8 }}</span>
                    </div>
                    <div class="relative h-2.5 w-full bg-gray-100 rounded-full overflow-hidden">
                       <!-- LWM 刻度标记 -->
                       <div class="absolute top-0 bottom-0 w-px bg-blue-400 z-10 opacity-60"
                            :style="{ left: `${(statsData.poolStats?.lwm || 2) / (statsData.poolStats?.hwm || 8) * 100}%` }"></div>
                       <!-- 水位填充 -->
                       <div class="h-full rounded-full transition-all duration-500 relative"
                            :class="getWaterBarClass(detail)"
                            :style="{ width: `${Math.min(detail.idle / (statsData.poolStats?.hwm || 8) * 100, 100)}%` }">
                            <div class="absolute inset-0 bg-white/20 animate-[shimmer_2s_infinite]"></div>
                       </div>
                    </div>
                    <!-- LWM / HWM 标注 -->
                    <div class="flex justify-between text-[10px] text-gray-400">
                       <span>0</span>
                       <span class="text-blue-400 font-medium" :style="{ marginLeft: `${Math.max(0, (statsData.poolStats?.lwm || 2) / (statsData.poolStats?.hwm || 8) * 100 - 8)}%` }">LWM</span>
                       <span class="text-red-400 font-medium">HWM</span>
                    </div>
                 </div>
              </div>
           </div>
           </div>
        </div>
        
        <!-- 管理操作 -->
        <div class="mt-8 border-t border-gray-200 pt-8">
           <div class="bg-red-50 border border-red-100 rounded-xl p-6 flex flex-col md:flex-row items-center justify-between gap-6 hover:bg-red-50/80 transition-colors duration-200">
              <div class="flex gap-4">
                 <div class="p-3 rounded-full bg-red-100 text-red-500 h-fit">
                    <i class="ph ph-warning-octagon text-2xl"></i>
                 </div>
                 <div>
                    <h3 class="text-red-700 font-bold mb-1">危险区域</h3>
                    <p class="text-gray-600 text-sm max-w-lg">
                       强制重置容器池将立即销毁所有 "Idle" 状态的容器。弹性调度监控协程将自动根据 LWM 重新预热。
                       <br>
                       注意：正在进行的 Active 任务不受影响，但系统负载会短暂升高。
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

      <!-- AI 大模型管理 -->
      <div v-show="activeTab === 'llm'" class="space-y-6 tab-content">
        <!-- 加载状态 -->
        <div v-if="llmLoading" class="flex items-center justify-center py-16">
          <div class="flex items-center gap-3 text-gray-500">
            <i class="ph ph-spinner text-2xl animate-spin"></i>
            <span>加载中...</span>
          </div>
        </div>

        <template v-else>
          <!-- 状态概览卡片 -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
            <!-- 启用状态卡片 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6 hover:shadow-md transition-all">
              <div class="flex items-center justify-between mb-4">
                <div class="w-12 h-12 rounded-xl flex items-center justify-center" :class="llmStatus.enabled ? 'bg-green-100' : 'bg-red-100'">
                  <i class="ph text-xl" :class="llmStatus.enabled ? 'ph-check-circle text-green-600' : 'ph-x-circle text-red-500'"></i>
                </div>
                <span class="px-3 py-1 rounded-full text-xs font-semibold"
                  :class="llmStatus.enabled ? 'bg-green-100 text-green-700 border border-green-200' : 'bg-red-100 text-red-600 border border-red-200'">
                  {{ llmStatus.enabled ? '运行中' : '已禁用' }}
                </span>
              </div>
              <p class="text-sm text-gray-500 mb-1">服务状态</p>
              <p class="text-2xl font-bold" :class="llmStatus.enabled ? 'text-green-600' : 'text-red-500'">{{ llmStatus.enabled ? '已启用' : '已禁用' }}</p>
            </div>

            <!-- 当前模型卡片 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6 hover:shadow-md transition-all">
              <div class="flex items-center justify-between mb-4">
                <div class="w-12 h-12 rounded-xl bg-blue-100 flex items-center justify-center">
                  <i class="ph ph-brain text-xl text-blue-600"></i>
                </div>
                <span class="px-3 py-1 rounded-full text-xs font-semibold bg-blue-100 text-blue-700 border border-blue-200">
                  {{ llmStatus.apiKeySet ? 'Key 已配置' : 'Key 未配置' }}
                </span>
              </div>
              <p class="text-sm text-gray-500 mb-1">当前模型</p>
              <p class="text-lg font-bold text-gray-900 truncate" :title="llmStatus.model">{{ llmStatus.model || '—' }}</p>
            </div>

            <!-- 调用次数卡片 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6 hover:shadow-md transition-all">
              <div class="flex items-center justify-between mb-4">
                <div class="w-12 h-12 rounded-xl bg-purple-100 flex items-center justify-center">
                  <i class="ph ph-chart-line-up text-xl text-purple-600"></i>
                </div>
                <span class="px-3 py-1 rounded-full text-xs font-semibold bg-purple-100 text-purple-700 border border-purple-200">
                  ✓{{ llmStatus.successCalls }} / ✗{{ llmStatus.failedCalls }}
                </span>
              </div>
              <p class="text-sm text-gray-500 mb-1">总调用次数</p>
              <p class="text-2xl font-bold text-gray-900">{{ llmStatus.totalCalls }}</p>
            </div>

            <!-- 最近调用卡片 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6 hover:shadow-md transition-all">
              <div class="flex items-center justify-between mb-4">
                <div class="w-12 h-12 rounded-xl bg-amber-100 flex items-center justify-center">
                  <i class="ph ph-clock text-xl text-amber-600"></i>
                </div>
              </div>
              <p class="text-sm text-gray-500 mb-1">最近调用</p>
              <p class="text-lg font-bold text-gray-900">{{ llmStatus.lastCallAt ? formatTime(llmStatus.lastCallAt) : '暂无记录' }}</p>
            </div>
          </div>

          <!-- 操作面板 -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 左侧：启停 & 模型管理 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6">
              <h3 class="text-lg font-bold mb-6 flex items-center gap-2">
                <i class="ph ph-sliders text-cyan-500"></i>
                功能控制
              </h3>

              <!-- 启停开关 -->
              <div class="flex items-center justify-between p-4 bg-gray-50 rounded-xl mb-5">
                <div>
                  <p class="font-medium text-gray-900">AI 分析功能</p>
                  <p class="text-sm text-gray-500 mt-0.5">控制全局 AI 代码分析功能的启停</p>
                </div>
                <button
                  @click="handleToggleLLM"
                  :disabled="llmToggling"
                  class="relative inline-flex h-8 w-16 items-center rounded-full transition-colors duration-300 focus:outline-none cursor-pointer disabled:opacity-50"
                  :class="llmStatus.enabled ? 'bg-green-500' : 'bg-gray-300'"
                >
                  <span
                    class="inline-block h-6 w-6 transform rounded-full bg-white shadow-md transition-transform duration-300"
                    :class="llmStatus.enabled ? 'translate-x-9' : 'translate-x-1'"
                  />
                </button>
              </div>

              <!-- 模型选择 -->
              <div class="space-y-3">
                <p class="font-medium text-gray-900">模型选择</p>
                <div class="flex gap-2">
                  <select
                    v-model="selectedModel"
                    class="flex-1 bg-gray-50 border border-gray-200 rounded-xl px-4 py-3 text-gray-900 focus:outline-none focus:border-cyan-500 transition-colors cursor-pointer text-sm"
                  >
                    <option v-for="m in presetModels" :key="m" :value="m">{{ m }}</option>
                  </select>
                  <button
                    @click="handleChangeModel"
                    :disabled="llmModelChanging || selectedModel === llmStatus.model"
                    class="px-5 py-3 bg-gradient-to-r from-cyan-500 to-blue-500 text-white rounded-xl font-medium hover:brightness-110 transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 text-sm"
                  >
                    <i class="ph ph-swap" :class="{'animate-spin': llmModelChanging}"></i>
                    {{ llmModelChanging ? '切换中...' : '应用' }}
                  </button>
                </div>
                <p class="text-xs text-gray-400">切换模型后立即生效，无需重启服务</p>
              </div>
            </div>

            <!-- 右侧：使用统计 -->
            <div class="bg-white border border-gray-200 rounded-2xl p-6">
              <h3 class="text-lg font-bold mb-6 flex items-center gap-2">
                <i class="ph ph-chart-pie text-purple-500"></i>
                使用统计
              </h3>

              <!-- 成功率环形进度 -->
              <div class="flex items-center gap-8 mb-6">
                <div class="relative w-28 h-28">
                  <svg class="w-28 h-28 -rotate-90" viewBox="0 0 120 120">
                    <circle cx="60" cy="60" r="52" fill="none" stroke="#f3f4f6" stroke-width="10"/>
                    <circle cx="60" cy="60" r="52" fill="none"
                      :stroke="llmSuccessRate >= 90 ? '#22c55e' : llmSuccessRate >= 60 ? '#f59e0b' : '#ef4444'"
                      stroke-width="10"
                      stroke-linecap="round"
                      :stroke-dasharray="`${llmSuccessRate * 3.267} 326.7`"
                      class="transition-all duration-700"
                    />
                  </svg>
                  <div class="absolute inset-0 flex items-center justify-center">
                    <span class="text-2xl font-bold text-gray-900">{{ llmSuccessRate }}%</span>
                  </div>
                </div>
                <div class="space-y-3 flex-1">
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-gray-500">成功</span>
                    <span class="text-sm font-semibold text-green-600">{{ llmStatus.successCalls }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-gray-500">失败</span>
                    <span class="text-sm font-semibold text-red-500">{{ llmStatus.failedCalls }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-gray-500">总计</span>
                    <span class="text-sm font-semibold text-gray-900">{{ llmStatus.totalCalls }}</span>
                  </div>
                </div>
              </div>

              <!-- API 地址 -->
              <div class="p-4 bg-gray-50 rounded-xl mb-4">
                <p class="text-xs text-gray-400 mb-1">API 地址</p>
                <p class="text-sm text-gray-700 font-mono truncate" :title="llmStatus.apiUrl">{{ llmStatus.apiUrl || '—' }}</p>
              </div>

              <!-- 危险区域：重置统计 -->
              <div class="border border-red-200 rounded-xl p-4 bg-red-50/50">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-sm font-medium text-red-700">重置统计数据</p>
                    <p class="text-xs text-red-500/70 mt-0.5">此操作不可恢复</p>
                  </div>
                  <button
                    @click="handleResetLLMStats"
                    :disabled="llmStatsResetting"
                    class="px-4 py-2 bg-red-500/10 text-red-600 border border-red-200 rounded-lg text-sm font-medium hover:bg-red-500/20 transition-colors cursor-pointer disabled:opacity-50 flex items-center gap-1.5"
                  >
                    <i class="ph ph-trash" :class="{'animate-spin': llmStatsResetting}"></i>
                    {{ llmStatsResetting ? '重置中...' : '重置' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </template>
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
  { id: 'pool', label: '容器池', icon: 'ph ph-cube' },
  { id: 'llm', label: 'AI 大模型', icon: 'ph ph-brain' }
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

// 语言图标映射（使用 assets/icons 中的实际图标）
import iconC from '@/assets/icons/C.webp'
import iconCpp from '@/assets/icons/C++.webp'
import iconGo from '@/assets/icons/go.webp'
import iconJava from '@/assets/icons/java.webp'
import iconJS from '@/assets/icons/node_js.webp'
import iconPython from '@/assets/icons/Python.webp'
import iconRust from '@/assets/icons/Rust.webp'
import iconTS from '@/assets/icons/typescript.webp'

const langIconMap = {
  'c': iconC,
  'cpp': iconCpp,
  'go': iconGo,
  'java': iconJava,
  'javascript': iconJS,
  'python': iconPython,
  'rust': iconRust,
  'typescript': iconTS,
}

const getLangIconSrc = (lang) => {
  return langIconMap[lang.toLowerCase()] || ''
}

// 水位状态标签样式
const getStateClass = (state) => {
  const map = {
    'steady':   'bg-green-100 text-green-700 border border-green-200',
    'prewarm':  'bg-amber-100 text-amber-700 border border-amber-200',
    'burst':    'bg-red-100 text-red-700 border border-red-200 animate-pulse',
    'recovery': 'bg-blue-100 text-blue-700 border border-blue-200',
  }
  return map[state] || 'bg-gray-100 text-gray-600 border border-gray-200'
}

// 水位状态中文标签
const getStateLabel = (state) => {
  const map = {
    'steady':   '✅ 平稳',
    'prewarm':  '🔥 预热中',
    'burst':    '⚡ 爆发',
    'recovery': '♻️ 退水',
  }
  return map[state] || state
}

// 语言池卡片边框颜色（基于水位状态）
const getPoolCardBorderClass = (state) => {
  const map = {
    'steady':   'border-gray-200 hover:border-green-300',
    'prewarm':  'border-amber-200 hover:border-amber-400',
    'burst':    'border-red-300 hover:border-red-400',
    'recovery': 'border-blue-200 hover:border-blue-300',
  }
  return map[state] || 'border-gray-200'
}

// 水位进度条颜色（基于水位状态）
const getWaterBarClass = (detail) => {
  if (detail.state === 'burst') return 'bg-gradient-to-r from-red-500 to-rose-400'
  if (detail.state === 'prewarm') return 'bg-gradient-to-r from-amber-500 to-yellow-400'
  if (detail.state === 'recovery') return 'bg-gradient-to-r from-blue-500 to-cyan-400'
  return 'bg-gradient-to-r from-emerald-500 to-green-400'
}

// EMA 流量预测器计算属性
const predictorAcceleration = computed(() => {
  return statsData.value.poolStats?.predictorStats?.acceleration ?? 0
})

const predictorLWMMin = computed(() => {
  // 使用后端返回的区间，后端未提供时使用合理默认值
  return 1
})

const predictorLWMMax = computed(() => {
  // LWM 上界近似为 HWM 的一半或至少 LWM + 5
  const hwm = statsData.value.poolStats?.hwm ?? 8
  return Math.max(hwm - 2, 3)
})

const dynamicLWMPercent = computed(() => {
  const current = statsData.value.poolStats?.dynamicLWM ?? 0
  const min = predictorLWMMin.value
  const max = predictorLWMMax.value
  if (max <= min) return 0
  return Math.min(100, Math.max(0, ((current - min) / (max - min)) * 100))
})

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

// ==================== LLM 大模型管理 ====================

const llmStatus = ref({
  enabled: false,
  model: '',
  apiUrl: '',
  apiKeySet: false,
  totalCalls: 0,
  successCalls: 0,
  failedCalls: 0,
  lastCallAt: null
})
const llmLoading = ref(false)
const llmToggling = ref(false)
const llmModelChanging = ref(false)
const llmStatsResetting = ref(false)

const presetModels = [
  'codegeex-4',
  'glm-4-flashx',
  'glm-4-flash',
  'glm-4',
  'glm-4-plus'
]

// 初始化 selectedModel 为当前模型
const selectedModel = ref('')

const llmSuccessRate = computed(() => {
  if (llmStatus.value.totalCalls === 0) return 0
  return Math.round((llmStatus.value.successCalls / llmStatus.value.totalCalls) * 100)
})

const loadLLMStatus = async () => {
  llmLoading.value = true
  try {
    const data = await adminAPI.getLLMStatus()
    llmStatus.value = data
    if (!selectedModel.value || selectedModel.value === '') {
      selectedModel.value = data.model
    }
  } catch (e) {
    console.error('Failed to load LLM status:', e)
  } finally {
    llmLoading.value = false
  }
}

const handleToggleLLM = async () => {
  const newState = !llmStatus.value.enabled
  const action = newState ? '启用' : '禁用'
  if (!confirm(`确定要${action} AI 分析功能吗？`)) return

  llmToggling.value = true
  try {
    await adminAPI.toggleLLM(newState)
    llmStatus.value.enabled = newState
  } catch (e) {
    alert(`${action}失败: ` + (e.message || '未知错误'))
  } finally {
    llmToggling.value = false
  }
}

const handleChangeModel = async () => {
  if (selectedModel.value === llmStatus.value.model) return
  llmModelChanging.value = true
  try {
    await adminAPI.setLLMModel(selectedModel.value)
    llmStatus.value.model = selectedModel.value
  } catch (e) {
    alert('切换模型失败: ' + (e.message || '未知错误'))
    selectedModel.value = llmStatus.value.model
  } finally {
    llmModelChanging.value = false
  }
}

const handleResetLLMStats = async () => {
  if (!confirm('确定要重置 LLM 使用统计吗？此操作不可恢复。')) return
  llmStatsResetting.value = true
  try {
    await adminAPI.resetLLMStats()
    llmStatus.value.totalCalls = 0
    llmStatus.value.successCalls = 0
    llmStatus.value.failedCalls = 0
    llmStatus.value.lastCallAt = null
  } catch (e) {
    alert('重置失败: ' + (e.message || '未知错误'))
  } finally {
    llmStatsResetting.value = false
  }
}

// ==================== 后端日志查看 ====================

const serverLogs = ref([])
const serverLogsLoading = ref(false)
const serverLogContainer = ref(null)

const loadServerLogs = async () => {
  serverLogsLoading.value = true
  try {
    const data = await adminAPI.getLogs(300)
    serverLogs.value = data.logs || []
    // 自动滚动到底部
    nextTick(() => {
      if (serverLogContainer.value) {
        serverLogContainer.value.scrollTop = serverLogContainer.value.scrollHeight
      }
    })
  } catch (e) {
    console.error('Failed to load server logs:', e)
  } finally {
    serverLogsLoading.value = false
  }
}

const clearServerLogs = async () => {
  if (!confirm('确定要清空后端日志缓冲区吗？')) return
  try {
    await adminAPI.clearLogs()
    serverLogs.value = []
  } catch (e) {
    alert('清空失败: ' + (e.message || '未知错误'))
  }
}

// 日志颜色：根据关键词高亮
const getLogColor = (msg) => {
  if (!msg) return 'text-gray-300'
  if (msg.includes('失败') || msg.includes('错误') || msg.includes('Fatal') || msg.includes('Error') || msg.includes('panic')) {
    return 'text-red-400'
  }
  if (msg.includes('警告') || msg.includes('Warning') || msg.includes('⚡')) {
    return 'text-yellow-400'
  }
  if (msg.includes('启动') || msg.includes('完成') || msg.includes('成功') || msg.includes('已创建')) {
    return 'text-green-400'
  }
  if (msg.includes('[Pool]') || msg.includes('[EBPF')) {
    return 'text-cyan-400'
  }
  if (msg.includes('[Admin]') || msg.includes('[Auth]')) {
    return 'text-purple-400'
  }
  return 'text-gray-300'
}

// Tab 切换时重触发淡入动画 + 加载数据
watch(activeTab, (tab) => {
  // 重触发 tab-content 淡入动画
  nextTick(() => {
    const el = document.querySelector('.tab-content[style=""]') 
      || document.querySelector(`.tab-content:not([style*="display: none"])`)
    if (el) {
      el.style.animation = 'none'
      // 强制 reflow
      void el.offsetHeight
      el.style.animation = ''
    }
  })

  if (tab === 'llm') {
    loadLLMStatus()
  }
  if (tab === 'logs') {
    loadServerLogs()
  }
})

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
/* Tab 切换淡入动画 */
.tab-content {
  animation: tabFadeIn 0.2s ease-out;
}

@keyframes tabFadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

pre, code {
  font-family: 'JetBrains Mono', monospace;
}

/* 自定义滚动条 */
.scrollbar-thin::-webkit-scrollbar {
  width: 6px;
}
.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}
.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>
