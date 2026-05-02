import type { Integration } from '../domain/types'
import { ConnectButton } from './connect-button'

interface ProviderCardProps {
  providerKey: string
  label: string
  description: string
  enabled: boolean
  integration?: Integration
}

function ProviderIcon({ providerKey }: { providerKey: string }) {
  switch (providerKey) {
    case 'slack':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
        </svg>
      )
    case 'email':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"
          />
        </svg>
      )
    case 'teams':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path d="M20.625 7.125a2.625 2.625 0 1 0 0-5.25 2.625 2.625 0 0 0 0 5.25zm-5.08 2.277a4.5 4.5 0 0 0-6.09 0 4.5 4.5 0 0 0 0 6.09 4.5 4.5 0 0 0 6.09 0 4.5 4.5 0 0 0 0-6.09zm-3.045 7.89a3.375 3.375 0 1 1 0-6.75 3.375 3.375 0 0 1 0 6.75zm7.687-6.792H18c-.207 0-.367.16-.367.367v4.633c0 1.013-.82 1.833-1.833 1.833s-1.833-.82-1.833-1.833V11.1a5.625 5.625 0 0 1-1.717-1.282 4.883 4.883 0 0 1 1.05-7.143A5.576 5.576 0 0 1 15.5 2h4.75C21.216 2 22 2.784 22 3.75v3.75c0 .414-.336.75-.75.75h-.063V10.5z" />
        </svg>
      )
    case 'discord':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057c.002.022.013.04.031.053a19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
        </svg>
      )
    default:
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244"
          />
        </svg>
      )
  }
}

export function ProviderCard({
  providerKey,
  label,
  description,
  enabled,
  integration,
}: Readonly<ProviderCardProps>) {
  return (
    <div
      className={`relative flex flex-col gap-4 rounded-2xl border border-border bg-card p-5 transition-shadow ${
        enabled ? 'hover:shadow-md' : 'opacity-60'
      }`}
    >
      {/* Coming soon badge */}
      {!enabled && (
        <span className="absolute top-3 right-3 text-[10px] font-medium bg-muted text-muted-foreground px-2 py-0.5 rounded-full">
          Coming soon
        </span>
      )}

      {/* Header: icon + status */}
      <div className="flex items-start gap-3">
        <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center text-muted-foreground flex-shrink-0">
          <ProviderIcon providerKey={providerKey} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">{label}</span>
            {integration && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-green-500/15 text-green-600 dark:text-green-400">
                Connected
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{description}</p>
        </div>
      </div>

      {/* Action */}
      {enabled && (
        <ConnectButton
          providerKey={providerKey}
          integration={integration}
          label={label}
        />
      )}
    </div>
  )
}
