import { skillsApi } from '@/api'

export type SkillSummary = {
  key: string
  name: string
  description: string
  directory: string
  readme_url: string
  installed: boolean
  enabled: boolean
  license_file?: string
  platform: 'claude' | 'codex' | ''
  install_location: 'user' | 'project' | ''
  repo_owner?: string
  repo_name?: string
  repo_branch?: string
}

export type SkillRepoConfig = {
  owner: string
  name: string
  branch: string
  enabled: boolean
}

export type InstallSkillPayload = {
  directory: string
  repo_owner?: string
  repo_name?: string
  repo_branch?: string
  platform?: 'claude' | 'codex'
  location?: 'user' | 'project'
}

export const fetchSkills = async (): Promise<SkillSummary[]> => {
  const res = await skillsApi.list()
  return res.data
}

export const fetchSkillsForPlatform = async (platform: 'claude' | 'codex'): Promise<SkillSummary[]> => {
  const res = await skillsApi.list(platform)
  return res.data
}

export const installSkill = async (payload: InstallSkillPayload): Promise<void> => {
  await skillsApi.install(payload)
}

export const uninstallSkill = async (directory: string): Promise<void> => {
  await skillsApi.uninstall(directory)
}

export const uninstallSkillEx = async (
  directory: string,
  platform: string,
  location: string
): Promise<void> => {
  await skillsApi.uninstall(directory, platform, location)
}

export const toggleSkill = async (
  directory: string,
  platform: string,
  location: string,
  enabled: boolean
): Promise<void> => {
  await skillsApi.toggle(directory, platform, location, enabled)
}

export const getSkillContent = async (
  directory: string,
  platform: string,
  location: string
): Promise<string> => {
  const res = await skillsApi.getContent(directory, platform, location)
  return res.data
}

export const saveSkillContent = async (
  directory: string,
  platform: string,
  location: string,
  content: string
): Promise<void> => {
  await skillsApi.saveContent(directory, platform, location, content)
}

export const openSkillFolder = async (_platform: string, _location: string): Promise<void> => {
  // In web app, this is not applicable — no local filesystem access
  console.info('[web] openSkillFolder is not available in web mode')
}

export const fetchSkillRepos = async (): Promise<SkillRepoConfig[]> => {
  const res = await skillsApi.getRepos()
  return res.data
}

export const addSkillRepo = async (repo: Partial<SkillRepoConfig>): Promise<SkillRepoConfig[]> => {
  const res = await skillsApi.addRepo(repo as { owner: string; name: string; branch: string })
  return res.data
}

export const removeSkillRepo = async (owner: string, name: string): Promise<SkillRepoConfig[]> => {
  const res = await skillsApi.removeRepo(owner, name)
  return res.data
}
