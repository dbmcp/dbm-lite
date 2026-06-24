<!--
@Project: DBM-Lite 轻量级全域数据库管控平台
@Version: v0.1.0
@Author: DB老王
@License: Apache-2.0 OR MulanPSL-2.0
-->
<template>
	<div class="dashboard-container">
		<!-- 顶部欢迎横幅 -->
		<div class="welcome-banner">
			<div class="welcome-left">
				<div class="welcome-title">
					<span class="waving">👋</span>
					欢迎回来，{{ userStore.displayName }}
				</div>
				<div class="welcome-subtitle">
					{{ greetingText }} · {{ todayText }} · 系统时间 {{ currentTimeText }}
				</div>
			</div>
			<div class="welcome-right">
				<div class="quick-actions">
					<el-button round @click="loadData" :loading="loading">
						<span class="refresh-icon">⟳</span>刷新数据
					</el-button>
					<el-button type="primary" :icon="EditPen" round @click="router.push('/sql/sqlide')">
						打开 SQL IDE
					</el-button>
					<el-button :icon="Coin" round @click="router.push('/datasources')">
						管理数据源
					</el-button>
					<el-button v-if="userStore.isAdmin" :icon="User" round @click="router.push('/accounts')">
						用户管理
					</el-button>
				</div>
			</div>
		</div>

		<!-- 加载中 -->
		<el-skeleton v-if="loading" :rows="8" animated />

		<template v-else>
			<!-- 核心指标卡 -->
			<div class="stat-grid">
				<div v-for="(card, idx) in statCards" :key="idx" class="stat-card" :style="{ background: card.gradient }">
					<div class="stat-card-top">
						<div>
							<div class="stat-card-label">{{ card.label }}</div>
							<div class="stat-card-number">{{ card.value }}</div>
							<div v-if="card.sub" class="stat-card-sub">{{ card.sub }}</div>
						</div>
						<div class="stat-card-icon" :style="{ background: card.iconBg }">
							<el-icon :size="26" color="#fff">
								<component :is="card.icon" />
							</el-icon>
						</div>
					</div>
					<div v-if="card.trend && card.trend.length" class="stat-card-trend">
						<svg viewBox="0 0 200 40" preserveAspectRatio="none" width="100%" height="40">
							<polyline
								:points="buildTrendPoints(card.trend)"
								fill="none"
								stroke="#ffffff"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								opacity="0.85"
							/>
						</svg>
					</div>
				</div>
			</div>

			<!-- 第二行：数据库类型分布 + 数据源健康状态 -->
			<div class="row-grid">
				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">按数据库类型分布</span>
						<span class="panel-sub">总计 {{ stats.totalDatasources }} 个数据源</span>
					</div>
					<div v-if="stats.datasourceByType && stats.datasourceByType.length" class="bar-list">
						<div v-for="(item, i) in stats.datasourceByType" :key="i" class="bar-item">
							<div class="bar-item-head">
								<span class="bar-item-name">{{ normalizeDBType(item.name) }}</span>
								<span class="bar-item-count">{{ item.count }}</span>
							</div>
							<div class="bar-track">
								<div
									class="bar-fill"
									:style="{
										width: percentOf(item.count, stats.totalDatasources) + '%',
										background: barColor(i)
									}"
								></div>
							</div>
							<div class="bar-percent">{{ percentOf(item.count, stats.totalDatasources).toFixed(1) }}%</div>
						</div>
					</div>
					<el-empty v-else description="暂无数据源" :image-size="80" />
				</div>

				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">数据源健康状态</span>
						<span class="panel-sub">在线 / 失败 / 未检测</span>
					</div>
					<div class="health-grid">
						<div class="health-item health-online">
							<div class="health-dot"></div>
							<div class="health-text">
								<div class="health-num">{{ stats.onlineDatasources }}</div>
								<div class="health-label">在线</div>
							</div>
						</div>
						<div class="health-item health-failed">
							<div class="health-dot"></div>
							<div class="health-text">
								<div class="health-num">{{ stats.failedDatasources }}</div>
								<div class="health-label">连接失败</div>
							</div>
						</div>
						<div class="health-item health-unknown">
							<div class="health-dot"></div>
							<div class="health-text">
								<div class="health-num">{{ Math.max(0, stats.totalDatasources - stats.onlineDatasources - stats.failedDatasources) }}</div>
								<div class="health-label">未检测</div>
							</div>
						</div>
					</div>

					<div class="metrics-row">
						<div class="metric-card">
							<div class="metric-label">平均执行耗时</div>
							<div class="metric-value">{{ stats.avgLatencyMs }} <span class="metric-unit">ms</span></div>
						</div>
						<div class="metric-card">
							<div class="metric-label">执行失败率</div>
							<div class="metric-value" :class="failRateClass">{{ (stats.failRate * 100).toFixed(2) }}<span class="metric-unit">%</span></div>
						</div>
					</div>
				</div>
			</div>

			<!-- 第三行：SQL 执行趋势 -->
			<div class="panel">
				<div class="panel-header">
					<span class="panel-title">近 14 天 SQL 执行趋势</span>
					<span class="panel-sub">总计 {{ stats.totalSqlExec }} 次执行</span>
				</div>
				<div class="chart-wrap">
					<svg viewBox="0 0 900 240" preserveAspectRatio="xMidYMid meet" width="100%" height="240">
						<!-- 横向网格线 -->
						<g stroke="#e4e7ed" stroke-width="1">
							<line x1="40" y1="30" x2="880" y2="30" stroke-dasharray="3,3" />
							<line x1="40" y1="80" x2="880" y2="80" stroke-dasharray="3,3" />
							<line x1="40" y1="130" x2="880" y2="130" stroke-dasharray="3,3" />
							<line x1="40" y1="180" x2="880" y2="180" />
						</g>
						<!-- Y 轴刻度 -->
						<g font-size="11" fill="#909399">
							<text x="30" y="34" text-anchor="end">{{ yMax }}</text>
							<text x="30" y="84" text-anchor="end">{{ Math.round(yMax * 0.75) }}</text>
							<text x="30" y="134" text-anchor="end">{{ Math.round(yMax * 0.5) }}</text>
							<text x="30" y="184" text-anchor="end">0</text>
						</g>
						<!-- 柱形 -->
						<g v-if="chartPoints.length">
							<rect
								v-for="(p, i) in chartPoints"
								:key="'bar-' + i"
								:x="p.x - 18"
								:y="p.y"
								width="36"
								:height="180 - p.y"
								rx="4"
								fill="url(#barGradient)"
							/>
						</g>
						<!-- 折线 -->
						<polyline
							v-if="chartPoints.length"
							:points="chartPoints.map((p: any) => p.x + ',' + p.y).join(' ')"
							fill="none"
							stroke="#409eff"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						/>
						<!-- 点 -->
						<circle
							v-for="(p, i) in chartPoints"
							:key="'dot-' + i"
							:cx="p.x"
							:cy="p.y"
							r="3.5"
							fill="#fff"
							stroke="#409eff"
							stroke-width="2"
						/>
						<!-- X 轴标签 -->
						<g font-size="11" fill="#909399">
							<text
								v-for="(p, i) in chartPoints"
								:key="'label-' + i"
								:x="p.x"
								y="210"
								text-anchor="middle"
							>{{ p.label }}</text>
						</g>
						<defs>
							<linearGradient id="barGradient" x1="0" y1="0" x2="0" y2="1">
								<stop offset="0%" stop-color="#79bbff" />
								<stop offset="100%" stop-color="#c6e2ff" />
							</linearGradient>
						</defs>
					</svg>
				</div>
			</div>

			<!-- 第四行：Top 数据源 & Top 用户 + 最近执行 -->
			<div class="row-grid">
				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">活跃数据源 TOP</span>
						<span class="panel-sub">按执行次数</span>
					</div>
					<div v-if="stats.topDatasources && stats.topDatasources.length" class="rank-list">
						<div v-for="(item, i) in stats.topDatasources" :key="i" class="rank-item">
							<span class="rank-badge" :style="{ background: rankColor(i) }">{{ i + 1 }}</span>
							<span class="rank-name">{{ item.name }}</span>
							<span class="rank-count">{{ formatNumber(item.count) }} 次</span>
						</div>
					</div>
					<el-empty v-else description="暂无执行数据" :image-size="80" />
				</div>

				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">活跃用户 TOP</span>
						<span class="panel-sub">按执行次数</span>
					</div>
					<div v-if="stats.topUsers && stats.topUsers.length" class="rank-list">
						<div v-for="(item, i) in stats.topUsers" :key="i" class="rank-item">
							<span class="rank-badge" :style="{ background: rankColor(i) }">{{ i + 1 }}</span>
							<span class="rank-name">{{ item.name }}</span>
							<span class="rank-count">{{ formatNumber(item.count) }} 次</span>
						</div>
					</div>
					<el-empty v-else description="暂无执行数据" :image-size="80" />
				</div>
			</div>

			<!-- 第五行：最近 SQL 执行 & 最近操作审计 -->
			<div class="row-grid">
				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">最近 SQL 执行</span>
						<el-button link type="primary" @click="router.push('/sql/sqlide')">打开 IDE</el-button>
					</div>
					<el-table
						v-if="stats.recentSqlHistory && stats.recentSqlHistory.length"
						:data="stats.recentSqlHistory"
						style="width:100%"
						size="small"
						empty-text="暂无数据"
					>
						<el-table-column prop="datasourceName" label="数据源" width="160" show-overflow-tooltip />
						<el-table-column prop="sqlText" label="SQL 语句" show-overflow-tooltip>
							<template #default="{ row }">
								<span class="sql-cell">{{ row.sqlText }}</span>
							</template>
						</el-table-column>
						<el-table-column prop="durationMs" label="耗时" width="80" align="right">
							<template #default="{ row }">{{ row.durationMs }} ms</template>
						</el-table-column>
						<el-table-column prop="status" label="状态" width="90" align="center">
							<template #default="{ row }">
								<el-tag v-if="row.status === 'success'" type="success" size="small" effect="light">成功</el-tag>
								<el-tag v-else type="danger" size="small" effect="light">失败</el-tag>
							</template>
						</el-table-column>
						<el-table-column prop="createdAt" label="时间" width="170">
							<template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
						</el-table-column>
					</el-table>
					<el-empty v-else description="暂无执行记录" :image-size="80" />
				</div>

				<div class="panel">
					<div class="panel-header">
						<span class="panel-title">最近操作记录</span>
						<el-button v-if="userStore.isAdmin" link type="primary" @click="router.push('/audit')">查看全部</el-button>
					</div>
					<div v-if="stats.recentAudit && stats.recentAudit.length" class="timeline">
						<div v-for="(item, i) in stats.recentAudit" :key="i" class="timeline-item">
							<div class="timeline-dot"></div>
							<div class="timeline-content">
								<div class="timeline-title">
									<strong>{{ item.username || 'system' }}</strong>
									· <span class="timeline-module">{{ item.module }}</span>
									<span class="timeline-action">· {{ item.action }}</span>
								</div>
								<div class="timeline-detail">{{ item.detail }}</div>
								<div class="timeline-time">{{ formatTime(item.createdAt) }}</div>
							</div>
						</div>
					</div>
					<el-empty v-else description="暂无操作记录" :image-size="80" />
				</div>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getDashboardStats } from '@/api/sql'
