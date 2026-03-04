'use client'

import { useCallback, useMemo, useRef, useState } from 'react'
import type { PreviewElement, SelectorOffsets } from '../domain/types'

export interface ElementSelection {
  cssSelector: string
  xpathSelector: string
  rect: { x: number; y: number; w: number; h: number }
  offsets: SelectorOffsets
  textPreview: string
}

interface PagePreviewSelectorProps {
  screenshotBase64: string
  viewport: { width: number; height: number }
  pageHeight: number
  elements: PreviewElement[]
  onSelect: (selection: ElementSelection | null) => void
  selectedSelector?: string
}

export function PagePreviewSelector({
  screenshotBase64,
  viewport,
  pageHeight,
  elements,
  onSelect,
  selectedSelector,
}: Readonly<PagePreviewSelectorProps>) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)
  const [selectedIndex, setSelectedIndex] = useState<number | null>(() => {
    if (!selectedSelector) return null
    const idx = elements.findIndex((el) => el.selector === selectedSelector)
    return idx >= 0 ? idx : null
  })
  const [offsets, setOffsets] = useState<SelectorOffsets>({ top: 0, right: 0, bottom: 0, left: 0 })

  // Scale factor: fit the screenshot width into the container
  const containerWidth = 580 // max width of the preview area
  const scale = containerWidth / viewport.width
  const scaledHeight = pageHeight * scale

  const handleElementClick = useCallback(
    (index: number) => {
      if (selectedIndex === index) {
        // Deselect
        setSelectedIndex(null)
        setOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
        onSelect(null)
        return
      }

      setSelectedIndex(index)
      const el = elements[index]
      if (!el) return
      const newOffsets = { top: 0, right: 0, bottom: 0, left: 0 }
      setOffsets(newOffsets)
      onSelect({
        cssSelector: el.selector,
        xpathSelector: el.xpath,
        rect: el.rect,
        offsets: newOffsets,
        textPreview: el.text_preview,
      })
    },
    [elements, onSelect, selectedIndex]
  )

  const handleOffsetChange = useCallback(
    (side: keyof SelectorOffsets, value: number) => {
      const newOffsets = { ...offsets, [side]: value }
      setOffsets(newOffsets)
      if (selectedIndex !== null) {
        const el = elements[selectedIndex]
        if (!el) return
        onSelect({
          cssSelector: el.selector,
          xpathSelector: el.xpath,
          rect: el.rect,
          offsets: newOffsets,
          textPreview: el.text_preview,
        })
      }
    },
    [offsets, selectedIndex, elements, onSelect]
  )

  const selectedElement = selectedIndex !== null ? elements[selectedIndex] : null
  const hoveredElement = hoveredIndex !== null ? elements[hoveredIndex] : null

  // Filter to elements that are large enough to be meaningful at this scale
  const visibleElements = useMemo(
    () =>
      elements
        .filter((el) => {
          const scaledW = el.rect.w * scale
          const scaledH = el.rect.h * scale
          return scaledW >= 30 && scaledH >= 15
        })
        .sort((a, b) => b.rect.w * b.rect.h - a.rect.w * a.rect.h), // largest first → children render later → sit on top
    [elements, scale]
  )

  return (
    <div className="space-y-3">
      <div
        ref={containerRef}
        className="relative border border-border rounded-lg overflow-auto bg-muted/50"
        style={{ maxHeight: 400, width: containerWidth }}
      >
        {/* Screenshot */}
        <div className="relative" style={{ width: containerWidth, height: scaledHeight }}>
          <img
            src={`data:image/png;base64,${screenshotBase64}`}
            alt="Page preview"
            className="block"
            style={{ width: containerWidth, height: scaledHeight }}
            draggable={false}
          />

          {/* Element overlays */}
          {visibleElements.map((el, idx) => {
            const realIdx = elements.indexOf(el)
            const isHovered = hoveredIndex === realIdx
            const isSelected = selectedIndex === realIdx

            const area = el.rect.w * el.rect.h
            const baseZ = Math.max(1, Math.round(1000000 / area)) // smaller area → higher z

            return (
              <div
                key={el.selector}
                className="absolute cursor-pointer transition-all duration-100"
                style={{
                  left: el.rect.x * scale,
                  top: el.rect.y * scale,
                  width: el.rect.w * scale,
                  height: el.rect.h * scale,
                  border: isSelected
                    ? '2px solid rgb(59 130 246)'
                    : isHovered
                      ? '2px solid rgb(59 130 246 / 0.6)'
                      : '1px solid transparent',
                  backgroundColor: isSelected
                    ? 'rgb(59 130 246 / 0.1)'
                    : isHovered
                      ? 'rgb(59 130 246 / 0.05)'
                      : 'transparent',
                  zIndex: baseZ,
                }}
                onMouseEnter={() => setHoveredIndex(realIdx)}
                onMouseLeave={() => setHoveredIndex(null)}
                onClick={() => handleElementClick(realIdx)}
              />
            )
          })}

          {/* Selected element offset visualization */}
          {selectedElement && (
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
        </div>
      </div>

      {/* Element info tooltip */}
      {(hoveredElement || selectedElement) && (
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

      {/* Offset controls when element is selected */}
      {selectedElement && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Expand selection area (px)</p>
          <div className="grid grid-cols-4 gap-2">
            {(['top', 'right', 'bottom', 'left'] as const).map((side) => (
              <div key={side} className="space-y-1">
                <label className="text-[10px] text-muted-foreground capitalize">{side}</label>
                <input
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
    </div>
  )
}
