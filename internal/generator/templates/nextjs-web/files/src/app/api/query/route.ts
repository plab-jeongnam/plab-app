import { NextRequest, NextResponse } from "next/server"

import { plabQuery } from "@/lib/plab"

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { query, params } = body

    if (!query) {
      return NextResponse.json(
        { success: false, error: "쿼리를 입력해 주세요" },
        { status: 400 },
      )
    }

    const result = await plabQuery(query, params)
    return NextResponse.json(result)
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "알 수 없는 오류"
    return NextResponse.json(
      { success: false, error: message },
      { status: 500 },
    )
  }
}
