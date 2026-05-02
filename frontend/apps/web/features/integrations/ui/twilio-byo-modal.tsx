'use client'

import { IntegrationsApi } from '@workspace/services'
import type { Integration } from '../domain/types'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@workspace/ui/components/atoms/dialog'

interface TwilioBYOModalProps {
  open: boolean
  onClose: () => void
  onConnected: (integration: Integration) => void
}

export function TwilioBYOModal({ open, onClose, onConnected }: Readonly<TwilioBYOModalProps>) {
  const [accountSid, setAccountSid] = useState('')
  const [authToken, setAuthToken] = useState('')
  const [fromNumber, setFromNumber] = useState('')
  const [saving, setSaving] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!accountSid.trim() || !authToken.trim() || !fromNumber.trim()) {
      notification.error({ title: 'All fields are required' })
      return
    }

    setSaving(true)
    try {
      const integration = await IntegrationsApi.connectBYO('twilio', {
        account_sid: accountSid.trim(),
        auth_token: authToken.trim(),
        from_number: fromNumber.trim(),
      })
      notification.success({ title: 'Twilio connected', description: 'BYO credentials saved.' })
      onConnected(integration)
      onClose()
    } catch (err) {
      notification.error({
        title: 'Failed to connect Twilio',
        description: err instanceof Error ? err.message : 'Check your credentials and try again.',
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose() }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect your Twilio account</DialogTitle>
          <DialogDescription>
            Enter your Twilio credentials. These are stored securely and used only to send SMS
            notifications on your behalf.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div>
            <label
              htmlFor="twilio-account-sid"
              className="text-xs font-medium text-muted-foreground block mb-1.5"
            >
              Account SID
            </label>
            <input
              id="twilio-account-sid"
              type="text"
              value={accountSid}
              onChange={(e) => setAccountSid(e.target.value)}
              placeholder="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
              className="w-full h-9 rounded-lg border border-border bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
              autoComplete="off"
            />
          </div>

          <div>
            <label
              htmlFor="twilio-auth-token"
              className="text-xs font-medium text-muted-foreground block mb-1.5"
            >
              Auth Token
            </label>
            <input
              id="twilio-auth-token"
              type="password"
              value={authToken}
              onChange={(e) => setAuthToken(e.target.value)}
              placeholder="Your Twilio auth token"
              className="w-full h-9 rounded-lg border border-border bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
              autoComplete="new-password"
            />
          </div>

          <div>
            <label
              htmlFor="twilio-from-number"
              className="text-xs font-medium text-muted-foreground block mb-1.5"
            >
              From number (E.164)
            </label>
            <input
              id="twilio-from-number"
              type="text"
              value={fromNumber}
              onChange={(e) => setFromNumber(e.target.value)}
              placeholder="+15551234567"
              className="w-full h-9 rounded-lg border border-border bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Must be a Twilio-verified number in E.164 format (e.g. +15551234567).
            </p>
          </div>

          <DialogFooter className="pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={saving}
              className="text-xs font-medium bg-muted hover:bg-muted/70 text-foreground px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
            >
              {saving ? 'Connecting...' : 'Connect Twilio'}
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
