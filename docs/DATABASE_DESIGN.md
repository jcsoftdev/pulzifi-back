# Pulzifi Database Design

## Product Overview
**Pulzifi** is an AI-powered website monitoring platform that turns web changes into competitive advantage. It tracks changes on any webpage and uses AI to explain why they matter, delivering actionable insights for both business teams and personal users.

**Core Value Propositions:**
- Detect visual, text, HTML/CSS, SEO, and pricing changes instantly
- AI-powered summaries with Marketing, Product, SEO, and Business lenses
- Smart notifications via email, phone, Slack, or spreadsheet
- **Dual use cases**: Business intelligence + Personal monitoring

**Target Users:**
- **Business**: Marketers, Product Managers, SEO specialists, Founders, Agencies
- **Personal**: Job hunters, scholarship seekers, ticket trackers, price watchers

---

## Multi-Tenant Strategy

**Approach**: Schema per Tenant (Organization) - PostgreSQL

### Core Concept: 1 Tenant = 1 Subdomain = 1 Schema

**Each tenant (organization) is:**
- Identified by a **unique subdomain** (e.g., `jcsoftdev-inc.pulzifi.com`)
- Mapped to a **dedicated PostgreSQL schema** (e.g., `jcsoftdev_inc`)
- Completely **isolated** from other tenants
- Has the **exact same database structure** (tables, indexes, constraints)

**Key Points:**
- **Tenant = Organization = Subdomain = Schema**
- Each schema contains **identical table structures** but **separate data**
- A **user** can belong to **multiple organizations** (multiple tenants/schemas)
- User switches tenant context by accessing different subdomains
- `public` schema contains **only** shared data (users, organizations, memberships, auth)
- All business data lives in **tenant-specific schemas**

### Visual Representation

```
PostgreSQL Database: pulzifi
│
├── public (shared schema)
│   ├── users (all users across all organizations)
│   ├── organizations (all organizations)
│   ├── organization_members (user-org mappings)
│   ├── refresh_tokens
│   └── password_resets
│
├── jcsoftdev_inc (tenant schema)
│   ├── workspaces
│   ├── pages
│   ├── page_tags
│   ├── monitoring_configs
│   ├── checks
│   ├── alerts
│   ├── notification_preferences
│   ├── email_logs
│   ├── insights
│   ├── insight_rules
│   ├── reports
│   ├── integrations
│   ├── usage_tracking
│   └── usage_logs
│
├── toyota_corp (tenant schema - SAME STRUCTURE)
│   ├── workspaces
│   ├── pages
│   ├── page_tags
│   ├── monitoring_configs
│   ├── checks
│   ├── alerts
│   ├── notification_preferences
│   ├── email_logs
│   ├── insights
│   ├── insight_rules
│   ├── reports
│   ├── integrations
│   ├── usage_tracking
│   └── usage_logs
│
└── nissan_ltd (tenant schema - SAME STRUCTURE)
    ├── workspaces
    ├── pages
    ├── page_tags
    ├── monitoring_configs
    ├── checks
    ├── alerts
    ├── notification_preferences
    ├── email_logs
    ├── insights
    ├── insight_rules
    ├── reports
    ├── integrations
    ├── usage_tracking
    └── usage_logs
```

### Example User Scenario

**User**: dania@example.com
- Can access **3 different tenants** (3 different organizations)
- Stored **once** in `public.users`
- Memberships stored in `public.organization_members`

```
dania@example.com accesses:

1. jcsoftdev-inc.pulzifi.com
   → Tenant: jcsoftdev_inc
   → Role: ADMIN
   → Sees data ONLY from schema: jcsoftdev_inc
   
2. toyota-corp.pulzifi.com
   → Tenant: toyota_corp
   → Role: MEMBER
   → Sees data ONLY from schema: toyota_corp
   
3. nissan-ltd.pulzifi.com
   → Tenant: nissan_ltd
   → Role: MEMBER
   → Sees data ONLY from schema: nissan_ltd
```

### Request Flow with Subdomain → Schema Mapping

1. **User Request**: `GET https://jcsoftdev-inc.pulzifi.com/api/v1/workspaces`
2. **Gateway extracts subdomain**: `jcsoftdev-inc`
3. **Gateway normalizes to schema name**: `jcsoftdev_inc`
4. **Gateway validates**: User is member of organization with this schema
5. **Gateway passes tenant**: `jcsoftdev_inc` in gRPC metadata
6. **Module sets search_path**: `SET search_path TO jcsoftdev_inc`
7. **Query executes**: `SELECT * FROM workspaces` (in jcsoftdev_inc schema)
8. **Result**: Only jcsoftdev's workspaces, completely isolated from Toyota/Nissan

---

## Schema: `public` (Shared across tenants)

### Table: `organizations`
Stores organization/tenant information.