import { EditPen, Coin, User, DataLine, DataBoard, Monitor, MagicStick, Refresh } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(true)

interface StatsData {
	totalUsers: number
	totalDatasources: number
	totalSqlExec: number
	todaySqlExec: number
	totalAuditLogs: number
	todayAuditLogs: number
	onlineDatasources: number
	failedDatasources: number
	totalServers: number
	totalBusinesses: number
	totalProjects: number
	datasourceByType: Array<{ name: string; count: number }>
	recentSqlHistory: Array<{
		datasourceId: string
		datasourceName: string
		username: string
		sqlText: string
		durationMs: number
		status: string
		createdAt: string
	}>
	recentAudit: Array<{
		username: string
		module: string
		action: string
		detail: string
		status: string
		createdAt: string
	}>
	sqlDailyTrend: Array<{ day: string; count: number }>
	auditDailyTrend: Array<{ day: string; count: number }>
	topDatasources: Array<{ name: string; count: number }>
	topUsers: Array<{ name: string; count: number }>
	avgLatencyMs: number
	failRate: number
	systemNow: string
}

const stats = ref<StatsData>({
	totalUsers: 0,
	totalDatasources: 0,
	totalSqlExec: 0,
	todaySqlExec: 0,
	totalAuditLogs: 0,
	todayAuditLogs: 0,
	onlineDatasources: 0,
	failedDatasources: 0,
	totalServers: 0,
	totalBusinesses: 0,
	totalProjects: 0,
	datasourceByType: [],
	recentSqlHistory: [],
	recentAudit: [],
	sqlDailyTrend: [],
	auditDailyTrend: [],
	topDatasources: [],
	topUsers: [],
	avgLatencyMs: 0,
	failRate: 0,
	systemNow: ''
})

