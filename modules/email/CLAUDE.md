# Email Module

Email service with pluggable providers and HTML templates.

## Domain Entities

- `Email` — email record with status (pending, sent, failed, bounced)

## Use Cases (application/ directories)

- `send_email` — send email via provider
- `get_email_status` — check email delivery status

## HTTP Routes (`/emails/*`)

- POST `/emails/send` — send email

## Domain Services

- `EmailService` — business logic for email operations
- `EmailProvider` — interface for email delivery providers

## Infrastructure

- Resend provider: `infrastructure/providers/resend_provider.go` (production)
- In-memory repository: `infrastructure/persistence/memory_repository.go` (no database)
- Templates: `infrastructure/templates/templates.go` (registration, approval, rejection, password reset, invitation)
- Fire-and-forget sending from other module handlers

## Notes

- Uses in-memory persistence (not PostgreSQL) for email records
- The `EmailProvider` interface is shared across modules (admin, team, auth use it for sending emails)

## Cross-Module Dependencies

This module's `EmailProvider` interface and template functions are imported directly by:
- `admin` — approval/rejection emails
- `auth` — password reset, registration emails
- `team` — invitation emails

This creates a dependency from those modules' infrastructure layers into this module's infrastructure layer, violating hexagonal boundaries.

## Architecture Improvements

### Shared Email Interface
Move the `EmailProvider` interface to `shared/` so other modules depend on the shared interface rather than importing from `modules/email/infrastructure/`. Wire the Resend implementation via dependency injection.

### Persistent Email Records
The in-memory repository loses all email records on restart. For production:
- Add a PostgreSQL-backed repository for email audit trail
- Track delivery status callbacks from Resend for bounce/complaint handling

### Async Email Sending
Email sends currently happen synchronously in request handlers. Consider:
- Publishing email requests to the EventBus and processing in the worker
- This prevents slow email API calls from blocking HTTP responses