```sql
CREATE TABLE public.organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL, -- e.g., 'jcsoftdev-inc'
    schema_name VARCHAR(100) UNIQUE NOT NULL, -- e.g., 'jcsoftdev_inc'
    owner_user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_organizations_subdomain ON public.organizations(subdomain);
CREATE INDEX idx_organizations_schema_name ON public.organizations(schema_name);
```

### Table: `users`
Global user accounts (can belong to multiple organizations).

```sql
CREATE TABLE public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    email_notifications_enabled BOOLEAN DEFAULT TRUE,
    notification_frequency VARCHAR(50) DEFAULT 'immediate', -- 'immediate', 'daily_digest', 'weekly_digest'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON public.users(email);
CREATE INDEX idx_users_email_verified ON public.users(email_verified);
CREATE INDEX idx_users_email_notifications_enabled ON public.users(email_notifications_enabled);
```

### Table: `organization_members`
Maps users to organizations with roles. This is the key table for multi-tenant authorization.

**Purpose**: 
- Determine which users belong to which organizations
- Store user roles per organization (ADMIN, MEMBER)
- Validate user access when switching between tenants

```sql
CREATE TABLE public.organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'MEMBER', -- ADMIN, MEMBER
    invited_by UUID REFERENCES public.users(id),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_organization_members_org_id ON public.organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON public.organization_members(user_id);
CREATE INDEX idx_organization_members_user_org ON public.organization_members(user_id, organization_id);
```

**Usage Example:**
```sql
-- Check if user has access to organization
SELECT role FROM public.organization_members
WHERE user_id = $1 AND organization_id = $2;

-- List all organizations for a user
SELECT o.*, om.role 
FROM public.organizations o
JOIN public.organization_members om ON o.id = om.organization_id
WHERE om.user_id = $1;
```

### Table: `refresh_tokens`
Stores JWT refresh tokens.

```sql
CREATE TABLE public.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP NULL
);

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON public.refresh_tokens(token);
```

### Table: `password_resets`
Temporary tokens for password reset flow.

```sql
CREATE TABLE public.password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP NULL
);

CREATE INDEX idx_password_resets_token ON public.password_resets(token);
```

---

## Tenant Schema: `<tenant_name>` (Per organization)

**CRITICAL**: Every tenant schema has the **EXACT SAME STRUCTURE**.

When a new organization is created:
1. A new schema is created (e.g., `jcsoftdev_inc`, `toyota_corp`)
2. All tables below are created in that schema
3. Structure is identical across all tenant schemas
4. Data is completely isolated per tenant

**Important Notes:**
- Users are stored in `public.users` (shared across all tenants)
- All references to users in tenant schemas use UUID only
- No FK constraints to public schema (maintains schema independence)
- Each tenant schema is a complete, isolated copy of the structure

---

### Tenant Schema Structure (Applied to ALL tenants)

The following table structure is created in **every tenant schema**:

### Table: `workspaces`
Workspaces within an organization.

**Schema**: Created in each tenant schema (e.g., `jcsoftdev_inc.workspaces`, `toyota_corp.workspaces`)

```sql
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'Personal', 'Competitor'
    description TEXT,
    created_by UUID NOT NULL, -- references public.users(id) but no FK constraint
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_workspaces_type ON workspaces(type);
CREATE INDEX idx_workspaces_created_by ON workspaces(created_by);
CREATE INDEX idx_workspaces_deleted_at ON workspaces(deleted_at);
CREATE INDEX idx_workspaces_active ON workspaces(deleted_at) WHERE deleted_at IS NULL;
```

**Example Data Isolation:**
```sql
-- In jcsoftdev_inc schema
jcsoftdev_inc.workspaces:
  - id: uuid-1, name: "jcsoftdev Competitors"
  - id: uuid-2, name: "Product Pages"

-- In toyota_corp schema (completely separate)
toyota_corp.workspaces:
  - id: uuid-3, name: "Toyota Market Analysis"
  - id: uuid-4, name: "Pricing Monitoring"
```

### Table: `pages`
Monitored pages/URLs within workspaces.

```sql
CREATE TABLE pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    last_checked_at TIMESTAMP NULL,
    last_change_detected_at TIMESTAMP NULL,
    check_count INT DEFAULT 0,
    created_by UUID NOT NULL, -- references public.users(id) but no FK
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_pages_workspace_id ON pages(workspace_id);
CREATE INDEX idx_pages_url ON pages(url);
CREATE INDEX idx_pages_last_checked_at ON pages(last_checked_at);
CREATE INDEX idx_pages_active ON pages(deleted_at) WHERE deleted_at IS NULL;
```

### Table: `page_tags`
Tags assigned to pages.

```sql
CREATE TABLE page_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL, -- e.g., 'Hero section', 'Follow up', 'Promotion'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(page_id, tag)
);

CREATE INDEX idx_page_tags_page_id ON page_tags(page_id);
CREATE INDEX idx_page_tags_tag ON page_tags(tag);
```

### Table: `monitoring_configs`
Monitoring configuration per page.

