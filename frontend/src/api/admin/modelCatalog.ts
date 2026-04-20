import { apiClient } from '../client'

export interface AdminModelCatalogEntry {
  id: string
  display_name: string
  capability: string
  protocol_family: string
  required_tier: string
  aliases: string[]
  supports_stream: boolean
  supports_tools: boolean
}

export interface AdminModelCatalogResponse {
  platform: string
  models: AdminModelCatalogEntry[]
}

export async function getModelCatalog(platform: string): Promise<AdminModelCatalogResponse> {
  const { data } = await apiClient.get<AdminModelCatalogResponse>('/admin/model-catalog', {
    params: { platform }
  })
  return data
}

export default {
  getModelCatalog
}
