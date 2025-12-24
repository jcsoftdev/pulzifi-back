const TOKEN_KEY = 'auth_token'
const TOKEN_EXPIRY_DAYS = 7

export class TokenManager {
  static async getServerToken(): Promise<string | null> {
    try {
      // Dynamic import to avoid bundling next/headers in client code
      const { cookies } = await import('next/headers')
      const cookieStore = await cookies()
      const token = cookieStore.get(TOKEN_KEY)
      return token?.value || null
    } catch {
      return null
    }
  }

  static getClientToken(): string | null {
    if (typeof document === 'undefined') return null
    
    const cookies = document.cookie.split(';')
    const tokenCookie = cookies.find(c => c.trim().startsWith(`${TOKEN_KEY}=`))
    
    if (!tokenCookie) return null
    
    return tokenCookie.split('=')[1] || null
  }

  static setToken(token: string): void {
    if (typeof document === 'undefined') return

    const expires = new Date()
    expires.setDate(expires.getDate() + TOKEN_EXPIRY_DAYS)

    document.cookie = `${TOKEN_KEY}=${token}; expires=${expires.toUTCString()}; path=/; SameSite=Lax`
  }

  static removeToken(): void {
    if (typeof document === 'undefined') return

    document.cookie = `${TOKEN_KEY}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`
  }

  static async isAuthenticated(): Promise<boolean> {
    if (typeof window === 'undefined') {
      const token = await this.getServerToken()
      return !!token
    } else {
      const token = this.getClientToken()
      return !!token
    }
  }
}