```sql
CREATE TABLE monitoring_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID UNIQUE NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    check_frequency VARCHAR(50) NOT NULL DEFAULT 'Every day', -- 'Every day', 'Every 3 hours', etc.
    schedule_type VARCHAR(50) DEFAULT 'all_time', -- 'all_time', 'work_days', 'work_days_hours', 'weekdays'
    timezone VARCHAR(100) DEFAULT 'America/Boise',
    block_ads_cookies BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitoring_configs_page_id ON monitoring_configs(page_id);
```

### Table: `checks`
Individual check/scan records.

```sql
CREATE TABLE checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- 'completed', 'failed', 'pending'
    screenshot_url TEXT,
    html_snapshot_url TEXT,
    change_detected BOOLEAN DEFAULT FALSE,
    change_type VARCHAR(50), -- 'visual', 'text', 'html', 'none'
    error_message TEXT,
    duration_ms INT,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_checks_page_id ON checks(page_id);
CREATE INDEX idx_checks_checked_at ON checks(checked_at);
CREATE INDEX idx_checks_change_detected ON checks(change_detected);
```

### Table: `alerts`
Alerts generated from detected changes.

```sql
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    check_id UUID NOT NULL REFERENCES checks(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'Visual', 'Text', 'HTML'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_workspace_id ON alerts(workspace_id);
CREATE INDEX idx_alerts_page_id ON alerts(page_id);
CREATE INDEX idx_alerts_type ON alerts(type);
CREATE INDEX idx_alerts_read_at ON alerts(read_at);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);
```

### Table: `notification_preferences`
Per-workspace and per-page email notification preferences.

```sql
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, -- references public.users(id), no FK constraint
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT TRUE,
    change_types VARCHAR(100)[], -- ['visual', 'text', 'html'] or NULL for all
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Either workspace_id OR page_id must be set (not both)
    CONSTRAINT check_workspace_or_page CHECK (
        (workspace_id IS NOT NULL AND page_id IS NULL) OR
        (workspace_id IS NULL AND page_id IS NOT NULL)
    ),
    
    -- Unique per user + workspace OR user + page
    UNIQUE(user_id, workspace_id),
    UNIQUE(user_id, page_id)
);

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_workspace_id ON notification_preferences(workspace_id);
CREATE INDEX idx_notification_preferences_page_id ON notification_preferences(page_id);
```

**Purpose:**
- Control email notifications at workspace level (e.g., disable all emails for "Competitor X" workspace)
- Control email notifications at page level (e.g., only notify on visual changes for specific page)
- Page-level preferences override workspace-level preferences
- If no preference exists, default is enabled with all change types

**Usage Examples:**
```sql
-- Disable email notifications for entire workspace
INSERT INTO notification_preferences (user_id, workspace_id, email_enabled)
VALUES ('user-uuid', 'workspace-uuid', FALSE);

-- Only notify on visual changes for specific page
INSERT INTO notification_preferences (user_id, page_id, email_enabled, change_types)
VALUES ('user-uuid', 'page-uuid', TRUE, ARRAY['visual']);

-- Get user's page notification preference
SELECT * FROM notification_preferences
WHERE user_id = 'user-uuid' AND page_id = 'page-uuid';
```

### Table: `email_logs`
Track all sent emails for debugging and analytics.

```sql
CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES alerts(id) ON DELETE SET NULL,
    recipient_user_id UUID NOT NULL, -- references public.users(id), no FK
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL, -- 'pending', 'sent', 'failed', 'bounced'
    provider VARCHAR(50), -- 'sendgrid', 'aws_ses'
    provider_message_id VARCHAR(255),
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_logs_alert_id ON email_logs(alert_id);
CREATE INDEX idx_email_logs_recipient_user_id ON email_logs(recipient_user_id);
CREATE INDEX idx_email_logs_recipient_email ON email_logs(recipient_email);
CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_sent_at ON email_logs(sent_at);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);
```

**Purpose:**
- Track all email delivery attempts
- Monitor email delivery success rate
- Debug failed email sends
- Analyze email delivery performance
- Track bounced emails (invalid addresses)

**Usage Examples:**
```sql
-- Get failed emails in last 24 hours
SELECT * FROM email_logs
WHERE status = 'failed' AND created_at >= NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

-- Email delivery statistics for last 7 days
SELECT 
    DATE(sent_at) as date,
    status,
    COUNT(*) as count
FROM email_logs
WHERE sent_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(sent_at), status
ORDER BY date DESC;

-- Top recipients by email volume
SELECT 
    recipient_email,
    COUNT(*) as email_count
FROM email_logs
WHERE sent_at >= NOW() - INTERVAL '30 days'
GROUP BY recipient_email
ORDER BY email_count DESC
LIMIT 10;
```

### Table: `insights`
AI-generated insights from page changes.

