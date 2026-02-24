'use client'

import { motion } from 'motion/react'
import {
  type CSSProperties,
  type ComponentPropsWithoutRef,
  type ComponentType,
  type ReactNode,
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import {
  BLUR_RATIO,
  DEFAULT_ROUNDNESS,
  SPRING,
  TOAST_WIDTH as BODY_WIDTH,
} from '../../constants'
import type { ToastState } from '../../domain/entities/types'
import { GooeyDefs } from './gooey-defs'

/* ─── Props ─── */

export interface NotixAnchorClassNames {
  /** Wrapper div */
  root?: string
  /** Expandable content panel below the pill */
  content?: string
  /** Title text inside the content panel */
  title?: string
  /** Description text inside the content panel */
  description?: string
}

interface NotixAnchorOwnProps {
  title?: string
  description?: ReactNode | string
  state?: ToastState
  /**
   * Fill color for the pill and body shapes.
   * Accepts any CSS color including `var(--background)`.
   * Defaults to `var(--notix-fill)` which resolves to `var(--background, #fff)`.
   */
  fill?: string
  /**
   * Border (stroke) color for both shapes.
   * Accepts any CSS color including `var(--border)`.
   * Defaults to `var(--notix-anchor-border-color)` which resolves to `var(--border, #d9d9d9)`.
   */
  borderColor?: string
  /**
   * Border thickness in px. Rendered as a drop-shadow on the gooey canvas so
   * it outlines the entire merged shape, not each rect individually.
   * Defaults to 1.
   */
  borderWidth?: number
  /** Text color on the fill background. Overrides `--notix-on-fill`. */
  onFill?: string
  /** Accent color for the title. Overrides the state-based color. */
  accentColor?: string
  /** Tailwind class names for individual slots. */
  classNames?: NotixAnchorClassNames
}

// biome-ignore lint/suspicious/noExplicitAny: needed for generic component constraint
export type NotixAnchorProps<
  C extends ComponentType<any> = ComponentType<any>,
> = NotixAnchorOwnProps & { as: C } & Omit<
    ComponentPropsWithoutRef<C>,
    keyof NotixAnchorOwnProps | 'as'
  >

/* ─── Component ─── */

// biome-ignore lint/suspicious/noExplicitAny: needed for generic component constraint
export function NotixAnchor<C extends ComponentType<any>>({
  as: Trigger,
  title,
  description,
  state = 'info',
  fill,
  borderColor,
  borderWidth = 1,
  onFill,
  accentColor,
  classNames,
  ...triggerProps
}: NotixAnchorProps<C>) {
  const [open, setOpen] = useState(false)
  const [ready, setReady] = useState(false)
  const [pillSize, setPillSize] = useState({ w: 0, h: 0 })
  const [triggerRadius, setTriggerRadius] = useState(0)
  const [contentHeight, setContentHeight] = useState(0)

  const wrapperRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)

  const uid = useId()
  const filterId = `notix-anchor-gooey${uid.replace(/:/g, '-')}`

  const hasContent = Boolean(title) || Boolean(description)
  const blur = DEFAULT_ROUNDNESS * BLUR_RATIO

  /* ─── Measure trigger: size + border-radius only ─── */
  useLayoutEffect(() => {
    const el = triggerRef.current
    if (!el) return
    setPillSize({ w: el.offsetWidth, h: el.offsetHeight })
    // Temporarily lift the data attribute so the CSS rule
    // `border-radius: transparent !important` doesn't interfere
    el.removeAttribute('data-notix-anchor-header')
    const r = parseFloat(getComputedStyle(el).borderTopLeftRadius)
    el.setAttribute('data-notix-anchor-header', '')
    setTriggerRadius(Number.isFinite(r) ? r : DEFAULT_ROUNDNESS)
  }, [])

  /* ─── Measure content height ─── */
  useLayoutEffect(() => {
    const el = contentRef.current
    if (!el) return
    let rafId = 0
    const measure = () => setContentHeight(el.scrollHeight)
    measure()
    const ro = new ResizeObserver(() => {
      cancelAnimationFrame(rafId)
      rafId = requestAnimationFrame(measure)
    })
    ro.observe(el)
    return () => {
      cancelAnimationFrame(rafId)
      ro.disconnect()
    }
  }, [])

  /* ─── Ready after first paint ─── */
  useEffect(() => {
    const raf = requestAnimationFrame(() => setReady(true))
    return () => cancelAnimationFrame(raf)
  }, [])

  /* ─── Click outside → close ─── */
  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (wrapperRef.current?.contains(e.target as Node)) return
      setOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [open])

  const handleClick = useCallback(() => setOpen((p) => !p), [])
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        setOpen(false)
      }
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        setOpen((p) => !p)
      }
    },
    [],
  )

  /* ─── Derived sizes ─── */
  const pillW = pillSize.w || 40
  const pillH = pillSize.h || 40
  const svgW = Math.max(pillW, BODY_WIDTH)
  const pillX = svgW - pillW

  const pillRectH = open ? pillH + blur * 3 : pillH
  const svgH = Math.max(open ? pillH + contentHeight : pillH, pillH)

  /* ─── Resolved visual tokens ─── */
  // SVG `fill` / `stroke` as SVG *presentation attributes* don't support
  // CSS variables — only CSS `style` properties do. We apply both via `style`
  // so that `var(--background)`, `var(--border)`, etc. resolve correctly.
  //
  // The gooey feComposite clips the outer half of the SVG stroke, so we
  // double the strokeWidth so the visible inner half matches the desired px.
  const resolvedFill = fill ?? 'var(--notix-fill)'
  const resolvedStroke = borderColor ?? 'var(--notix-anchor-border-color)'
  const resolvedStrokeWidth = borderWidth

  /* ─── Motion targets ─── */
  const pillAnimate = useMemo(
    () => ({ x: pillX, width: pillW, height: pillRectH }),
    [pillX, pillW, pillRectH]
  )

  const bodyAnimate = useMemo(
    () => ({ height: open ? contentHeight : 0, opacity: open ? 1 : 0 }),
    [open, contentHeight]
  )

  const pillTransition = useMemo(
    () => (ready ? SPRING : { duration: 0 }),
    [ready]
  )

  const bodyTransition = useMemo(
    () => (open ? SPRING : { ...SPRING, bounce: 0 }),
    [open]
  )

  /* ─── Styles ─── */
  const rootStyle = useMemo(
    () =>
      ({
        ...(onFill ? { '--notix-on-fill': onFill } : {}),
        ...(accentColor ? { '--_c': accentColor } : {}),
      }) as CSSProperties,
    [onFill, accentColor]
  )

  // Border is applied as a drop-shadow ON TOP of the gooey filter output
  // so it traces the outer edge of the merged organic shape — not each rect.
  // drop-shadow(0 0 Npx color) with a tiny blur makes a crisp outline when stacked.
  const canvasStyle = useMemo<CSSProperties>(() => {
    if (resolvedStrokeWidth > 0 && resolvedStroke !== 'none') {
      const blur = `${resolvedStrokeWidth * 0.5}px`
      const shadow = `drop-shadow(0 0 ${blur} ${resolvedStroke})`
      return { filter: `url(#${filterId}) ${shadow} ${shadow}` }
    }
    return { filter: `url(#${filterId})` }
  }, [filterId, resolvedStroke, resolvedStrokeWidth])

  // Fill only — no stroke on individual rects (border comes from canvas drop-shadow)
  const rectStyle = useMemo<CSSProperties>(
    () => ({ fill: resolvedFill }),
    [resolvedFill]
  )

  const viewBox = `0 0 ${svgW} ${svgH}`

  return (
    <div
      ref={wrapperRef}
      data-notix-anchor
      data-state={state}
      data-open={open || undefined}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role="button"
      tabIndex={0}
      aria-expanded={hasContent ? open : undefined}
      aria-label={title ? `${state}: ${title}` : undefined}
      style={rootStyle}
      className={classNames?.root}
    >
      {/* SVG canvas: gooey pill rect + body rect */}
      <div data-notix-canvas style={canvasStyle}>
        <svg data-notix-svg width={svgW} height={svgH} viewBox={viewBox}>
          <title>Notification anchor</title>
          <GooeyDefs filterId={filterId} blur={blur} />

          {/* Pill rect — fill/stroke via CSS style so var() resolves */}
          <motion.rect
            data-notix-pill-rect
            rx={triggerRadius}
            ry={triggerRadius}
            style={rectStyle}
            initial={false}
            animate={pillAnimate}
            transition={pillTransition}
          />

          {/* Body rect — same fill/stroke as pill */}
          {hasContent && (
            <motion.rect
              data-notix-body-rect
              y={pillH}
              x={0}
              width={svgW}
              rx={DEFAULT_ROUNDNESS}
              ry={DEFAULT_ROUNDNESS}
              style={rectStyle}
              initial={false}
              animate={bodyAnimate}
              transition={bodyTransition}
            />
          )}
        </svg>
      </div>

      {/* Trigger — data-notix-anchor-header strips its own bg/border so SVG shows through */}
      {/* biome-ignore lint/suspicious/noExplicitAny: generic spread requires cast */}
      <Trigger
        ref={triggerRef}
        {...({ 'data-notix-anchor-header': '', ...triggerProps } as any)}
      />

      {/* Content panel overlaid on the body rect */}
      <div
        data-notix-anchor-content
        data-visible={open || undefined}
        className={classNames?.content}
      >
        <div ref={contentRef} data-notix-anchor-content-inner>
          {title && (
            <div data-notix-anchor-title className={classNames?.title}>
              {title}
            </div>
          )}
          {description && (
            <div
              data-notix-anchor-description
              className={classNames?.description}
            >
              {description}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
