# Admin Module

Manages user registration requests and admin approval workflow.

## Domain Entities

- `RegistrationRequest` — pending registration with user info, org details, and approval status

## Use Cases

- `list_pending_users` — list pending registrations (with pagination)
- `approve_user` — approve a registration and create organization
- `reject_user` — reject a registration request

## HTTP Routes (`/admin/*`, requires SUPER_ADMIN role)

- GET `/admin/users/pending`
- PUT `/admin/users/{id}/approve`
- PUT `/admin/users/{id}/reject`

## Infrastructure

- PostgreSQL: `registration_requests` table (public schema)
- Email: sends approval/rejection notifications via templates
- Cross-module: integrates with Organization, Auth, Email modules
