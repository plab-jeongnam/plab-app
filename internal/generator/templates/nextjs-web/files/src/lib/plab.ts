const PLAB_API_URL = process.env.PLAB_API_URL || "https://vibe.techin.pe.kr"
const PLAB_API_KEY = process.env.PLAB_API_KEY || ""

interface PlabQueryResponse {
  success: boolean
  data?: Record<string, unknown>[]
  rowCount?: number
  executionTime?: string
  error?: string
}

interface PlabTablesResponse {
  success: boolean
  data?: string[]
  error?: string
}

export async function plabQuery(
  query: string,
  params?: unknown[],
): Promise<PlabQueryResponse> {
  const response = await fetch(`${PLAB_API_URL}/api/query`, {
    method: "POST",
    headers: {
      "X-API-Key": PLAB_API_KEY,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ query, params }),
  })

  if (!response.ok) {
    throw new Error(`Plab API error: ${response.status}`)
  }

  return response.json()
}

export async function plabTables(): Promise<PlabTablesResponse> {
  const response = await fetch(`${PLAB_API_URL}/api/tables`, {
    method: "GET",
    headers: {
      "X-API-Key": PLAB_API_KEY,
    },
  })

  if (!response.ok) {
    throw new Error(`Plab API error: ${response.status}`)
  }

  return response.json()
}
