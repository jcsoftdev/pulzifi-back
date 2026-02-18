'use client'

import { useCallback, useEffect, useRef, useState } from 'react'

interface VisualPulseProps {
  currentScreenshotUrl?: string
  previousScreenshotUrl?: string
}

export function VisualPulse({
  currentScreenshotUrl,
  previousScreenshotUrl,
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

  const handleMouseDown = useCallback(() => {
    setIsResizing(true)
  }, [])

  const handleMouseUp = useCallback(() => {
    setIsResizing(false)
  }, [])

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isResizing || !containerRef.current) return

      const rect = containerRef.current.getBoundingClientRect()
      const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width))
      const percentage = (x / rect.width) * 100
      setSliderPosition(percentage)
    },
    [
      isResizing,
    ]
  )

  useEffect(() => {
    if (isResizing) {
      globalThis.addEventListener('mousemove', handleMouseMove)
      globalThis.addEventListener('mouseup', handleMouseUp)
    } else {
      globalThis.removeEventListener('mousemove', handleMouseMove)
      globalThis.removeEventListener('mouseup', handleMouseUp)
    }

    return () => {
      globalThis.removeEventListener('mousemove', handleMouseMove)
      globalThis.removeEventListener('mouseup', handleMouseUp)
    }
  }, [
    isResizing,
    handleMouseMove,
    handleMouseUp,
  ])

  if (!currentScreenshotUrl) {
    return (
      <div className="w-full h-full min-h-[500px] flex items-center justify-center bg-muted/10 rounded-lg border border-border">
        <div className="text-center space-y-2">
          <p className="text-muted-foreground">No screenshot available.</p>
        </div>
      </div>
    )
  }

  if (!previousScreenshotUrl) {
    return (
      <div className="w-full h-full min-h-[500px] flex items-center justify-center bg-muted/10 rounded-lg border border-border">
        <div className="text-center space-y-2">
          <p className="text-muted-foreground">No previous check available for comparison.</p>
          <div className="relative w-full max-w-4xl mx-auto border border-border rounded-lg overflow-hidden shadow-sm">
            {/* biome-ignore lint/performance/noImgElement: screenshot URLs are dynamic external URLs with unknown dimensions */}
            <img
              src={currentScreenshotUrl}
              alt="Current snapshot"
              className="w-full h-auto object-contain"
            />
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
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
        {/* Previous Image (Background) — flow layout until height computed, then absolute */}
        {/* biome-ignore lint/performance/noImgElement: screenshot URLs are dynamic external URLs; ref + naturalWidth needed for height computation */}
        <img
          ref={prevImgRef}
          src={previousScreenshotUrl}
          alt="Previous snapshot"
          onLoad={recompute}
          className={
            containerHeight
              ? 'absolute inset-0 w-full h-full object-contain block object-top'
              : 'w-full h-auto object-contain block'
          }
        />

        {/* Current Image (Foreground - Clipped) */}
        <div
          className="absolute top-0 left-0 h-full overflow-hidden border-r border-primary/50"
          style={{
            width: `${sliderPosition}%`,
          }}
        >
          {/* biome-ignore lint/performance/noImgElement: foreground image must span beyond clipping container — incompatible with next/image fill */}
          <img
            ref={currImgRef}
            src={currentScreenshotUrl}
            alt="Current snapshot"
            onLoad={recompute}
            className="absolute top-0 left-0 h-full max-w-none object-contain object-top"
            style={{
              width: containerWidth || '100%',
            }}
          />
        </div>

        {/* Slider Handle */}
        <button
          type="button"
          aria-label="Drag to compare images"
          className="absolute top-0 bottom-0 w-1 bg-primary cursor-ew-resize z-10 flex items-center justify-center hover:bg-primary/90 transition-colors border-0 p-0"
          style={{
            left: `${sliderPosition}%`,
            transform: 'translateX(-50%)',
          }}
          onMouseDown={handleMouseDown}
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