```sql
CREATE TABLE insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    check_id UUID NOT NULL REFERENCES checks(id) ON DELETE CASCADE,
    insight_type VARCHAR(100) NOT NULL, -- 'Visual Pulse', 'Marketing Lens', 'Market Analysis'
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB, -- Additional structured data
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_insights_page_id ON insights(page_id);
CREATE INDEX idx_insights_check_id ON insights(check_id);
CREATE INDEX idx_insights_insight_type ON insights(insight_type);
CREATE INDEX idx_insights_created_at ON insights(created_at);
```

### Table: `insight_rules`
Custom rules for triggering specific insights.

```sql
CREATE TABLE insight_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    rule_type VARCHAR(100) NOT NULL, -- 'Marketing Lens', 'Market Analysis', 'Promotion opportunity', etc.
    enabled BOOLEAN DEFAULT TRUE,
    trigger_condition JSONB, -- Conditions for triggering
    created_by UUID NOT NULL, -- references public.users(id) but no FK
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_insight_rules_page_id ON insight_rules(page_id);
CREATE INDEX idx_insight_rules_rule_type ON insight_rules(rule_type);
CREATE INDEX idx_insight_rules_enabled ON insight_rules(enabled) WHERE enabled = TRUE;
```

### Table: `reports`
Generated reports for pages.

```sql
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    report_date DATE NOT NULL,
    content JSONB NOT NULL, -- Structured report data
    pdf_url TEXT,
    created_by UUID NOT NULL, -- references public.users(id) but no FK
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_page_id ON reports(page_id);
CREATE INDEX idx_reports_report_date ON reports(report_date);
CREATE INDEX idx_reports_created_at ON reports(created_at);
```

### Table: `integrations`
External service integrations per organization.

```sql
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_type VARCHAR(50) NOT NULL, -- 'slack', 'teams', 'telegram', 'spreadsheet', 'phone'
    config JSONB NOT NULL, -- Service-specific configuration (webhook URLs, tokens, etc.)
    enabled BOOLEAN DEFAULT TRUE,
    created_by UUID NOT NULL, -- references public.users(id) but no FK
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integrations_service_type ON integrations(service_type);
CREATE INDEX idx_integrations_enabled ON integrations(enabled) WHERE enabled = TRUE;
```

### Table: `usage_tracking`
Track check usage for billing/quotas.

```sql
CREATE TABLE usage_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    checks_allowed INT NOT NULL,
    checks_used INT DEFAULT 0,
    last_refill_at TIMESTAMP,
    next_refill_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_tracking_period ON usage_tracking(period_start, period_end);
```

### Table: `usage_logs`
Detailed log of each check for auditing.

```sql
CREATE TABLE usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    check_id UUID NOT NULL REFERENCES checks(id) ON DELETE CASCADE,
    checks_consumed INT DEFAULT 1,
    logged_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_logs_page_id ON usage_logs(page_id);
CREATE INDEX idx_usage_logs_logged_at ON usage_logs(logged_at);
```

---

## Multi-Tenant Authorization Flow

### Scenario: User Accesses Workspace Page

1. **Frontend Request**:
   ```
   GET https://jcsoftdev-inc.pulzifi.com/api/v1/workspaces
   Authorization: Bearer <JWT>
   ```

2. **Gateway Extracts Tenant**:
   ```go
   subdomain := "jcsoftdev-inc"  // from request host
   tenant := "jcsoftdev_inc"      // normalized
   ```

3. **Gateway Validates JWT**:
   ```go
   userID := extractUserIDFromJWT(token)
   ```

4. **Gateway Validates User Access** (gRPC call to Organization module):
   ```sql
   -- Query in public schema
   SELECT role FROM public.organization_members om
   JOIN public.organizations o ON om.organization_id = o.id
   WHERE om.user_id = $userID 
     AND o.schema_name = $tenant;
   
   -- If no result: 403 Forbidden
   -- If result exists: user has access with returned role
   ```

5. **Gateway Passes Tenant in gRPC Metadata**:
   ```go
   ctx = metadata.AppendToOutgoingContext(ctx, "tenant", tenant)
   ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
   ```

6. **Workspace Module Extracts from Metadata**:
   ```go
   md, _ := metadata.FromIncomingContext(ctx)
   tenant := md.Get("tenant")[0]
   userID := md.Get("user_id")[0]
   ```

7. **Workspace Repository Sets Search Path**:
   ```go
   db.Exec("SET search_path TO " + tenant)
   ```

8. **Query Workspaces**:
   ```sql
   -- Now querying in tenant schema
   SELECT * FROM workspaces WHERE deleted_at IS NULL;
   ```

### Scenario: User Switches Organization

1. **Frontend**: User selects "Toyota Corp" from organization dropdown
2. **Frontend**: Redirects to `https://toyota-corp.pulzifi.com`
3. **Gateway**: Extracts new tenant: `toyota_corp`
4. **Gateway**: Validates user is member of Toyota Corp
5. **All subsequent requests**: Use `toyota_corp` schema

---

## Relationships Summary

### Public Schema (User-Organization Relationship)
- `users` (1) ←→ (N) `organization_members` ←→ (N) `organizations` (1)
  - **Meaning**: A user can belong to many organizations, an organization can have many users
