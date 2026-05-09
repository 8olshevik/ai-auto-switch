<template>
  <div class="gateway-page">
    <div class="page-header">
      <h2>API 网关</h2>
      <button class="btn btn-primary" @click="showCreateDialog = true">
        <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
          <path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        创建 API Key
      </button>
    </div>

    <!-- API Keys Section -->
    <section class="section">
      <h3 class="section-title">API Keys</h3>
      <div class="card">
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="keys.length === 0" class="empty-state">
          暂无 API Keys，点击上方按钮创建
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>前缀</th>
              <th>速率限制</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>最近使用</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="key in keys" :key="key.id">
              <td>{{ key.name }}</td>
              <td><code class="key-prefix">{{ key.keyPrefix }}****</code></td>
              <td>
                <div class="rate-limit-cell">
                  <input
                    type="number"
                    :value="key.rateLimit"
                    min="1"
                    class="rate-input"
                    @change="(e) => updateRateLimit(key.id, Number((e.target as HTMLInputElement).value))"
                  />
                  <span class="unit">req/min</span>
                </div>
              </td>
              <td>
                <label class="switch">
                  <input
                    type="checkbox"
                    :checked="key.enabled"
                    @change="toggleKey(key.id, !key.enabled)"
                  />
                  <span class="slider"></span>
                </label>
              </td>
              <td>{{ formatDate(key.createdAt) }}</td>
              <td>{{ key.lastUsedAt ? formatDate(key.lastUsedAt) : '-' }}</td>
              <td>
                <button class="btn btn-danger btn-sm" @click="confirmDelete(key)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Usage Stats Section -->
    <section class="section">
      <h3 class="section-title">使用统计</h3>
      <div class="card">
        <div class="stats-header">
          <div class="total-requests">
            <span class="label">总请求数</span>
            <span class="value">{{ stats.totalRequests.toLocaleString() }}</span>
          </div>
          <button class="btn btn-secondary btn-sm" @click="loadStats">
            <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
              <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            刷新
          </button>
        </div>
        
        <div v-if="stats.bySourceApp.length > 0" class="chart-container">
          <Bar :data="chartData" :options="chartOptions" />
        </div>
        <div v-else class="empty-state">
          暂无统计数据
        </div>
      </div>
    </section>

    <!-- Create Key Dialog -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>创建 API Key</h3>
          <button class="close-btn" @click="showCreateDialog = false">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>名称</label>
            <input
              v-model="newKeyName"
              type="text"
              class="form-input"
              placeholder="例如：测试应用"
            />
          </div>
          <div class="form-group">
            <label>速率限制 (请求/分钟)</label>
            <input
              v-model.number="newKeyRateLimit"
              type="number"
              class="form-input"
              min="1"
              placeholder="60"
            />
          </div>
          
          <!-- Show newly created key -->
          <div v-if="createdKey" class="created-key-box">
            <p class="warning-text">⚠️ 请妥善保存此 Key，只显示一次！</p>
            <div class="key-display">
              <code>{{ createdKey }}</code>
              <button class="btn btn-secondary btn-sm" @click="copyKey">复制</button>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showCreateDialog = false">取消</button>
          <button 
            class="btn btn-primary" 
            :disabled="!newKeyName.trim() || creating"
            @click="createKey"
          >
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div v-if="showDeleteDialog" class="dialog-overlay" @click.self="showDeleteDialog = false">
      <div class="dialog dialog-sm">
        <div class="dialog-header">
          <h3>确认删除</h3>
        </div>
        <div class="dialog-body">
          <p>确定要删除 API Key「{{ keyToDelete?.name }}」吗？此操作无法撤销。</p>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showDeleteDialog = false">取消</button>
          <button class="btn btn-danger" :disabled="deleting" @click="deleteKey">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
} from 'chart.js'
import { gatewayApi, type GatewayKey, type GatewayStats } from '../api/gateway'

// Register Chart.js components
ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

// State
const loading = ref(true)
const creating = ref(false)
const deleting = ref(false)
const keys = ref<GatewayKey[]>([])
const stats = ref<GatewayStats>({ totalRequests: 0, bySourceApp: [] })

// Dialogs
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const newKeyName = ref('')
const newKeyRateLimit = ref(60)
const createdKey = ref('')
const keyToDelete = ref<GatewayKey | null>(null)

// Chart data
const chartData = computed(() => ({
  labels: stats.value.bySourceApp.map((s) => s.sourceApp || '未知'),
  datasets: [
    {
      label: '请求数',
      data: stats.value.bySourceApp.map((s) => s.count),
      backgroundColor: 'rgba(99, 102, 241, 0.7)',
      borderColor: 'rgb(99, 102, 241)',
      borderWidth: 1,
    },
  ],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        color: '#9ca3af',
      },
      grid: {
        color: 'rgba(255, 255, 255, 0.1)',
      },
    },
    x: {
      ticks: {
        color: '#9ca3af',
      },
      grid: {
        display: false,
      },
    },
  },
}

// Methods
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const loadKeys = async () => {
  try {
    const data = await gatewayApi.listKeys()
    keys.value = data
  } catch (error) {
    console.error('failed to load keys', error)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const data = await gatewayApi.getStats()
    stats.value = data
  } catch (error) {
    console.error('failed to load stats', error)
  }
}

const createKey = async () => {
  if (!newKeyName.value.trim() || creating.value) return

  creating.value = true
  try {
    const response = await gatewayApi.createKey({
      name: newKeyName.value.trim(),
      rateLimit: newKeyRateLimit.value || 60,
    })
    createdKey.value = response.rawKey
    keys.value.push(response.key)
  } catch (error) {
    console.error('failed to create key', error)
    alert('创建失败：' + (error as Error).message)
  } finally {
    creating.value = false
  }
}

