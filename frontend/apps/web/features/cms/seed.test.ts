import { describe, expect, it } from 'bun:test'
import { seedAll, validateRefs, type BrokenRef } from './seed'

type Doc = { id: number; [key: string]: unknown }
type Where = { [field: string]: { equals?: unknown } } | undefined

class FakePayload {
  private collections = new Map<string, Doc[]>()
  private globals = new Map<string, Record<string, unknown>>()
  private nextId = 1

  private store(collection: string): Doc[] {
    let docs = this.collections.get(collection)
    if (!docs) {
      docs = []
      this.collections.set(collection, docs)
    }
    return docs
  }

  async find(args: { collection: string; limit?: number; where?: Where }) {
    const docs = this.store(args.collection)
    const filtered = args.where
      ? docs.filter((d) =>
          Object.entries(args.where as Record<string, { equals?: unknown }>).every(
            ([k, v]) => d[k] === v.equals,
          ),
        )
      : docs
    return { docs: filtered, totalDocs: filtered.length }
  }

  async create(args: { collection: string; data: Record<string, unknown> }) {
    const doc: Doc = { id: this.nextId++, ...args.data }
    this.store(args.collection).push(doc)
    return doc
  }

  async update(args: { collection: string; id: number; data: Record<string, unknown> }) {
    const docs = this.store(args.collection)
    const idx = docs.findIndex((d) => d.id === args.id)
    if (idx === -1) throw new Error(`not found ${args.collection}#${args.id}`)
    const existing = docs[idx] as Doc
    const merged: Doc = { ...existing, ...args.data, id: existing.id }
    docs[idx] = merged
    return merged
  }

  async findByID(args: { collection: string; id: number | string }) {
    const docs = this.store(args.collection)
    const doc = docs.find((d) => d.id === args.id)
    if (!doc) throw new Error(`not found ${args.collection}#${args.id}`)
    return doc
  }

  async updateGlobal(args: { slug: string; data: Record<string, unknown> }) {
    this.globals.set(args.slug, args.data)
    return args.data
  }

  // Test helpers
  rawCollection(name: string): Doc[] {
    return this.store(name)
  }
}

describe('cms seed', () => {
  it('seedAll runs end-to-end without throwing and leaves zero broken refs', async () => {
    const payload = new FakePayload()
    // biome-ignore lint/suspicious/noExplicitAny: fake payload duck-types the surface seed.ts uses
    await seedAll(payload as any)
    // biome-ignore lint/suspicious/noExplicitAny: same as above
    const broken = await validateRefs(payload as any)
    expect(broken).toEqual([])
  })

  it('validateRefs flags pages whose block-ref points to a missing block-library doc', async () => {
    const payload = new FakePayload()
    payload.rawCollection('pages').push({
      id: 999,
      slug: 'broken',
      blocks: [{ blockType: 'block-ref', ref: 12345 }],
    })

    // biome-ignore lint/suspicious/noExplicitAny: see above
    const broken: BrokenRef[] = await validateRefs(payload as any)
    expect(broken).toHaveLength(1)
    expect(broken[0]).toMatchObject({ pageSlug: 'broken', missingRefId: 12345 })
  })

  it('validateRefs flags pages whose block-ref has empty ref', async () => {
    const payload = new FakePayload()
    payload.rawCollection('pages').push({
      id: 1,
      slug: 'empty-ref',
      blocks: [{ blockType: 'block-ref' }],
    })

    // biome-ignore lint/suspicious/noExplicitAny: see above
    const broken = await validateRefs(payload as any)
    expect(broken).toHaveLength(1)
    expect(broken[0]?.missingRefId).toBe('<empty>')
  })
})
