'use client'

// ─── Types ───────────────────────────────────────────────────────────────────
export type {
  AnimationMode,
  Duration,
  NotixButton,
  NotixOptions,
  NotixPosition,
  NotixPromiseOptions,
  NotixStyles,
  ToastData,
  ToastId,
  ToastLifecycle,
  ToastRenderProps,
  ToastState,
  TriggerRect,
} from './domain/entities/types'
// ─── Imperative API ──────────────────────────────────────────────────────────
export { notix } from './presentation/api'
export type { NotixAnchorClassNames, NotixAnchorProps } from './presentation/components/anchor'
export { NotixAnchor } from './presentation/components/anchor'
export type { NotixTriggerProps } from './presentation/components/toast-trigger'
export { NotixTrigger } from './presentation/components/toast-trigger'
export type { ToasterProps } from './presentation/components/toaster'
// ─── Components ──────────────────────────────────────────────────────────────
export { Toaster } from './presentation/components/toaster'
// ─── Hooks ───────────────────────────────────────────────────────────────────
export { useToast, useToastStore } from './presentation/hooks/use-toast'
export { useTriggerRect } from './presentation/hooks/use-trigger-rect'
