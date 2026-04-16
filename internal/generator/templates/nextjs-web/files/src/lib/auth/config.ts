import type { NextAuthOptions } from "next-auth"
import GoogleProvider from "next-auth/providers/google"

import "./types"

// Google OAuth 설정
// GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET은 .env.local에서 읽어요.
// plab-app create --researchers-only 로 생성하면 자동으로 세팅됩니다.
export const authOptions: NextAuthOptions = {
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID ?? "",
      clientSecret: process.env.GOOGLE_CLIENT_SECRET ?? "",
    }),
  ],
  pages: {
    signIn: "/login",
  },
  callbacks: {
    async signIn() {
      // 플랩 도메인 이메일만 허용하려면 아래처럼 수정하세요:
      // async signIn({ profile }) {
      //   return profile?.email?.endsWith("@plabfootball.com") ?? false
      // }
      return true
    },
    async session({ session, token }) {
      if (session.user) {
        session.user.id = token.sub ?? ""
      }
      return session
    },
  },
}
