import { settingsApi } from '@/api'

export type AppSettings = {
  show_heatmap: boolean
  show_home_title: boolean
  budget_total: number
  budget_used_adjustment: number
  budget_cycle_enabled: boolean
  budget_cycle_mode: string
  budget_refresh_time: string
  budget_refresh_day: number
  budget_show_countdown: boolean
  budget_show_forecast: boolean
  budget_forecast_method: string
  budget_total_codex: number
  budget_used_adjustment_codex: number
  budget_cycle_enabled_codex: boolean
  budget_cycle_mode_codex: string
  budget_refresh_time_codex: string
  budget_refresh_day_codex: number
  budget_show_countdown_codex: boolean
  budget_show_forecast_codex: boolean
  budget_forecast_method_codex: string
  auto_start: boolean
  auto_update: boolean
  auto_connectivity_test: boolean
  enable_switch_notify: boolean
  enable_round_robin: boolean
}

export const fetchAppSettings = async (): Promise<AppSettings> => {
  const res = await settingsApi.getApp()
  return res.data
}

export const saveAppSettings = async (settings: AppSettings): Promise<AppSettings> => {
  const res = await settingsApi.updateApp(settings)
  return res.data
}
