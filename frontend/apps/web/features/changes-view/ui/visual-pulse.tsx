'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { cn } from '@workspace/ui/lib/utils'

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
  const [containerWidth, setContainerWidth] = useState<number>(0)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!containerRef.current) return
    
    const updateWidth = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.offsetWidth)
      }
    }

    updateWidth()
    window.addEventListener('resize', updateWidth)
    
    // Also use ResizeObserver for better robustness
    const observer = new ResizeObserver(updateWidth)
    observer.observe(containerRef.current)

    return () => {
      window.removeEventListener('resize', updateWidth)
      observer.disconnect()
    }
  }, [])

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
    [isResizing]
  )

  useEffect(() => {
    if (isResizing) {
      window.addEventListener('mousemove', handleMouseMove)
      window.addEventListener('mouseup', handleMouseUp)
    } else {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isResizing, handleMouseMove, handleMouseUp])

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
      >
        {/* Previous Image (Background) */}
        <img
          src={previousScreenshotUrl}
          alt="Previous snapshot"
          className="w-full h-auto object-contain block"
        />

        {/* Current Image (Foreground - Clipped) */}
        <div
          className="absolute top-0 left-0 h-full overflow-hidden border-r border-primary/50"
          style={{ width: `${sliderPosition}%` }}
        >
          <img
            src={currentScreenshotUrl}
            alt="Current snapshot"
            className="h-full max-w-none object-contain"
            style={{ width: containerWidth || '100%' }}
          />
        </div>

        {/* Slider Handle */}
        <div
          className="absolute top-0 bottom-0 w-1 bg-primary cursor-ew-resize z-10 flex items-center justify-center hover:bg-primary/90 transition-colors"
          style={{ left: `${sliderPosition}%`, transform: 'translateX(-50%)' }}
          onMouseDown={handleMouseDown}
        >
          <div className="w-8 h-16 bg-primary rounded-lg flex items-center justify-center shadow-lg">
             <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-primary-foreground">
                <path d="m9 18 6-6-6-6"/>
             </svg>
          </div>
        </div>
      </div>
    </div>
  )
}
