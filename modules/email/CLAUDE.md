# Email Module

Email service with pluggable providers and HTML templates.

## Domain Entities

- `Email` — email record with status (pending, sent, failed, bounced)

## Use Cases

- `send_email` — send email via provider
- `get_email_status` — check email delivery status

## HTTP Routes (`/emails/*`)

- POST `/emails/send`

## Domain Services

- `EmailService` — business logic for email operations
- `EmailProvider` — Resend HTTP client for delivery

## Infrastructure

- Resend provider (production) or in-memory (dev)
- Templates: registration, approval, rejection, password reset
- Fire-and-forget sending from handlers
