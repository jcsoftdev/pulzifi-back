import { PageApi, type Check, type Insight } from '@workspace/services/page-api'

export const ChangesViewService = {
  async getPageChecks(pageId: string): Promise<Check[]> {
    return PageApi.listChecks(pageId)
  },

  async getPageInsights(pageId: string): Promise<Insight[]> {
    return PageApi.listInsights(pageId)
  },

  async getPage(pageId: string) {
    return PageApi.getPage(pageId)
  },

  async getHtmlContent(url: string): Promise<string> {
    try {
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error('Failed to fetch HTML content')
      }
      return await response.text()
    } catch (error) {
      console.error('Error fetching HTML:', error)
      return ''
    }
  },

  extractTextFromHtml(html: string): string {
    if (typeof window === 'undefined') return html // Server side fallback
    const parser = new DOMParser()
    const doc = parser.parseFromString(html, 'text/html')
    return doc.body.innerText || doc.body.textContent || ''
  }
}
