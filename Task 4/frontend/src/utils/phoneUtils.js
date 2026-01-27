export function normalizePhone(phone) {
  if (!phone) return ''
  
  let cleaned = phone.replace(/\D/g, '')
  
  if (cleaned.startsWith('8') && cleaned.length === 11) {
    cleaned = '7' + cleaned.substring(1)
  }
  
  if (cleaned.length === 10) {
    cleaned = '7' + cleaned
  }
  
  if (cleaned.length === 11 && cleaned.startsWith('7')) {
    return cleaned
  }
  
  return cleaned
}

export function formatPhone(phone) {
  if (!phone) return ''
  
  const normalized = normalizePhone(phone)
  
  if (normalized.length === 11 && normalized.startsWith('7')) {
    return `7 (${normalized.substring(1, 4)}) ${normalized.substring(4, 7)}-${normalized.substring(7, 9)}-${normalized.substring(9, 11)}`
  }
  
  return normalized
}