const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(createdKey.value)
    alert('已复制到剪贴板')
  } catch {
    alert('复制失败')
  }
}

const confirmDelete = (key: GatewayKey) => {
  keyToDelete.value = key
  showDeleteDialog.value = true
}

const deleteKey = async () => {
  if (!keyToDelete.value || deleting.value) return

  deleting.value = true
  try {
    await gatewayApi.deleteKey(keyToDelete.value.id)
    keys.value = keys.value.filter((k) => k.id !== keyToDelete.value!.id)
    showDeleteDialog.value = false
    keyToDelete.value = null
  } catch (error) {
    console.error('failed to delete key', error)
    alert('删除失败：' + (error as Error).message)
  } finally {
    deleting.value = false
  }
}

const toggleKey = async (id: number, enabled: boolean) => {
  const key = keys.value.find((k) => k.id === id)
  if (!key) return

  try {
    await gatewayApi.toggleKey(id, enabled)
    key.enabled = enabled
  } catch (error) {
    console.error('failed to toggle key', error)
    // Revert the toggle
    key.enabled = !enabled
    alert('操作失败：' + (error as Error).message)
  }
}

const updateRateLimit = async (id: number, rateLimit: number) => {
  const key = keys.value.find((k) => k.id === id)
  if (!key) return

  const oldRateLimit = key.rateLimit
  try {
    await gatewayApi.updateRateLimit({ keyId: id, rateLimit })
    key.rateLimit = rateLimit
  } catch (error) {
    console.error('failed to update rate limit', error)
    // Revert
    key.rateLimit = oldRateLimit
    alert('更新速率限制失败：' + (error as Error).message)
  }
}

// Reset create dialog when closed
const resetCreateDialog = () => {
  showCreateDialog.value = false
  newKeyName.value = ''
  newKeyRateLimit.value = 60
  createdKey.value = ''
}

onMounted(async () => {
  await Promise.all([loadKeys(), loadStats()])
})
</script>

<style scoped>
.gateway-page {
  padding: 1.5rem;
  color: var(--text-primary, #e0e0e0);
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.section {
  margin-bottom: 2rem;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 500;
  margin-bottom: 1rem;
  color: var(--text-secondary, #a0a0a0);
}

.card {
  background: var(--bg-secondary, #2a2a2a);
  border-radius: 8px;
  padding: 1.25rem;
  border: 1px solid var(--border-color, #3a3a3a);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color, #3a3a3a);
}

.data-table th {
  font-weight: 500;
  color: var(--text-secondary, #a0a0a0);
  font-size: 0.875rem;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.key-prefix {
  background: var(--bg-tertiary, #3a3a3a);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
}

.rate-limit-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.rate-input {
  width: 80px;
  padding: 0.375rem 0.5rem;
  background: var(--bg-tertiary, #3a3a3a);
  border: 1px solid var(--border-color, #4a4a4a);
  border-radius: 4px;
  color: var(--text-primary, #e0e0e0);
  font-size: 0.875rem;
}

.unit {
  font-size: 0.75rem;
  color: var(--text-secondary, #a0a0a0);
}

/* Toggle Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-tertiary, #4a4a4a);
  transition: 0.3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #10b981;
}

input:checked + .slider:before {
  transform: translateX(20px);
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #6366f1;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #4f46e5;
}

.btn-secondary {
  background: var(--bg-tertiary, #3a3a3a);
  color: var(--text-primary, #e0e0e0);
}

.btn-secondary:hover:not(:disabled) {
  background: #4a4a4a;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

/* Stats */
.stats-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.total-requests {
  display: flex;
  flex-direction: column;
}

.total-requests .label {
  font-size: 0.875rem;
  color: var(--text-secondary, #a0a0a0);
}

.total-requests .value {
  font-size: 1.75rem;
  font-weight: 600;
  color: #6366f1;
}

.chart-container {
  height: 300px;
}

/* Dialog */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--bg-secondary, #2a2a2a);
  border-radius: 12px;
  width: 100%;
  max-width: 480px;
  border: 1px solid var(--border-color, #3a3a3a);
}

.dialog-sm {
  max-width: 360px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  border-bottom: 1px solid var(--border-color, #3a3a3a);
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary, #a0a0a0);
  cursor: pointer;
  padding: 0.25rem;
}

.close-btn:hover {
  color: var(--text-primary, #e0e0e0);
}

.dialog-body {
  padding: 1.25rem;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem;
  border-top: 1px solid var(--border-color, #3a3a3a);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 0.625rem 0.75rem;
  background: var(--bg-tertiary, #3a3a3a);
  border: 1px solid var(--border-color, #4a4a4a);
  border-radius: 6px;
  color: var(--text-primary, #e0e0e0);
  font-size: 0.9375rem;
}

.form-input:focus {
  outline: none;
  border-color: #6366f1;
}

.created-key-box {
  margin-top: 1.5rem;
  padding: 1rem;
  background: rgba(234, 179, 8, 0.1);
  border: 1px solid rgba(234, 179, 8, 0.3);
  border-radius: 8px;
}

.warning-text {
  color: #eab308;
  margin: 0 0 0.75rem;
  font-size: 0.875rem;
}

.key-display {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.key-display code {
  flex: 1;
  padding: 0.5rem 0.75rem;
  background: var(--bg-tertiary, #3a3a3a);
  border-radius: 4px;
  font-size: 0.875rem;
  word-break: break-all;
}

.loading,
.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary, #a0a0a0);
}
</style>