- `users` (1) ← (N) `refresh_tokens`
- `users` (1) ← (N) `password_resets`

### Tenant Schema (Business Data)
- `workspaces` (1) ← (N) `pages`
- `pages` (1) ← (N) `page_tags`
- `pages` (1) ← (1) `monitoring_configs`
- `pages` (1) ← (N) `checks`
- `pages` (1) ← (N) `insight_rules`
- `pages` (1) ← (N) `reports`
- `checks` (1) ← (N) `alerts`
- `checks` (1) ← (N) `insights`
- `checks` (1) ← (N) `usage_logs`
- `workspaces` (1) ← (N) `alerts` (for quick filtering)

**Note**: All `created_by` fields in tenant schemas reference users but without FK constraints to maintain schema independence.

---

## Migration Strategy

### Overview

The migration strategy ensures that:
1. **Public schema** is created once (shared across all tenants)
2. **Tenant schema template** is defined as a function
3. **Every new organization** automatically gets a schema with identical structure
4. **All tenant schemas** have the exact same tables, indexes, and constraints

### Step-by-Step Migration Flow

```
1. Run public schema migrations
   ↓
2. Create tenant schema template function
   ↓
3. Create trigger on organization insert
   ↓
4. When organization created → trigger fires → new schema created with full structure
```

### Initial Setup

#### Phase 1: Public Schema Migrations

**Location**: `shared/database/migrations/public/`

```
001_create_users.up.sql
002_create_organizations.up.sql
003_create_organization_members.up.sql
004_create_refresh_tokens.up.sql
005_create_password_resets.up.sql
006_create_tenant_schema_function.up.sql  ← Creates template
007_create_tenant_trigger.up.sql          ← Auto-creates on org insert
```

#### Phase 2: Tenant Schema Template Function

**File**: `006_create_tenant_schema_function.up.sql`

This function contains the **complete structure** that will be replicated for each tenant:

```sql
CREATE OR REPLACE FUNCTION create_tenant_schema(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Create schema
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
    
    RAISE NOTICE 'Creating tenant schema: %', schema_name;
    
    -- ========================================
    -- TABLE 1: workspaces
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.workspaces (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            type VARCHAR(50) NOT NULL,
            description TEXT,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMP NULL
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_workspaces_type ON %I.workspaces(type)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_created_by ON %I.workspaces(created_by)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_deleted_at ON %I.workspaces(deleted_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_active ON %I.workspaces(deleted_at) WHERE deleted_at IS NULL', schema_name);
    
    -- ========================================
    -- TABLE 2: pages
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.pages (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            workspace_id UUID NOT NULL,
            name VARCHAR(255) NOT NULL,
            url TEXT NOT NULL,
            thumbnail_url TEXT,
            last_checked_at TIMESTAMP NULL,
            last_change_detected_at TIMESTAMP NULL,
            check_count INT DEFAULT 0,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMP NULL,
            CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) 
                REFERENCES %I.workspaces(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_pages_workspace_id ON %I.pages(workspace_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_url ON %I.pages(url)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_last_checked_at ON %I.pages(last_checked_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_active ON %I.pages(deleted_at) WHERE deleted_at IS NULL', schema_name);
    
    -- ========================================
    -- TABLE 3: page_tags
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.page_tags (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            tag VARCHAR(100) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            UNIQUE(page_id, tag)
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_page_tags_page_id ON %I.page_tags(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_page_tags_tag ON %I.page_tags(tag)', schema_name);
    
    -- ========================================
    -- TABLE 4: monitoring_configs
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.monitoring_configs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID UNIQUE NOT NULL,
            check_frequency VARCHAR(50) NOT NULL DEFAULT ''Every day'',
            schedule_type VARCHAR(50) DEFAULT ''all_time'',
            timezone VARCHAR(100) DEFAULT ''America/Boise'',
            block_ads_cookies BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_monitoring_configs_page_id ON %I.monitoring_configs(page_id)', schema_name);
    
    -- ========================================
    -- TABLE 5: checks
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.checks (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            status VARCHAR(50) NOT NULL,
            screenshot_url TEXT,
            html_snapshot_url TEXT,
            change_detected BOOLEAN DEFAULT FALSE,
            change_type VARCHAR(50),
            error_message TEXT,
            duration_ms INT,
            checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_checks_page_id ON %I.checks(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_checks_checked_at ON %I.checks(checked_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_checks_change_detected ON %I.checks(change_detected)', schema_name);
    
    -- ========================================
    -- TABLE 6: alerts
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.alerts (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            workspace_id UUID NOT NULL,
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            type VARCHAR(50) NOT NULL,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            read_at TIMESTAMP NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) 
                REFERENCES %I.workspaces(id) ON DELETE CASCADE,
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_alerts_workspace_id ON %I.alerts(workspace_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_page_id ON %I.alerts(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_type ON %I.alerts(type)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_read_at ON %I.alerts(read_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_created_at ON %I.alerts(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 7: insights
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.insights (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            insight_type VARCHAR(100) NOT NULL,
            title VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            metadata JSONB,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_insights_page_id ON %I.insights(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_check_id ON %I.insights(check_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_insight_type ON %I.insights(insight_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_created_at ON %I.insights(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 8: insight_rules
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.insight_rules (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            rule_type VARCHAR(100) NOT NULL,
            enabled BOOLEAN DEFAULT TRUE,
            trigger_condition JSONB,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_insight_rules_page_id ON %I.insight_rules(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insight_rules_rule_type ON %I.insight_rules(rule_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_insight_rules_enabled ON %I.insight_rules(enabled) WHERE enabled = TRUE', schema_name);
    
    -- ========================================
    -- TABLE 9: reports
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.reports (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            title VARCHAR(255) NOT NULL,
            report_date DATE NOT NULL,
            content JSONB NOT NULL,
            pdf_url TEXT,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_reports_page_id ON %I.reports(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_reports_report_date ON %I.reports(report_date)', schema_name);
    EXECUTE format('CREATE INDEX idx_reports_created_at ON %I.reports(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 10: integrations
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.integrations (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            service_type VARCHAR(50) NOT NULL,
            config JSONB NOT NULL,
            enabled BOOLEAN DEFAULT TRUE,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_integrations_service_type ON %I.integrations(service_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_integrations_enabled ON %I.integrations(enabled) WHERE enabled = TRUE', schema_name);
    
    -- ========================================
    -- TABLE 11: usage_tracking
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.usage_tracking (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            period_start DATE NOT NULL,
            period_end DATE NOT NULL,
            checks_allowed INT NOT NULL,
            checks_used INT DEFAULT 0,
            last_refill_at TIMESTAMP,
            next_refill_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_usage_tracking_period ON %I.usage_tracking(period_start, period_end)', schema_name);
    
    -- ========================================
    -- TABLE 12: usage_logs
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.usage_logs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            checks_consumed INT DEFAULT 1,
            logged_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_usage_logs_page_id ON %I.usage_logs(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_usage_logs_logged_at ON %I.usage_logs(logged_at)', schema_name);
    
    RAISE NOTICE 'Tenant schema % created successfully with all tables', schema_name;
END;
$$ LANGUAGE plpgsql;
```

