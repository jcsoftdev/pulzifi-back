'use client'

import { useState } from 'react'
import { INTEGRATION_TABS, WEBHOOK_INTEGRATIONS, type ServiceType } from './domain/types'
import { useSettings } from './application/use-settings'
import { UpgradeBanner } from './ui/upgrade-banner'
import { WebhookIntegrationRow } from './ui/webhook-integration-row'

interface SettingsFeatureProps {
  userEmail?: string
}

export function SettingsFeature({ userEmail }: Readonly<SettingsFeatureProps>) {
  const [activeTab, setActiveTab] = useState<string>('email')
  const { loading, connectIntegration, disconnectIntegration, getIntegrationByType } = useSettings()

  const handleConnect = async (serviceType: ServiceType, url: string) => {
    await connectIntegration(serviceType, url)
  }

  const handleDisconnect = async (id: string) => {
    await disconnectIntegration(id)
  }

  return (
    <div className="px-4 md:px-8 lg:px-24 py-8 max-w-3xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Customize where you will receive alerts and notifications.
        </p>
      </div>

      {/* Upgrade banner */}
      <UpgradeBanner />

      {/* Integrations section */}
      <div className="mt-10">
        <h2 className="text-base font-semibold text-foreground">Integrations</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Customize where you will receive alerts and notifications.
        </p>

        {/* Integration type tabs */}
        <div className="flex items-center gap-4 mt-5 flex-wrap">
          {INTEGRATION_TABS.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => !tab.premium && setActiveTab(tab.id)}
              className={`flex flex-col items-center gap-1.5 w-16 group ${tab.premium ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'}`}
            >
              <div
                className={`w-14 h-14 rounded-xl flex items-center justify-center transition-colors
                  ${activeTab === tab.id ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground group-hover:bg-muted/80'}`}
              >
                <TabIcon id={tab.id} />
                {tab.premium && (
                  <span className="absolute -top-1 -right-1 text-[10px]">🔒</span>
                )}
              </div>
              <span className="text-xs text-muted-foreground truncate w-full text-center">
                {tab.label}
              </span>
            </button>
          ))}
        </div>

        {/* Content area */}
        <div className="mt-6 space-y-3">
          {/* Email channel (always shown, free) */}
          <div className="bg-secondary dark:bg-muted border border-border rounded-xl px-4 py-3">
            <p className="text-xs text-muted-foreground mb-1">Emails</p>
            <div className="flex items-center gap-2 flex-wrap">
              {userEmail && (
                <span className="text-sm text-foreground bg-background px-2.5 py-1 rounded-md">
                  {userEmail}
                </span>
              )}
              <span className="text-sm text-muted-foreground">
                <span className="text-primary font-medium cursor-pointer">+ add more</span>{' '}
                in account settings
              </span>
            </div>
          </div>

          {/* Text/SMS channel (premium placeholder) */}
          <div className="bg-secondary dark:bg-muted border border-border rounded-xl px-4 py-3 opacity-60">
            <p className="text-xs text-muted-foreground mb-1">Text</p>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">
                <span className="text-primary font-medium">+ add more</span>{' '}
                in account settings
              </span>
            </div>
          </div>

          {/* Upgrade CTA */}
          <button
            type="button"
            className="border border-primary text-primary text-sm font-medium px-4 py-2 rounded-full hover:bg-primary/5 transition-colors"
          >
            Upgrade to get more
          </button>

          {/* Webhook-based integrations */}
          {loading ? (
            <div className="space-y-3 pt-2">
              {[1, 2, 3, 4].map((n) => (
                <div key={n} className="border border-border rounded-xl px-4 py-3 h-16 bg-muted/30 animate-pulse" />
              ))}
            </div>
          ) : (
            <div className="space-y-3 pt-2">
              {WEBHOOK_INTEGRATIONS.map((item) => (
                <WebhookIntegrationRow
                  key={item.serviceType}
                  label={item.label}
                  serviceType={item.serviceType}
                  placeholder={item.placeholder}
                  integration={getIntegrationByType(item.serviceType)}
                  onConnect={handleConnect}
                  onDisconnect={handleDisconnect}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function TabIcon({ id }: { id: string }) {
  switch (id) {
    case 'email':
      return (
        <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
        </svg>
      )
    case 'text':
      return (
        <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
        </svg>
      )
    case 'slack':
      return (
        <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
          <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z"/>
        </svg>
      )
    case 'google_sheets':
      return (
        <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
          <path d="M14.727 6.727H14V0H4.91C4.41 0 4 .409 4 .909V23.09c0 .5.41.909.91.909H19.09c.5 0 .91-.409.91-.909V6.727h-5.273zM9.273 17.454H7.636v-1.636h1.637v1.636zm0-2.727H7.636V13.09h1.637v1.636zm0-2.727H7.636v-1.637h1.637V12zm2.727 5.454H10.364v-1.636H12v1.636zm0-2.727H10.364V13.09H12v1.636zm0-2.727H10.364v-1.637H12V12zm2.727 5.454H13.09v-1.636h1.637v1.636zm0-2.727H13.09V13.09h1.637v1.636zm0-2.727H13.09v-1.637h1.637V12zm0-3.273H7.636V6.546h7.091v1.636zM20 5.818l-4.546-4.545v4.545H20z"/>
        </svg>
      )
    case 'teams':
      return (
        <svg xmlns="http://www.w3.org/2000/svg" className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
          <path d="M20.625 7.125a2.625 2.625 0 1 0 0-5.25 2.625 2.625 0 0 0 0 5.25zm-5.08 2.277a4.5 4.5 0 0 0-6.09 0 4.5 4.5 0 0 0 0 6.09 4.5 4.5 0 0 0 6.09 0 4.5 4.5 0 0 0 0-6.09zm-3.045 7.89a3.375 3.375 0 1 1 0-6.75 3.375 3.375 0 0 1 0 6.75zm7.687-6.792H18c-.207 0-.367.16-.367.367v4.633c0 1.013-.82 1.833-1.833 1.833s-1.833-.82-1.833-1.833V11.1a5.625 5.625 0 0 1-1.717-1.282 4.883 4.883 0 0 1 1.05-7.143A5.576 5.576 0 0 1 15.5 2h4.75C21.216 2 22 2.784 22 3.75v3.75c0 .414-.336.75-.75.75h-.063V10.5z"/>
        </svg>
      )
    default:
      return null
  }
}
