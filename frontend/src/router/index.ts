import { createRouter, createWebHistory } from 'vue-router'
import MainPage from '../components/Main/Index.vue'
import LogsPage from '../components/Logs/Index.vue'
import GeneralPage from '../components/General/Index.vue'
import McpPage from '../components/Mcp/index.vue'
import SkillPage from '../components/Skill/Index.vue'
import PromptsPage from '../components/Prompts/Index.vue'
import SpeedTestPage from '../components/SpeedTest/Index.vue'
import EnvCheckPage from '../components/EnvCheck/Index.vue'
import ConsolePage from '../components/Console/Index.vue'
import AvailabilityPage from '../components/Availability/Index.vue'
import LoginPage from '../pages/LoginPage.vue'
import GatewayPage from '../pages/GatewayPage.vue'
import AssistantPage from '../pages/AssistantPage.vue'

const routes = [
  { path: '/login', component: LoginPage, meta: { public: true } },
  { path: '/', component: MainPage },
  { path: '/prompts', component: PromptsPage },
  { path: '/mcp', component: McpPage },
  { path: '/skill', component: SkillPage },
  { path: '/availability', component: AvailabilityPage },
  { path: '/speedtest', component: SpeedTestPage },
  { path: '/env', component: EnvCheckPage },
  { path: '/logs', component: LogsPage },
  { path: '/console', component: ConsolePage },
  { path: '/settings', component: GeneralPage },
  { path: '/gateway', component: GatewayPage },
  { path: '/assistant', component: AssistantPage },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard: redirect to /login if not authenticated
router.beforeEach((to) => {
  const token = localStorage.getItem('token')

  // Allow access to public routes without authentication
  if (to.meta.public) {
    // If already logged in and going to login page, redirect to home
    if (token && to.path === '/login') {
      return '/'
    }
    return
  }

  // For non-public routes, require authentication
  if (!token) {
    return '/login'
  }
})

export default router
