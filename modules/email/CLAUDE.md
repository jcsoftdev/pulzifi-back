# Email Module

## Responsibility

Email sending via Resend API, status tracking, and template rendering for transactional emails.

## Entities

- **Email** — ID, To, Subject, Body, Status (pending/sent/failed/bounced), CreatedAt, SentAt

## Repository Interfaces

- `EmailRepository` — Save, GetByID, GetByTo, Update

## Routes

None — internal service only, called by other modules via gRPC.

## Infrastructure

- **Resend provider**: `infrastructure/providers/resend_provider.go`
- **Templates**: Registration approval, rejection, team invitation, password reset, monitoring alerts
- **gRPC server**: Exposes email sending to other modules

## Dependencies

- Resend API (`RESEND_API_KEY` env var)
- Config: `EMAIL_FROM_ADDRESS`, `EMAIL_FROM_NAME`

## Constraints

- Disabled if `RESEND_API_KEY` is not set
- Emails sent asynchronously
- Retry logic for transient failures
