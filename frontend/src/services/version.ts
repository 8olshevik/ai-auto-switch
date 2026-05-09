import { updateApi } from '@/api'

export const fetchCurrentVersion = async (): Promise<string> => {
  const res = await updateApi.getVersion()
  return res.data.version ?? res.data
}