const hourNow = new Date().getHours()
const greetingText = computed(() => {
	if (hourNow < 6) return '夜深了，注意休息'
	if (hourNow < 12) return '上午好'
	if (hourNow < 14) return '中午好'
	if (hourNow < 18) return '下午好'
	return '晚上好'
})

const currentTimeText = ref('')
const todayText = computed(() => {
	const d = new Date()
	return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
})

let timer: any = null
function updateClock() {
	const d = new Date()
	currentTimeText.value =
		String(d.getHours()).padStart(2, '0') +
		':' +
		String(d.getMinutes()).padStart(2, '0') +
		':' +
		String(d.getSeconds()).padStart(2, '0')
}

const statCards = computed(() => {
	return [
		{
			label: '总用户数',
			value: formatNumber(stats.value.totalUsers),
			sub: '今日新增 · 0',
			icon: 'User',
			gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
			iconBg: 'rgba(255,255,255,0.25)',
			trend: [] as number[]
		},
		{
			label: '数据源数',
			value: formatNumber(stats.value.totalDatasources),
			sub: `在线 ${stats.value.onlineDatasources} · 失败 ${stats.value.failedDatasources}`,
			icon: 'Coin',
			gradient: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)',
			iconBg: 'rgba(255,255,255,0.25)',
			trend: [] as number[]
		},
		{
			label: 'SQL 执行总量',
			value: formatNumber(stats.value.totalSqlExec),
			sub: `今日 ${formatNumber(stats.value.todaySqlExec)} 次`,
			icon: 'EditPen',
			gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
			iconBg: 'rgba(255,255,255,0.25)',
			trend: (stats.value.sqlDailyTrend || []).map(i => i.count)
		},
		{
			label: '操作审计记录',
			value: formatNumber(stats.value.totalAuditLogs),
			sub: `今日 ${formatNumber(stats.value.todayAuditLogs)} 条`,
			icon: 'DataLine',
			gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
			iconBg: 'rgba(255,255,255,0.25)',
			trend: (stats.value.auditDailyTrend || []).map(i => i.count)
		}
	]
})