#### Phase 3: Automatic Trigger
```

#### Phase 3: Automatic Trigger

**File**: `007_create_tenant_trigger.up.sql`

```sql
CREATE OR REPLACE FUNCTION trigger_create_tenant_schema()
RETURNS TRIGGER AS $$
BEGIN
    -- Call function to create tenant schema with full structure
    PERFORM create_tenant_schema(NEW.schema_name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_organization_insert
AFTER INSERT ON public.organizations
FOR EACH ROW
EXECUTE FUNCTION trigger_create_tenant_schema();
```

**How it works:**
1. Application creates new organization: `INSERT INTO public.organizations (name, subdomain, schema_name, ...) VALUES (...)`
2. Trigger fires automatically: `after_organization_insert`
3. Function executes: `create_tenant_schema('jcsoftdev_inc')`
4. Result: New schema `jcsoftdev_inc` created with all 12 tables
5. Organization can immediately start using their isolated schema

---

### Schema Creation Examples

#### Example 1: Creating First Organization

```sql
-- Insert organization (trigger will auto-create schema)
INSERT INTO public.organizations (name, subdomain, schema_name, owner_user_id)
VALUES ('jcsoftdev INC', 'jcsoftdev-inc', 'jcsoftdev_inc', 'user-uuid-123');

-- Result:
-- ✅ public.organizations has new row
-- ✅ Schema 'jcsoftdev_inc' created
-- ✅ 12 tables created in jcsoftdev_inc schema
-- ✅ All indexes created
-- ✅ All foreign keys created
```

#### Example 2: Multiple Organizations (Same Structure, Different Data)

```sql
-- Create 3 organizations
INSERT INTO public.organizations (name, subdomain, schema_name, owner_user_id) VALUES
  ('jcsoftdev INC', 'jcsoftdev-inc', 'jcsoftdev_inc', 'user-1'),
  ('Toyota Corp', 'toyota-corp', 'toyota_corp', 'user-2'),
  ('Nissan Ltd', 'nissan-ltd', 'nissan_ltd', 'user-3');

-- Result:
-- ✅ 3 rows in public.organizations
-- ✅ 3 schemas created: jcsoftdev_inc, toyota_corp, nissan_ltd
-- ✅ Each schema has IDENTICAL structure (12 tables each)
-- ✅ Each schema has completely ISOLATED data
```

#### Verify Schema Structure

```sql
-- List all tenant schemas
SELECT schema_name 
FROM information_schema.schemata 
WHERE schema_name NOT IN ('public', 'information_schema', 'pg_catalog', 'pg_toast')
ORDER BY schema_name;

-- Result:
-- nissan_ltd
-- toyota_corp
-- jcsoftdev_inc

-- Verify all schemas have same table count
SELECT 
    schemaname, 
    COUNT(*) as table_count 
FROM pg_tables 
WHERE schemaname NOT IN ('pg_catalog', 'information_schema', 'public')
GROUP BY schemaname;

-- Result (each should have 12 tables):
-- nissan_ltd     | 12
-- toyota_corp    | 12
-- jcsoftdev_inc | 12
```

---

### Manual Schema Operations (For Development/Testing)

```sql
### Manual Schema Operations (For Development/Testing)

#### Manually Create Schema for Testing

```sql
-- Create test tenant schema
SELECT create_tenant_schema('test_tenant_123');

-- Verify it was created
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'test_tenant_123'
ORDER BY table_name;

-- Expected result (12 tables):
-- alerts
-- checks
-- insight_rules
-- insights
-- integrations
-- monitoring_configs
-- page_tags
-- pages
-- reports
-- usage_logs
-- usage_tracking
-- workspaces
```

#### Manually Create Schema for Existing Organizations (Backfill)

```sql
-- If organizations exist before trigger was created, backfill schemas
SELECT create_tenant_schema(schema_name) 
FROM public.organizations
WHERE schema_name NOT IN (
    SELECT schema_name 
    FROM information_schema.schemata
);
```

#### Clone Schema Structure (For Testing)

```sql
-- Create a test clone of an existing tenant schema
CREATE SCHEMA test_clone;

-- Copy structure from existing tenant (no data)
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN 
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'jcsoftdev_inc'
    LOOP
        EXECUTE format('CREATE TABLE test_clone.%I (LIKE jcsoftdev_inc.%I INCLUDING ALL)', 
            rec.table_name, rec.table_name);
    END LOOP;
END $$;
```

---

### Schema Consistency Validation

#### Check All Schemas Have Same Structure

```sql
-- Get table counts per schema
SELECT 
    schemaname,
    COUNT(DISTINCT tablename) as table_count
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema', 'public')
GROUP BY schemaname
ORDER BY schemaname;

-- All tenants should have exactly 12 tables
```

#### Detect Schema Drift (Missing Tables)

```sql
-- Compare schema structure against reference tenant
WITH reference AS (
    SELECT tablename 
    FROM pg_tables 
    WHERE schemaname = 'jcsoftdev_inc'  -- Reference tenant
),
all_tenants AS (
    SELECT DISTINCT schemaname 
    FROM pg_tables 
    WHERE schemaname NOT IN ('pg_catalog', 'information_schema', 'public')
)
SELECT 
    t.schemaname,
    r.tablename as missing_table
FROM all_tenants t
CROSS JOIN reference r
LEFT JOIN pg_tables pt 
    ON pt.schemaname = t.schemaname 
    AND pt.tablename = r.tablename
WHERE pt.tablename IS NULL;

-- Should return 0 rows if all schemas are consistent
```

#### Validate Schema Isolation

```sql
-- Ensure no foreign keys cross schema boundaries
SELECT 
    tc.table_schema,
    tc.table_name,
    kcu.column_name,
    ccu.table_schema AS foreign_table_schema,
    ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu 
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu 
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema != 'public'
    AND tc.table_schema != ccu.table_schema;

-- Should return 0 rows (no cross-schema FKs)
```

---
```

---

## Indexes Strategy

- **Primary Keys**: All tables have UUID primary keys
- **Foreign Keys**: Indexed for join performance
- **Soft Deletes**: Index on `deleted_at` columns
- **Search Fields**: Index on frequently queried fields (email, subdomain, URL, dates)
- **Composite Indexes**: Consider for complex queries (e.g., workspace_id + deleted_at)

---

## Data Retention Policy

- **Checks**: Keep last 90 days, archive older data
- **Alerts**: Keep indefinitely (user history)
- **Insights**: Keep indefinitely
- **Usage Logs**: Keep 12 months for billing
- **Reports**: Keep indefinitely

---

## Backup Strategy

- **Daily**: Full database backup
- **Hourly**: Incremental backups
- **Point-in-time recovery**: Enabled
- **Retention**: 30 days

---

## Performance Considerations

1. **Connection Pooling**: Use pgbouncer or application-level pooling
2. **Read Replicas**: For reporting and analytics queries
3. **Partitioning**: Consider partitioning `checks` and `usage_logs` by date
4. **Materialized Views**: For dashboard statistics
5. **Caching**: Redis for frequently accessed data (workspace counts, usage stats)

---

## Security

- **Encryption at rest**: Enable PostgreSQL transparent data encryption
- **Encryption in transit**: SSL/TLS connections only (require SSL in pg_hba.conf)
- **Row-Level Security**: Not needed (schema isolation provides tenant separation)
- **Password Hashing**: bcrypt with cost factor 12 (in application layer)
- **API Keys**: Store hashed, not plain text
- **Tenant Isolation**: 
  - Strictly enforce `SET search_path` before every query
  - Validate user belongs to organization before granting access
  - Never expose tenant schema names to frontend
- **Cross-Tenant Prevention**:
  - Always extract tenant from subdomain (not from user input)
  - Validate user is member of organization via `organization_members` table
  - Log all cross-schema access attempts

---

## Query Examples

### Check User Access to Organization
```sql
-- Returns role if user has access, NULL otherwise
SELECT om.role, o.schema_name
FROM public.organization_members om
JOIN public.organizations o ON om.organization_id = o.id
WHERE om.user_id = $1 AND o.subdomain = $2;
```

### List All Organizations for User
```sql
SELECT 
    o.id,
    o.name,
    o.subdomain,
    o.schema_name,
    om.role,
    om.joined_at
FROM public.organizations o
JOIN public.organization_members om ON o.id = om.organization_id
WHERE om.user_id = $1
ORDER BY om.joined_at DESC;
```

### Set Search Path and Query (Go Example)
```go
// In repository layer
func (r *WorkspaceRepository) List(ctx context.Context, tenant string) ([]*Workspace, error) {
    // Set search path
    _, err := r.db.ExecContext(ctx, "SET search_path TO "+tenant)
    if err != nil {
        return nil, err
    }
    
    // Now query in tenant schema
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, type, description, created_at 
        FROM workspaces 
        WHERE deleted_at IS NULL
    `)
    // ... process rows
}
```

### Transaction with Tenant Schema
```go
func (r *PageRepository) Create(ctx context.Context, tenant string, page *Page) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Set search path in transaction
    _, err = tx.ExecContext(ctx, "SET search_path TO "+tenant)
    if err != nil {
        return err
    }
    
    // Insert page
    err = tx.QueryRowContext(ctx, `
        INSERT INTO pages (name, url, workspace_id, created_by)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, page.Name, page.URL, page.WorkspaceID, page.CreatedBy).Scan(&page.ID)
    if err != nil {
        return err
    }
    
    // Insert tags
    for _, tag := range page.Tags {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO page_tags (page_id, tag) VALUES ($1, $2)
        `, page.ID, tag)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

---

## Testing Strategy

### Unit Tests (Application Layer)
- Use in-memory repositories (`*_memory.go`)
- No database required
- Fast execution

### Integration Tests (Infrastructure Layer)
- Use testcontainers-go for PostgreSQL
- Create test tenant schemas
- Clean up after each test

```go
func TestWorkspaceRepository_Create(t *testing.T) {
    // Setup testcontainer
    ctx := context.Background()
    container, db := setupTestDB(ctx, t)
    defer container.Terminate(ctx)
    
    // Create test tenant schema
    tenant := "test_tenant"
    _, err := db.Exec("SELECT create_tenant_schema($1)", tenant)
    require.NoError(t, err)
    
    // Test repository
    repo := NewWorkspacePostgresRepository(db)
    workspace := &Workspace{Name: "Test", Type: "Personal"}
    
    err = repo.Create(ctx, tenant, workspace)
    assert.NoError(t, err)
    assert.NotEmpty(t, workspace.ID)
    
    // Cleanup
    _, _ = db.Exec("DROP SCHEMA " + tenant + " CASCADE")
}
```

---

## Monitoring & Maintenance

### Schema Management
- **List all tenant schemas**:
  ```sql
  SELECT schema_name 
  FROM information_schema.schemata 
  WHERE schema_name NOT IN ('public', 'information_schema', 'pg_catalog', 'pg_toast');
  ```

- **Get schema size**:
  ```sql
  SELECT 
      schema_name,
      pg_size_pretty(sum(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename)))::bigint) AS size
  FROM pg_tables
  WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
  GROUP BY schema_name
  ORDER BY sum(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename))) DESC;
  ```

### Orphan Schema Detection
```sql
-- Find schemas without corresponding organization
SELECT s.schema_name
FROM information_schema.schemata s
WHERE s.schema_name NOT IN ('public', 'information_schema', 'pg_catalog', 'pg_toast')
  AND s.schema_name NOT IN (SELECT schema_name FROM public.organizations);
```

### Tenant Schema Deletion (When Organization Deleted)
```sql
-- In application (Go)
func (r *OrganizationRepository) Delete(ctx context.Context, orgID string) error {
    // Get schema name
    var schemaName string
    err := r.db.QueryRowContext(ctx, 
        "SELECT schema_name FROM public.organizations WHERE id = $1", orgID).Scan(&schemaName)
    if err != nil {
        return err
    }
    
    // Soft delete organization (or hard delete)
    _, err = r.db.ExecContext(ctx, 
        "UPDATE public.organizations SET deleted_at = NOW() WHERE id = $1", orgID)
    if err != nil {
        return err
    }
    
    // Drop schema (background job recommended for safety)
    // _, err = r.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
    // return err
    
    return nil
}
```

**Recommendation**: Don't drop schemas immediately. Mark organization as deleted and schedule schema drop after grace period (e.g., 30 days).
