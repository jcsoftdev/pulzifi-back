'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import type { SectionRect } from '@workspace/services/page-api'

interface SectionOverlay {
  name: string
  rect: SectionRect
  viewportWidth: number
  color: string
}

interface VisualPulseProps {
  currentScreenshotUrl?: string
  previousScreenshotUrl?: string
  sectionOverlay?: SectionOverlay
  /** @deprecated Use sectionOverlay instead */
  sectionName?: string
}

export function VisualPulse({
  currentScreenshotUrl,
  previousScreenshotUrl,
  sectionOverlay,
  sectionName,
}: Readonly<VisualPulseProps>) {
  const [sliderPosition, setSliderPosition] = useState(50)
  const [isResizing, setIsResizing] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const prevImgRef = useRef<HTMLImageElement>(null)
  const currImgRef = useRef<HTMLImageElement>(null)
  const [containerWidth, setContainerWidth] = useState(0)
  const [containerHeight, setContainerHeight] = useState(0)

  const recompute = useCallback(() => {
    if (!containerRef.current) return
    const w = containerRef.current.offsetWidth
    if (!w) return
    setContainerWidth(w)

    let maxH = 0
    const prev = prevImgRef.current
    const curr = currImgRef.current
    if (prev?.naturalWidth) maxH = Math.max(maxH, (prev.naturalHeight / prev.naturalWidth) * w)
    if (curr?.naturalWidth) maxH = Math.max(maxH, (curr.naturalHeight / curr.naturalWidth) * w)
    if (maxH > 0) setContainerHeight(maxH)
  }, [])

  useEffect(() => {
    if (!containerRef.current) return
    const el = containerRef.current
    const observer = new ResizeObserver(recompute)
    observer.observe(el)
    recompute()
    return () => observer.disconnect()
  }, [
    recompute,
  ])

  const updatePosition = useCallback((clientX: number) => {
    if (!containerRef.current) return
    const rect = containerRef.current.getBoundingClientRect()
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width))
    setSliderPosition((x / rect.width) * 100)
  }, [])

  const handleMouseDown = useCallback(() => {
    setIsResizing(true)
  }, [])

  const handleMouseUp = useCallback(() => {
    setIsResizing(false)
  }, [])

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isResizing) return
      updatePosition(e.clientX)
    },
    [isResizing, updatePosition]
  )

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    e.preventDefault()
    setIsResizing(true)
  }, [])

  const handleTouchEnd = useCallback(() => {
    setIsResizing(false)
  }, [])

  const handleTouchMove = useCallback(
    (e: TouchEvent) => {
      if (!isResizing || !e.touches[0]) return
      e.preventDefault()
      updatePosition(e.touches[0].clientX)
    },
    [isResizing, updatePosition]
  )

  useEffect(() => {
    if (isResizing) {
      globalThis.addEventListener('mousemove', handleMouseMove)
      globalThis.addEventListener('mouseup', handleMouseUp)
      globalThis.addEventListener('touchmove', handleTouchMove, { passive: false })
      globalThis.addEventListener('touchend', handleTouchEnd)
    } else {
      globalThis.removeEventListener('mousemove', handleMouseMove)
      globalThis.removeEventListener('mouseup', handleMouseUp)
      globalThis.removeEventListener('touchmove', handleTouchMove)
      globalThis.removeEventListener('touchend', handleTouchEnd)
    }

    return () => {
      globalThis.removeEventListener('mousemove', handleMouseMove)
      globalThis.removeEventListener('mouseup', handleMouseUp)
      globalThis.removeEventListener('touchmove', handleTouchMove)
      globalThis.removeEventListener('touchend', handleTouchEnd)
    }
  }, [isResizing, handleMouseMove, handleMouseUp, handleTouchMove, handleTouchEnd])

  // Compute overlay geometry once containerWidth is known
  const overlayStyle = sectionOverlay && containerWidth > 0
    ? (() => {
        const { rect, viewportWidth } = sectionOverlay
        const scale = containerWidth / viewportWidth
        return {
          left: rect.x * scale,
          top: rect.y * scale,
          width: rect.w * scale,
          height: rect.h * scale,
        }
      })()
    : null

  // Section label (from overlay or legacy prop)
  const displaySectionName = sectionOverlay?.name ?? sectionName

  if (!currentScreenshotUrl) {
    return (
      <div className="w-full min-h-[400px] flex flex-col items-center justify-center gap-4 bg-muted/5 rounded-xl border border-dashed border-border">
        <div className="w-12 h-12 rounded-full bg-muted/30 flex items-center justify-center">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="text-muted-foreground" aria-hidden="true">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <circle cx="9" cy="9" r="2" />
            <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21" />
          </svg>
        </div>
        <div className="text-center">
          <p className="text-sm font-medium text-foreground">No snapshot yet</p>
          <p className="text-xs text-muted-foreground mt-1">Select a detected change from the dropdown above</p>
        </div>
      </div>
    )
  }

  if (!previousScreenshotUrl) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800/50 px-4 py-2.5 rounded-lg">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="shrink-0" aria-hidden="true">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          <span>No previous snapshot to compare — this is the first recorded check.</span>
        </div>
        <div className="relative w-full max-w-4xl mx-auto border border-border rounded-xl overflow-hidden shadow-sm">
          {/* biome-ignore lint/performance/noImgElement: screenshot URLs are dynamic external URLs with unknown dimensions */}
          <img
            src={currentScreenshotUrl}
            alt="Current snapshot"
            className="w-full h-auto object-contain"
          />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Section indicator — text badge when no rect available (legacy), hidden when overlay handles it */}
      {displaySectionName && !overlayStyle && (
        <div className="flex items-center gap-2 text-sm bg-blue-50 dark:bg-blue-950/30 text-blue-700 dark:text-blue-300 px-4 py-2.5 rounded-lg border border-blue-200 dark:border-blue-800">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="shrink-0" aria-hidden="true">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
            <line x1="3" y1="9" x2="21" y2="9" />
          </svg>
          <span>
            Comparing section: <strong>{displaySectionName}</strong>
          </span>
        </div>
      )}
      <div
        ref={containerRef}
        className="relative w-full select-none overflow-hidden rounded-lg border border-border shadow-sm bg-muted/10"
        style={
          containerHeight
            ? {
                height: containerHeight,
              }
            : undefined
        }
      >
        {/* Current Image (Background) — flow layout until height computed, then absolute */}
        {/* biome-ignore lint/performance/noImgElement: screenshot URLs are dynamic external URLs; ref + naturalWidth needed for height computation */}
        <img
          ref={currImgRef}
          src={currentScreenshotUrl}
          alt="Current snapshot"
          onLoad={recompute}
          className={
            containerHeight
              ? 'absolute inset-0 w-full h-full object-contain block object-top'
              : 'w-full h-auto object-contain block'
          }
        />

        {/* Section boundary overlay on current image */}
        {overlayStyle && sectionOverlay && (
          <div
            className="absolute pointer-events-none rounded-sm"
            style={{
              left: overlayStyle.left,
              top: overlayStyle.top,
              width: overlayStyle.width,
              height: overlayStyle.height,
              border: `2px solid ${sectionOverlay.color}`,
              backgroundColor: `${sectionOverlay.color}18`,
              zIndex: 5,
            }}
          >
            <span
              className="absolute top-1 left-1 px-1.5 py-0.5 rounded text-[10px] font-medium text-white"
              style={{ backgroundColor: sectionOverlay.color }}
            >
              {sectionOverlay.name}
            </span>
          </div>
        )}

        {/* Previous Image (Foreground - Clipped) */}
        <div
          className="absolute top-0 left-0 h-full overflow-hidden border-r border-primary/50"
          style={{
            width: `${sliderPosition}%`,
          }}
        >
          {/* biome-ignore lint/performance/noImgElement: foreground image must span beyond clipping container — incompatible with next/image fill */}
          <img
            ref={prevImgRef}
            src={previousScreenshotUrl}
            alt="Previous snapshot"
            onLoad={recompute}
            className="absolute top-0 left-0 h-full max-w-none object-contain object-top"
            style={{
              width: containerWidth || '100%',
            }}
          />

          {/* Section boundary overlay on previous image (inside clip container, same absolute coords) */}
          {overlayStyle && sectionOverlay && (
            <div
              className="absolute pointer-events-none rounded-sm"
              style={{
                left: overlayStyle.left,
                top: overlayStyle.top,
                width: overlayStyle.width,
                height: overlayStyle.height,
                border: `2px solid ${sectionOverlay.color}`,
                backgroundColor: `${sectionOverlay.color}18`,
                zIndex: 5,
              }}
            />
          )}
        </div>

        {/* Slider Handle */}
        <button
          type="button"
          aria-label="Drag to compare images"
          className="absolute top-0 bottom-0 w-1 bg-primary cursor-ew-resize z-10 flex items-center justify-center hover:bg-primary/90 transition-colors border-0 p-0 touch-none"
          style={{
            left: `${sliderPosition}%`,
            transform: 'translateX(-50%)',
          }}
          onMouseDown={handleMouseDown}
          onTouchStart={handleTouchStart}
        >
          <div className="w-8 h-16 bg-primary rounded-lg flex items-center justify-center shadow-lg">
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="text-primary-foreground"
              aria-hidden="true"
              focusable="false"
            >
              <path d="m9 18 6-6-6-6" />
            </svg>
          </div>
        </button>
      </div>
    </div>
  )
}
