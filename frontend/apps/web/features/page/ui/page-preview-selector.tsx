'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { X } from 'lucide-react'
import type { PreviewElement, SelectorOffsets } from '../domain/types'
import { getSectionColor } from '../domain/section-colors'

export interface ElementSelection {
  cssSelector: string
  xpathSelector: string
  rect: { x: number; y: number; w: number; h: number }
  offsets: SelectorOffsets
  textPreview: string
  name: string
}

interface PagePreviewSelectorProps {
  screenshotBase64: string
  viewport: { width: number; height: number }
  pageHeight: number
  elements: PreviewElement[]
  /** Called in single-select mode (legacy). */
  onSelect?: (selection: ElementSelection | null) => void
  /** Called in multi-select mode with the full list of selections. */
  onMultiSelect?: (selections: ElementSelection[]) => void
  selectedSelector?: string
  /** Enable multi-select mode. Default false for backward compatibility. */
  multiSelect?: boolean
}


/** Build a descriptive default name from element metadata. */
function buildSectionName(el: PreviewElement, order: number): string {
  // 1. Use semantic role if it's meaningful
  if (el.semantic_role && el.semantic_role !== 'generic') {
    // Capitalize: "navigation" → "Navigation"
    const label = el.semantic_role.charAt(0).toUpperCase() + el.semantic_role.slice(1)
    return label.replace(/_/g, ' ')
  }

  // 2. Try to extract a readable label from text_preview (first meaningful words)
  if (el.text_preview) {
    const cleaned = el.text_preview.trim().replace(/\s+/g, ' ')
    if (cleaned.length > 0) {
      // Take first ~30 chars, cut at word boundary
      const snippet = cleaned.length <= 30 ? cleaned : `${cleaned.slice(0, 30).replace(/\s\S*$/, '')}…`
      return snippet
    }
  }

  // 3. Try to derive a name from the CSS selector (class or id)
  const idMatch = el.selector.match(/#([\w-]+)/)
  if (idMatch) {
    // "#hero-banner" → "Hero banner"
    return (idMatch[1] ?? '').replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
  }
  const classMatch = el.selector.match(/\.([\w-]+)(?:[:.[]|$)/)
  if (classMatch) {
    return (classMatch[1] ?? '').replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
  }

  // 4. Use the HTML tag as a hint
  const tagLabels: Record<string, string> = {
    header: 'Header', footer: 'Footer', nav: 'Navigation', main: 'Main Content',
    aside: 'Sidebar', section: 'Section', article: 'Article', form: 'Form',
    table: 'Table', ul: 'List', ol: 'List', figure: 'Figure', img: 'Image',
  }
  if (tagLabels[el.tag]) {
    return tagLabels[el.tag]!
  }

  // 5. Fallback
  return `Section ${order + 1}`
}

export function PagePreviewSelector({
  screenshotBase64,
  viewport,
  pageHeight,
  elements,
  onSelect,
  onMultiSelect,
  selectedSelector,
  multiSelect = false,
}: Readonly<PagePreviewSelectorProps>) {
  const wrapperRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)
  const [containerWidth, setContainerWidth] = useState(0)

  // Dynamically measure the available width so the preview fills its parent
  useEffect(() => {
    const el = wrapperRef.current
    if (!el) return
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerWidth(Math.floor(entry.contentRect.width))
      }
    })
    ro.observe(el)
    setContainerWidth(el.clientWidth)
    return () => ro.disconnect()
  }, [])

  // Single-select state (legacy)
  const [selectedIndex, setSelectedIndex] = useState<number | null>(() => {
    if (!selectedSelector) return null
    const idx = elements.findIndex((el) => el.selector === selectedSelector)
    return idx >= 0 ? idx : null
  })
  const [offsets, setOffsets] = useState<SelectorOffsets>({ top: 0, right: 0, bottom: 0, left: 0 })

  // Multi-select state
  const [selectedIndices, setSelectedIndices] = useState<number[]>([])
  const [multiOffsets, setMultiOffsets] = useState<Map<number, SelectorOffsets>>(new Map())
  const [sectionNames, setSectionNames] = useState<Map<number, string>>(new Map())

  const scale = containerWidth > 0 ? containerWidth / viewport.width : 1
  const scaledHeight = pageHeight * scale

  // Build and emit multi-select selections
    const emitMultiSelections = useCallback(
    (indices: number[], names: Map<number, string>, offsetsMap: Map<number, SelectorOffsets>) => {
      if (!onMultiSelect) return
      const selections: ElementSelection[] = indices.map((idx, i) => {
        const el = elements[idx]!
        return {
          cssSelector: el.selector,
          xpathSelector: el.xpath,
          rect: el.rect,
          offsets: offsetsMap.get(idx) ?? { top: 0, right: 0, bottom: 0, left: 0 },
          textPreview: el.text_preview,
          name: names.get(idx) ?? buildSectionName(el, i),
        }
      })
      onMultiSelect(selections)
    },
    [elements, onMultiSelect]
  )

  const handleElementClick = useCallback(
    (index: number) => {
      if (multiSelect) {
        // Compute all next state upfront — never call setState or parent callbacks
        // inside another setState updater (causes "setState during render" error).
        const nextIndices = selectedIndices.includes(index)
          ? selectedIndices.filter((i) => i !== index)
          : [...selectedIndices, index]

        const nextNames = new Map(sectionNames)
        if (!nextNames.has(index) && nextIndices.includes(index)) {
          const el = elements[index]
          if (el) nextNames.set(index, buildSectionName(el, nextIndices.length - 1))
        }

        const nextOffsets = new Map(multiOffsets)
        if (!nextOffsets.has(index)) {
          nextOffsets.set(index, { top: 0, right: 0, bottom: 0, left: 0 })
        }

        // Apply all state updates, then notify parent
        setSelectedIndices(nextIndices)
        setSectionNames(nextNames)
        setMultiOffsets(nextOffsets)
        emitMultiSelections(nextIndices, nextNames, nextOffsets)
      } else {
        // Single-select (legacy)
        if (selectedIndex === index) {
          setSelectedIndex(null)
          setOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
          onSelect?.(null)
          return
        }
        setSelectedIndex(index)
        const el = elements[index]
        if (!el) return
        const newOffsets = { top: 0, right: 0, bottom: 0, left: 0 }
        setOffsets(newOffsets)
        onSelect?.({
          cssSelector: el.selector,
          xpathSelector: el.xpath,
          rect: el.rect,
          offsets: newOffsets,
          textPreview: el.text_preview,
          name: '',
        })
      }
    },
    [elements, onSelect, selectedIndex, selectedIndices, multiSelect, sectionNames, multiOffsets, emitMultiSelections]
  )

  const handleOffsetChange = useCallback(
    (side: keyof SelectorOffsets, value: number) => {
      if (multiSelect) return // offsets handled per-section below
      const newOffsets = { ...offsets, [side]: value }
      setOffsets(newOffsets)
      if (selectedIndex !== null) {
        const el = elements[selectedIndex]
        if (!el) return
        onSelect?.({
          cssSelector: el.selector,
          xpathSelector: el.xpath,
          rect: el.rect,
          offsets: newOffsets,
          textPreview: el.text_preview,
          name: '',
        })
      }
    },
    [offsets, selectedIndex, elements, onSelect, multiSelect]
  )

  const handleRemoveSection = useCallback(
    (index: number) => {
      const next = selectedIndices.filter((i) => i !== index)
      setSelectedIndices(next)
      emitMultiSelections(next, sectionNames, multiOffsets)
    },
    [selectedIndices, sectionNames, multiOffsets, emitMultiSelections]
  )

  const handleSectionNameChange = useCallback(
    (index: number, name: string) => {
      const next = new Map(sectionNames)
      next.set(index, name)
      setSectionNames(next)
      emitMultiSelections(selectedIndices, next, multiOffsets)
    },
    [selectedIndices, sectionNames, multiOffsets, emitMultiSelections]
  )

  const selectedElement = selectedIndex !== null ? elements[selectedIndex] : null
  const hoveredElement = hoveredIndex !== null ? elements[hoveredIndex] : null

  const visibleElements = useMemo(
    () =>
      elements
        .filter((el) => {
          const scaledW = el.rect.w * scale
          const scaledH = el.rect.h * scale
          return scaledW >= 30 && scaledH >= 15
        })
        .sort((a, b) => b.rect.w * b.rect.h - a.rect.w * a.rect.h),
    [elements, scale]
  )

  const isSelected = (realIdx: number) =>
    multiSelect ? selectedIndices.includes(realIdx) : selectedIndex === realIdx

  const getSelectionOrder = (realIdx: number) =>
    multiSelect ? selectedIndices.indexOf(realIdx) : -1

  return (
    <div ref={wrapperRef} className="flex flex-col gap-3 w-full flex-1 min-h-[55dvh] md:min-h-0 md:h-full">
      <div
        ref={containerRef}
        className="relative border border-border rounded-lg overflow-auto bg-muted/50 w-full flex-1 min-h-0"
        style={{ WebkitOverflowScrolling: 'touch' } as React.CSSProperties}
      >
        {containerWidth > 0 && (
        <div className="relative" style={{ width: containerWidth }}>
          <img
            src={`data:image/webp;base64,${screenshotBase64}`}
            alt="Page preview"
            className="block w-full h-auto"
            draggable={false}
          />

          {visibleElements.map((el, visibleIdx) => {
            const realIdx = elements.indexOf(el)
            const isHovered = hoveredIndex === realIdx
            const selected = isSelected(realIdx)
            const order = getSelectionOrder(realIdx)
            const color = multiSelect && selected ? getSectionColor(order) : 'rgb(59 130 246)'

            const area = el.rect.w * el.rect.h
            const baseZ = Math.max(1, Math.round(1000000 / area))

            return (
              <button
                type="button"
                key={visibleIdx}
                className="absolute cursor-pointer transition-all duration-100 p-0 bg-transparent"
                style={{
                  left: el.rect.x * scale,
                  top: el.rect.y * scale,
                  width: el.rect.w * scale,
                  height: el.rect.h * scale,
                  border: selected
                    ? `2px solid ${color}`
                    : isHovered
                      ? '2px solid rgb(59 130 246 / 0.6)'
                      : '1px solid transparent',
                  backgroundColor: selected
                    ? `${color}19` // ~10% opacity
                    : isHovered
                      ? 'rgb(59 130 246 / 0.05)'
                      : 'transparent',
                  zIndex: baseZ,
                  // pan-y: let vertical swipes pass through to the scroll container (iOS)
                  // while still registering taps on this button.
                  touchAction: 'pan-y',
                }}
                onMouseEnter={() => setHoveredIndex(realIdx)}
                onMouseLeave={() => setHoveredIndex(null)}
                onClick={() => handleElementClick(realIdx)}
                aria-label={`Select element: ${el.tag} - ${el.text_preview?.slice(0, 40) || 'no text'}`}
              >
                {/* Badge number for multi-select */}
                {multiSelect && selected && (
                  <span
                    className="absolute -top-2 -left-2 w-5 h-5 rounded-full text-white text-[10px] font-bold flex items-center justify-center"
                    style={{ backgroundColor: color, zIndex: baseZ + 1 }}
                  >
                    {order + 1}
                  </span>
                )}
                {/* Hover tooltip: tag + first class */}
                {isHovered && (
                  <span
                    className="absolute bottom-full left-0 mb-1 px-1.5 py-0.5 rounded text-[10px] font-mono text-white whitespace-nowrap pointer-events-none"
                    style={{ backgroundColor: 'rgba(0,0,0,0.75)', zIndex: baseZ + 2 }}
                  >
                    {el.tag}
                    {el.selector.startsWith('#')
                      ? el.selector
                      : el.selector.match(/\.([\w-]+)/)?.[0] ?? ''}
                  </span>
                )}
              </button>
            )
          })}

          {/* Single-select offset visualization */}
          {!multiSelect && selectedElement && (
            <div
              className="absolute pointer-events-none border-2 border-dashed border-blue-400/50"
              style={{
                left: (selectedElement.rect.x - offsets.left) * scale,
                top: (selectedElement.rect.y - offsets.top) * scale,
                width: (selectedElement.rect.w + offsets.left + offsets.right) * scale,
                height: (selectedElement.rect.h + offsets.top + offsets.bottom) * scale,
                zIndex: 25,
              }}
            />
          )}

          {/* Multi-select offset visualizations */}
          {multiSelect &&
            selectedIndices.map((idx, i) => {
              const el = elements[idx]
              if (!el) return null
              const off = multiOffsets.get(idx) ?? { top: 0, right: 0, bottom: 0, left: 0 }
              const order = selectedIndices.indexOf(idx)
              const color = getSectionColor(order)
              return (
                <div
                  key={`offset-${i}`}
                  className="absolute pointer-events-none border-2 border-dashed"
                  style={{
                    borderColor: `${color}80`,
                    left: (el.rect.x - off.left) * scale,
                    top: (el.rect.y - off.top) * scale,
                    width: (el.rect.w + off.left + off.right) * scale,
                    height: (el.rect.h + off.top + off.bottom) * scale,
                    zIndex: 25,
                  }}
                />
              )
            })}
        </div>
        )}
      </div>

      {/* Element info tooltip */}
      {!multiSelect && (hoveredElement || selectedElement) && (
        <div className="text-xs text-muted-foreground bg-muted rounded px-3 py-2">
          <span className="font-mono text-foreground">
            {(selectedElement ?? hoveredElement)?.tag}
          </span>
          {' — '}
          <span>{(selectedElement ?? hoveredElement)?.text_preview?.slice(0, 80) || 'No text content'}</span>
          {(selectedElement ?? hoveredElement)?.semantic_role !== 'generic' && (
            <span className="ml-1 text-blue-500">
              [{(selectedElement ?? hoveredElement)?.semantic_role}]
            </span>
          )}
        </div>
      )}

      {/* Single-select offset controls */}
      {!multiSelect && selectedElement && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Expand selection area (px)</p>
          <div className="grid grid-cols-4 gap-2">
            {(['top', 'right', 'bottom', 'left'] as const).map((side) => (
              <div key={side} className="space-y-1">
                <label htmlFor={`offset-${side}`} className="text-[10px] text-muted-foreground capitalize">{side}</label>
                <input
                  id={`offset-${side}`}
                  type="number"
                  min={0}
                  max={200}
                  value={offsets[side]}
                  onChange={(e) => handleOffsetChange(side, Number(e.target.value) || 0)}
                  className="w-full h-7 px-2 text-xs rounded border border-border bg-background"
                />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Multi-select: Selected sections list */}
      {multiSelect && selectedIndices.length > 0 && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">
            Selected sections ({selectedIndices.length})
          </p>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {selectedIndices.map((idx, i) => {
              const el = elements[idx]
              if (!el) return null
              const color = getSectionColor(i)
              return (
                <div
                  key={idx}
                  className="flex items-center gap-2 bg-muted rounded px-3 py-2 text-xs"
                >
                  <span
                    className="w-5 h-5 rounded-full text-white text-[10px] font-bold flex items-center justify-center shrink-0"
                    style={{ backgroundColor: color }}
                  >
                    {i + 1}
                  </span>
                  <input
                    type="text"
                    value={sectionNames.get(idx) ?? `Section ${i + 1}`}
                    onChange={(e) => handleSectionNameChange(idx, e.target.value)}
                    className="flex-1 h-6 px-2 text-xs rounded border border-border bg-background min-w-0"
                    placeholder="Section name"
                  />
                  <span className="text-muted-foreground font-mono truncate max-w-[120px]" title={el.selector}>
                    {el.tag}
                  </span>
                  <button
                    type="button"
                    onClick={() => handleRemoveSection(idx)}
                    className="p-0.5 hover:bg-destructive/10 rounded text-muted-foreground hover:text-destructive"
                    aria-label={`Remove section ${i + 1}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
