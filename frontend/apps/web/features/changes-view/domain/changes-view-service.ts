import { type Check, type Insight, PageApi } from '@workspace/services/page-api'

export const ChangesViewService = {
  async getPageChecks(pageId: string): Promise<Check[]> {
    return PageApi.listChecks(pageId)
  },

  async getPageInsights(pageId: string, checkId?: string): Promise<Insight[]> {
    return PageApi.listInsights(pageId, checkId)
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

    for (const el of doc.querySelectorAll('script, style, noscript, link, meta')) {
      el.remove()
    }

    // Block-level tags that should produce paragraph breaks in the output
    const BLOCK_TAGS = new Set([
      'P', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6',
      'LI', 'TD', 'TH', 'BLOCKQUOTE', 'PRE',
      'DIV', 'ARTICLE', 'SECTION', 'HEADER',
      'FOOTER', 'NAV', 'MAIN', 'ASIDE', 'FIGCAPTION', 'TR',
    ])

    // Recursively walk the DOM, wrapping block elements with newlines
    function walk(node: Node): string {
      if (node.nodeType === Node.TEXT_NODE) return node.textContent ?? ''
      if (node.nodeType !== Node.ELEMENT_NODE) return ''
      const el = node as Element
      const inner = Array.from(el.childNodes).map(walk).join('')
      return BLOCK_TAGS.has(el.tagName) ? `\n${inner}\n` : inner
    }

    // Each non-empty line becomes one diffable paragraph
    return walk(doc.body)
      .split('\n')
      .map((l) => l.replace(/\s+/g, ' ').trim())
      .filter(Boolean)
      .join('\n')
  },
}
