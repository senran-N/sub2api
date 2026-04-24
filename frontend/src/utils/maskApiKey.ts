export function maskApiKey(key: string): string {
  const value = String(key || '')
  if (value.length <= 12) return value
  return `${value.slice(0, 8)}...${value.slice(-4)}`
}