const failRateClass = computed(() => (stats.value.failRate > 0.05 ? 'metric-danger' : ''))

const chartPoints = computed(() => {
	const trend = stats.value.sqlDailyTrend || []
	if (!trend.length) return []
	const max = Math.max(1, ...trend.map((t: any) => t.count))
	const left = 50
	const right = 870
	const step = (right - left) / (trend.length - 1 || 1)
	return trend.map((t: any, i: number) => ({
		x: left + i * step,
		y: 180 - (t.count / max) * 150,
		label: (t.day || '').slice(5),
		value: t.count
	}))
})

const yMax = computed(() => {
	const trend = stats.value.sqlDailyTrend || []
	const max = Math.max(1, ...trend.map((t: any) => t.count))
	return Math.ceil(max / 10) * 10
})

function formatNumber(n: number): string {
	if (n === null || n === undefined) return '0'
	return String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function pad2(n: number): string { return n < 10 ? '0' + n : '' + n }

function formatTime(t: any): string {
	if (!t) return ''
	let d: Date
	if (t instanceof Date) {
		d = t
	} else if (typeof t === 'string') {
		const s = t.replace('T', ' ').replace(/\.[0-9]+.*$/, '').replace(/[+-]\d{2}:?\d{2}$/, '').trim()
		d = new Date(s.replace(/-/g, '/'))
		if (isNaN(d.getTime())) d = new Date(t)
	} else {
		d = new Date(t)
	}
	if (isNaN(d.getTime())) return String(t)
	return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`
}

function percentOf(n: number, total: number): number {
	if (!total) return 0
	return (n / total) * 100
}

function normalizeDBType(t: string): string {
	if (!t) return '未知'
	const lower = t.toLowerCase()
	const map: Record<string, string> = {
		mysql: 'MySQL',
		tidb: 'TiDB',
		sqlite: 'SQLite',
		postgres: 'PostgreSQL',
		postgresql: 'PostgreSQL',
		mssql: 'SQL Server',
		sqlserver: 'SQL Server',
		oracle: 'Oracle'
	}
	return map[lower] || t
}

function barColor(i: number): string {
	const palette = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#9b59b6', '#1abc9c']
	return palette[i % palette.length]
}

function rankColor(i: number): string {
	if (i === 0) return '#f56c6c'
	if (i === 1) return '#e6a23c'
	if (i === 2) return '#409eff'
	return '#909399'
}

function buildTrendPoints(arr: number[]): string {
	if (!arr || !arr.length) return ''
	const max = Math.max(1, ...arr)
	const w = 200
	const h = 40
	const step = w / (arr.length - 1 || 1)
	return arr
		.map((v: number, i: number) => `${i * step},${h - (v / max) * (h - 4) - 2}`)
		.join(' ')
}

async function loadData(): Promise<void> {
	try {
		loading.value = true
		const data: any = await getDashboardStats()
		stats.value = {
			totalUsers: data.totalUsers ?? 0,
			totalDatasources: data.totalDatasources ?? 0,
			totalSqlExec: data.totalSqlExec ?? 0,
			todaySqlExec: data.todaySqlExec ?? 0,
			totalAuditLogs: data.totalAuditLogs ?? 0,
			todayAuditLogs: data.todayAuditLogs ?? 0,
			onlineDatasources: data.onlineDatasources ?? 0,
			failedDatasources: data.failedDatasources ?? 0,
			totalServers: data.totalServers ?? 0,
			totalBusinesses: data.totalBusinesses ?? 0,
			totalProjects: data.totalProjects ?? 0,
			datasourceByType: data.datasourceByType ?? [],
			recentSqlHistory: data.recentSqlHistory ?? [],
			recentAudit: data.recentAudit ?? [],
			sqlDailyTrend: data.sqlDailyTrend ?? [],
			auditDailyTrend: data.auditDailyTrend ?? [],
			topDatasources: data.topDatasources ?? [],
			topUsers: data.topUsers ?? [],
			avgLatencyMs: data.avgLatencyMs ?? 0,
			failRate: data.failRate ?? 0,
			systemNow: data.systemNow ?? ''
		}
	} catch (e) {
		console.error(e)
	} finally {
		loading.value = false
	}
}

onMounted(() => {
	updateClock()
	timer = setInterval(updateClock, 1000)
	loadData()
})

onBeforeUnmount(() => {
	if (timer) clearInterval(timer)
})
</script>

<style scoped>
.dashboard-container {
	padding: 16px 20px 24px;
	background: #f5f7fa;
	min-height: calc(100vh - 40px);
}

.welcome-banner {
	background: linear-gradient(135deg, #ffffff 0%, #ecf5ff 100%);
	border-radius: 10px;
	padding: 20px 24px;
	display: flex;
	align-items: center;
	justify-content: space-between;
	box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
	margin-bottom: 16px;
	border: 1px solid #e4e7ed;
}

.welcome-title {
	font-size: 22px;
	font-weight: 600;
	color: #303133;
}

.waving {
	display: inline-block;
	animation: wave 1.8s infinite;
	transform-origin: 70% 70%;
	margin-right: 4px;
}

@keyframes wave {
	0%, 60%, 100% { transform: rotate(0); }
	10%, 30% { transform: rotate(14deg); }
	20%, 40% { transform: rotate(-8deg); }
	50% { transform: rotate(10deg); }
}

.welcome-subtitle {
	margin-top: 6px;
	font-size: 13px;
	color: #606266;
}

.quick-actions {
	display: flex;
	gap: 10px;
}

.stat-grid {
	display: grid;
	grid-template-columns: repeat(4, minmax(0, 1fr));
	gap: 14px;
	margin-bottom: 16px;
}

.stat-card {
	color: #fff;
	border-radius: 10px;
	padding: 18px 20px;
	box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
	min-height: 130px;
	position: relative;
	overflow: hidden;
}

.stat-card-top {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.stat-card-label {
	font-size: 13px;
	opacity: 0.9;
}

.stat-card-number {
	font-size: 30px;
	font-weight: 700;
	margin-top: 6px;
	letter-spacing: 0.5px;
}

.stat-card-sub {
	font-size: 12px;
	opacity: 0.85;
	margin-top: 4px;
}

.stat-card-icon {
	width: 48px;
	height: 48px;
	border-radius: 10px;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.stat-card-trend {
	margin-top: 8px;
	opacity: 0.9;
}

.row-grid {
	display: grid;
	grid-template-columns: repeat(2, minmax(0, 1fr));
	gap: 14px;
	margin-bottom: 16px;
}

.panel {
	background: #fff;
	border-radius: 10px;
	padding: 18px 20px;
	box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
	border: 1px solid #ebeef5;
}

.panel-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 14px;
}

.panel-title {
	font-size: 15px;
	font-weight: 600;
	color: #303133;
}

.panel-sub {
	font-size: 12px;
	color: #909399;
}

.bar-list {
	display: flex;
	flex-direction: column;
	gap: 14px;
}

.bar-item {
	display: grid;
	grid-template-columns: 1fr 60px;
	grid-column-gap: 10px;
	align-items: center;
}

.bar-item-head {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 6px;
	grid-column: 1 / span 2;
}

.bar-item-name {
	font-size: 13px;
	color: #303133;
	font-weight: 500;
}

.bar-item-count {
	font-size: 13px;
	color: #606266;
}

.bar-track {
	grid-column: 1;
	height: 8px;
	background: #f0f2f5;
	border-radius: 4px;
	overflow: hidden;
}

.bar-fill {
	height: 100%;
	border-radius: 4px;
	transition: width 0.4s ease;
}

.bar-percent {
	grid-column: 2;
	text-align: right;
	font-size: 12px;
	color: #909399;
}

.health-grid {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	gap: 12px;
	margin-bottom: 16px;
}

.health-item {
	padding: 14px;
	border-radius: 8px;
	display: flex;
	gap: 12px;
	align-items: center;
	border: 1px solid #ebeef5;
}

.health-online { background: #f0f9eb; }
.health-failed { background: #fef0f0; }
.health-unknown { background: #f4f4f5; }

.health-dot {
	width: 10px;
	height: 10px;
	border-radius: 50%;
	background: #67c23a;
	flex-shrink: 0;
}

.health-failed .health-dot { background: #f56c6c; }
.health-unknown .health-dot { background: #909399; }

.health-num {
	font-size: 22px;
	font-weight: 700;
	color: #303133;
	line-height: 1;
}

.health-label {
	font-size: 12px;
	color: #606266;
	margin-top: 4px;
}

.metrics-row {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 12px;
}

.metric-card {
	padding: 14px;
	background: #f5f7fa;
	border-radius: 8px;
}

.metric-label {
	font-size: 12px;
	color: #606266;
}

.metric-value {
	font-size: 22px;
	font-weight: 700;
	color: #303133;
	margin-top: 6px;
}

.metric-unit {
	font-size: 12px;
	font-weight: 400;
	color: #909399;
	margin-left: 2px;
}

.metric-danger { color: #f56c6c; }

.chart-wrap {
	background: #fff;
	border-radius: 8px;
	padding: 8px 0 4px;
}

.rank-list {
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.rank-item {
	display: flex;
	align-items: center;
	gap: 12px;
	padding: 8px 10px;
	border-radius: 6px;
	background: #fafbfc;
	transition: background 0.2s;
}

.rank-item:hover { background: #f0f7ff; }

.rank-badge {
	width: 24px;
	height: 24px;
	border-radius: 6px;
	color: #fff;
	font-size: 12px;
	font-weight: 600;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.rank-name {
	flex: 1;
	font-size: 13px;
	color: #303133;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.rank-count {
	font-size: 12px;
	color: #606266;
	font-weight: 500;
}

.sql-cell {
	font-family: 'SF Mono', Consolas, Menlo, monospace;
	font-size: 12px;
	color: #303133;
}

.timeline {
	position: relative;
	padding-left: 4px;
}

.timeline-item {
	display: flex;
	gap: 12px;
	padding-bottom: 14px;
	position: relative;
}

.timeline-item:not(:last-child)::after {
	content: '';
	position: absolute;
	left: 5px;
	top: 14px;
	bottom: 0;
	width: 1px;
	background: #ebeef5;
}

.timeline-dot {
	width: 10px;
	height: 10px;
	border-radius: 50%;
	background: #409eff;
	margin-top: 5px;
	flex-shrink: 0;
	box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.15);
}

.timeline-content { flex: 1; min-width: 0; }

.timeline-title {
	font-size: 13px;
	color: #303133;
	line-height: 1.5;
}

.timeline-module {
	color: #67c23a;
	font-size: 12px;
}

.timeline-action {
	color: #909399;
	font-size: 12px;
}

.timeline-detail {
	font-size: 12px;
	color: #606266;
	margin-top: 4px;
	word-break: break-all;
	line-height: 1.5;
}

.timeline-time {
	font-size: 11px;
	color: #c0c4cc;
	margin-top: 4px;
}

@media (max-width: 1200px) {
	.stat-grid { grid-template-columns: repeat(2, 1fr); }
	.row-grid { grid-template-columns: 1fr; }
	.welcome-banner { flex-direction: column; align-items: flex-start; gap: 12px; }
}

@media (max-width: 720px) {
	.stat-grid { grid-template-columns: 1fr; }
	.health-grid { grid-template-columns: 1fr; }
	.quick-actions { flex-wrap: wrap; }
}
</style>
