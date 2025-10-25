# Pulzifi Backend Analysis

## Overview
Pulzifi is a competitive intelligence and website monitoring platform that allows users to track changes on competitor websites, receive alerts, and get AI-powered insights.

**Architecture:** Go + Hexagonal + Vertical Slicing + Screaming Architecture + Multi-Tenant by Organization

**Key Principles:**
- Each module is an independent, deployable microservice
- Modules communicate via gRPC (sync) or Kafka (async)
- No direct imports between modules
- Multi-tenant by organization: one user can belong to multiple organizations (multiple tenants)

---

## Multi-Tenancy Model

### Tenant = Organization
- Each **organization** has its own PostgreSQL schema
- A **user** can belong to **multiple organizations** (multiple tenants)
- Tenant identification: based on subdomain (e.g., `jcsoftdev-inc.pulzifi.com`)
- User switches context when accessing different organizations

### Data Isolation
- **Public schema**: users, organizations, organization_members, auth tokens
- **Tenant schemas**: workspaces, pages, checks, alerts, insights, etc.
- `SET search_path TO <tenant_schema>` on every tenant-scoped query

### User-Tenant Relationship
```
User (dania@example.com)
  ├── Organization: jcsoftdev INC (tenant: jcsoftdev_inc)
  │   └── Role: ADMIN
  ├── Organization: Toyota Corp (tenant: toyota_corp)
  │   └── Role: MEMBER
  └── Organization: Nissan Ltd (tenant: nissan_ltd)
      └── Role: MEMBER
```

---

## Features Analysis (Based on UI Screenshots)

### 1. **Authentication & User Management**
- User registration/login
- Organization management (jcsoftdev INC)
- Multi-user support (Dania Morales - ADMIN)
- Team member management

### 2. **Workspace Management**
- Create/manage multiple workspaces (Toyota, Jeep, Nissan)
- Workspace types: Personal, Competitor
- Workspace metadata: creation date, page count
- Active/Deleted workspace states
- Search workspaces functionality

### 3. **Page Monitoring**
- Add webpage URLs to workspaces
- Configure monitoring frequency (Every day, Every 3 hours)
- Tag system for pages (Hero section, Follow up, Promotion, Price Tracking, Job offer, Update, Questions)
- Thumbnail capture
- Last change tracking
- Check history
- Visual and Text change detection

### 4. **Alert System**
- Real-time change notifications
- Alert types: Visual, Text, HTML
- Recent alerts dashboard
- Change timeline (Today, Yesterday)
- Alert status tracking

### 5. **AI-Powered Insights**
- Visual Pulse analysis
- Marketing Lens insights
- Market Analysis
- Promotion opportunity detection
- Job recommendation tracking
- Article publication detection
- Comment monitoring
- Navigation menu change detection
- Custom insight triggers
- Text change search functionality

### 6. **Reporting**
- Generate reports per page
- Date-based report filtering
- Copy insights functionality
- Visual comparison screenshots
- Detailed change descriptions

### 7. **Dashboard & Analytics**
- Usage statistics (checks left, refill countdown)
- Workspace statistics (count, page count, daily checks)
- Found changes visualization (by workspace)
- Monthly check tracking (100/2000, Usage 5%)
- Last scanning timestamp

### 8. **Integrations**
- Slack
- Teams
- Telegram
- Spreadsheet export
- Phone number notifications

### 9. **Settings & Configuration**
- Check frequency settings
- Work days/hours scheduling
- Timezone configuration (America/Boise)
- Block ads and cookie banners option
- Advanced frequency scheduling
- Page-specific settings

---

## Core Backend Modules (Independent Microservices)

### Module 1: `auth` (Public Schema Only)
**Responsibility:** User authentication, JWT management, password reset
**Communication:** Provides gRPC service for user validation
**Schema:** Public only
- User registration, login, logout
- JWT token generation/validation/refresh
- Password reset flow
- Session management

### Module 2: `organization` (Public Schema Only)
**Responsibility:** Organization management, tenant creation, member management
**Communication:** Provides gRPC service, consumes `auth` module
**Schema:** Public only
- Organization CRUD
- Tenant schema creation (triggers on organization creation)
- Member invitation/removal
- Role-based access control (ADMIN, MEMBER)
- List user's organizations

### Module 3: `workspace` (Tenant Schema)
**Responsibility:** Workspace management within organizations
**Communication:** Provides gRPC service, consumes `organization` module
**Schema:** Tenant-specific
- Workspace CRUD
- Workspace types (Personal, Competitor)
- Soft delete functionality
- Search and statistics
- **Tenant extracted from gRPC metadata**

### Module 4: `page` (Tenant Schema)
**Responsibility:** Page URL management, tagging, monitoring configuration
**Communication:** Provides gRPC service, consumes `workspace` module
**Schema:** Tenant-specific
- Page CRUD
- Tag management
- Monitoring configuration (frequency, schedule, timezone)
- **Tenant extracted from gRPC metadata**

### Module 5: `monitoring` (Tenant Schema + Background Workers)
**Responsibility:** Check execution, screenshot capture, change detection
**Communication:** Provides gRPC service, consumes `page` module, publishes events
**Schema:** Tenant-specific
- Manual and scheduled check execution
- Screenshot capture (Puppeteer/Playwright)
- HTML snapshot
- Change detection (visual, text, HTML diff)
- Storage integration (S3/MinIO)
- **Background workers for scheduled checks**

### Module 6: `alert` (Tenant Schema)
**Responsibility:** Alert management and email notification
**Communication:** Subscribes to `monitoring` events via Kafka, provides gRPC service
**Schema:** Tenant-specific
- Alert creation from detected changes
- Alert types: Visual, Text, HTML
- Mark as read/unread
- Real-time WebSocket notifications
- Alert filtering and search
- **Email notifications with HTML templates** (alert reports with screenshots, page info, change summary)
- **User email preferences management** (per workspace and per page)
- **Email delivery tracking and logging**
- **Unsubscribe functionality**
- **Async email sending with retry logic**

### Module 7: `insight` (Tenant Schema + AI Integration)
**Responsibility:** AI-powered insights generation
**Communication:** Provides gRPC service, subscribes to `monitoring` events
**Schema:** Tenant-specific
- AI analysis integration (OpenAI, Anthropic)
- Insight generation (Visual Pulse, Marketing Lens, Market Analysis)
- Custom insight rules
- Text change search
- **Background workers for insight processing**

### Module 8: `report` (Tenant Schema)
**Responsibility:** Report generation and export
**Communication:** Provides gRPC service, consumes `insight` and `monitoring` modules
**Schema:** Tenant-specific
- Report generation per page
- Date filtering
- PDF/Excel export
- Report sharing

### Module 9: `integration` (Tenant Schema)
**Responsibility:** External service integrations
**Communication:** Provides gRPC service, subscribes to `alert` events
**Schema:** Tenant-specific
- Slack, Teams, Telegram integration
- Spreadsheet export
- SMS notifications
- Webhook management
- **Background workers for notification dispatch**

### Module 10: `usage` (Tenant Schema)
**Responsibility:** Usage tracking and billing
**Communication:** Provides gRPC service, subscribes to `monitoring` events
**Schema:** Tenant-specific
- Check quota management
- Usage tracking per period
- Monthly refill
- Usage history
- **Background workers for quota refill**

### ~~Module 11: `gateway`~~ → ELIMINADO (no es un módulo)

**En su lugar: Load Balancer (Nginx/Traefik/Kong)**
- Infraestructura, NO código
- Termina SSL/TLS
- Extrae subdomain → header `X-Tenant`
- Enruta por path: `/api/auth/*` → auth module, `/api/workspaces/*` → workspace module
- Rate limiting

**Cada módulo expone su propia REST API:**
- HTTP server en cada módulo (además de gRPC)
- Extrae tenant desde header `X-Tenant`
- Valida JWT en middleware propio
- Responde directamente al frontend

---

## Technical Requirements

### Multi-Tenancy Architecture
- **Strategy**: Organization-based (schema per tenant)
- **Tenant Identification**: Subdomain extraction via Load Balancer (e.g., `jcsoftdev-inc.pulzifi.com` → `jcsoftdev_inc`)
- **Database**: PostgreSQL with dynamic schema per organization
- **Implementation**: 
  - Load Balancer (Nginx/Traefik) extracts subdomain, passes as header `X-Tenant`
  - Each module's HTTP middleware extracts tenant from header
  - Repository layer executes `SET search_path TO <tenant>` before queries
- **User-Tenant Relationship**: A user can belong to multiple organizations
  - User authenticates once
  - Context switches when accessing different organizations (different subdomain)
  - Authorization checked per organization (via organization_members table)

### Inter-Module Communication

#### gRPC (Synchronous)
- All modules expose gRPC services
- Proto definitions in `infrastructure/grpc/proto/<module>.proto`
- gRPC clients in consuming modules: `infrastructure/grpc/<module>_client.go`
- gRPC interceptor for tenant injection/extraction
- Example: `workspace` module calls `organization` module to validate tenant exists

#### Kafka (Asynchronous - Event-Driven)
- Domain events published to Kafka topics
- Each module subscribes to relevant events
- Event format: JSON with tenant included in message
- Examples:
  - `monitoring` publishes `check_completed` → `alert` subscribes
  - `alert` publishes `alert_created` → `integration` subscribes
  - `monitoring` publishes `check_completed` → `insight` subscribes
  - `monitoring` publishes `check_completed` → `usage` subscribes

### Module Independence
- Each module has its own `main.go` (independent deployment)
- No shared code between modules (except shared/ for technical utilities)
- Each module can scale independently
- Each module has its own database connection
- Modules treat each other as external services

### Background Jobs
- Scheduled monitoring checks (cron-based scheduler)
- Screenshot capture and processing
- Change detection processing
- AI insight generation
- Alert dispatching
- Notification sending
- Usage quota refill (monthly)
- **Technology**: Asynq (Redis-based) or Kafka consumers

### External Services
- AI/LLM API for insights (OpenAI, Anthropic, etc.)
- Screenshot service (Playwright, Puppeteer as library or service)
- Email service (SendGrid, AWS SES) for alert notifications
- Notification services (Slack, Teams, Telegram APIs)
- SMS service (Twilio)

### Real-time Features
- WebSocket for live alerts (in `alert` module HTTP server)
- Server-Sent Events (SSE) for dashboard updates (in modules that need it)

---

## gRPC Service Definitions (High-Level)

### Auth Module
```protobuf
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RequestPasswordReset(PasswordResetRequest) returns (PasswordResetResponse);
  rpc ResetPassword(ResetPasswordRequest) returns (ResetPasswordResponse);
}
```

### Organization Module
```protobuf
service OrganizationService {
  rpc CreateOrganization(CreateOrgRequest) returns (CreateOrgResponse);
  rpc GetOrganization(GetOrgRequest) returns (GetOrgResponse);
  rpc GetOrganizationBySubdomain(GetOrgBySubdomainRequest) returns (GetOrgResponse);
  rpc UpdateOrganization(UpdateOrgRequest) returns (UpdateOrgResponse);
  rpc DeleteOrganization(DeleteOrgRequest) returns (DeleteOrgResponse);
  rpc InviteMember(InviteMemberRequest) returns (InviteMemberResponse);
  rpc RemoveMember(RemoveMemberRequest) returns (RemoveMemberResponse);
  rpc ListMembers(ListMembersRequest) returns (ListMembersResponse);
  rpc GetUserOrganizations(GetUserOrgsRequest) returns (GetUserOrgsResponse);
}
```

### Workspace Module (Tenant-scoped)
```protobuf
service WorkspaceService {
  rpc CreateWorkspace(CreateWorkspaceRequest) returns (CreateWorkspaceResponse);
  rpc ListWorkspaces(ListWorkspacesRequest) returns (ListWorkspacesResponse);
  rpc GetWorkspace(GetWorkspaceRequest) returns (GetWorkspaceResponse);
  rpc UpdateWorkspace(UpdateWorkspaceRequest) returns (UpdateWorkspaceResponse);
  rpc DeleteWorkspace(DeleteWorkspaceRequest) returns (DeleteWorkspaceResponse);
  rpc GetWorkspaceStatistics(GetStatsRequest) returns (GetStatsResponse);
}
// Note: Tenant passed in gRPC metadata
```

### Page Module (Tenant-scoped)
```protobuf
service PageService {
  rpc CreatePage(CreatePageRequest) returns (CreatePageResponse);
  rpc ListPagesByWorkspace(ListPagesRequest) returns (ListPagesResponse);
  rpc GetPage(GetPageRequest) returns (GetPageResponse);
  rpc UpdatePage(UpdatePageRequest) returns (UpdatePageResponse);
  rpc DeletePage(DeletePageRequest) returns (DeletePageResponse);
  rpc AddTag(AddTagRequest) returns (AddTagResponse);
  rpc RemoveTag(RemoveTagRequest) returns (RemoveTagResponse);
  rpc UpdateMonitoringConfig(UpdateConfigRequest) returns (UpdateConfigResponse);
}
```

### Monitoring Module (Tenant-scoped)
```protobuf
service MonitoringService {
  rpc ExecuteManualCheck(ExecuteCheckRequest) returns (ExecuteCheckResponse);
  rpc GetCheckHistory(GetCheckHistoryRequest) returns (GetCheckHistoryResponse);
  rpc GetCheck(GetCheckRequest) returns (GetCheckResponse);
}
```

### Alert Module (Tenant-scoped)
```protobuf
service AlertService {
  rpc ListAlerts(ListAlertsRequest) returns (ListAlertsResponse);
  rpc GetAlert(GetAlertRequest) returns (GetAlertResponse);
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  rpc MarkAllAsRead(MarkAllAsReadRequest) returns (MarkAllAsReadResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
  
  // Email notification preferences
  rpc GetUserEmailPreferences(GetUserEmailPreferencesRequest) returns (GetUserEmailPreferencesResponse);
  rpc UpdateUserEmailPreferences(UpdateUserEmailPreferencesRequest) returns (UpdateUserEmailPreferencesResponse);
  rpc GetWorkspaceNotificationPreferences(GetWorkspaceNotificationPreferencesRequest) returns (GetWorkspaceNotificationPreferencesResponse);
  rpc UpdateWorkspaceNotificationPreferences(UpdateWorkspaceNotificationPreferencesRequest) returns (UpdateWorkspaceNotificationPreferencesResponse);
  rpc GetPageNotificationPreferences(GetPageNotificationPreferencesRequest) returns (GetPageNotificationPreferencesResponse);
  rpc UpdatePageNotificationPreferences(UpdatePageNotificationPreferencesRequest) returns (UpdatePageNotificationPreferencesResponse);
  rpc Unsubscribe(UnsubscribeRequest) returns (UnsubscribeResponse);
}
```

---

## Email Notification System (Alert Module)

### Overview
When the monitoring system detects changes on a page, the alert module automatically sends email notifications to relevant users with a detailed alert report including screenshots, change summary, and action links.

### Key Features
- **Immediate alert emails** when changes are detected
- **HTML email templates** with before/after screenshots
- **User email preferences** (enable/disable, frequency, change types)
- **Workspace-level preferences** (opt-in/opt-out per workspace)
- **Page-level preferences** (opt-in/opt-out per page, change type filters)
- **Unsubscribe functionality** (secure token-based)
- **Email delivery tracking** (sent, failed, bounced)
- **Async email sending** with retry logic (doesn't block alert creation)
- **Multi-provider support** (SendGrid, AWS SES)

### Email Notification Flow

```
1. Monitoring Module → Execute Check
   ↓
2. Detect Change → Create Check Record
   ↓
3. Alert Module → Create Alert Record (via Kafka event)
   ↓
4. Alert Module → Determine Recipients
   - Get organization members from public.organization_members
   - Filter by user email preferences (public.users)
   - Filter by workspace preferences (tenant.notification_preferences)
   - Filter by page preferences (tenant.notification_preferences)
   - Filter by change type preferences
   ↓
5. Alert Module → Send Email Notifications (async, in goroutines)
   - Render HTML template with alert data
   - Include before/after screenshots
   - Add unsubscribe link with encrypted token
   ↓
6. Email Service → Send via SendGrid/SES
   ↓
7. Log Email Status → email_logs table
   - Status: pending → sent/failed
   - Track provider message ID
   - Log errors if any
```

### Email Notification Preferences

#### User-Level Preferences (Public Schema)
Stored in `public.users` table:
- `email_verified` (BOOLEAN) - Must be true to receive emails
- `email_notifications_enabled` (BOOLEAN) - Global email notification toggle
- `notification_frequency` (VARCHAR) - 'immediate', 'daily_digest', 'weekly_digest' (MVP: immediate only)

#### Workspace-Level Preferences (Tenant Schema)
Stored in `<tenant>.notification_preferences` table:
- Enable/disable email notifications for entire workspace
- Filter by change types: ['visual', 'text', 'html'] or NULL for all types

#### Page-Level Preferences (Tenant Schema)
Stored in `<tenant>.notification_preferences` table:
- Enable/disable email notifications for specific page (overrides workspace setting)
- Filter by change types: ['visual', 'text', 'html'] or NULL for all types

### Email Template Design

#### Immediate Alert Email (MVP)

**Subject:**
```
[Pulzifi] Change detected: {Page Name} - {Change Type}
```

**HTML Template Structure:**
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>/* Inline CSS for email compatibility */</style>
</head>
<body>
    <div style="max-width: 600px; margin: 0 auto; font-family: Arial, sans-serif;">
        
        <!-- Header with Pulzifi branding -->
        <div style="background: #4F46E5; padding: 20px; text-align: center;">
            <h1 style="color: white; margin: 0;">Pulzifi Alert</h1>
        </div>
        
        <!-- Alert Summary -->
        <div style="padding: 20px; background: #F9FAFB;">
            <h2>Change Detected</h2>
            <p><strong>Page:</strong> {Page Name}</p>
            <p><strong>URL:</strong> {Page URL}</p>
            <p><strong>Workspace:</strong> {Workspace Name}</p>
            <p><strong>Change Type:</strong> {Change Type Badge}</p>
            <p><strong>Detected:</strong> {Timestamp}</p>
        </div>
        
        <!-- Screenshot Comparison (before/after) -->
        <div style="padding: 20px;">
            <h3>Visual Changes</h3>
            <table width="100%">
                <tr>
                    <td width="50%">
                        <p><strong>Before</strong></p>
                        <img src="{Previous Screenshot URL}" style="max-width: 100%;" />
                    </td>
                    <td width="50%">
                        <p><strong>After</strong></p>
                        <img src="{Current Screenshot URL}" style="max-width: 100%;" />
                    </td>
                </tr>
            </table>
        </div>
        
        <!-- Change Description -->
        <div style="padding: 20px; background: #F9FAFB;">
            <h3>Change Summary</h3>
            <p>{Alert Description}</p>
        </div>
        
        <!-- Call to Action -->
        <div style="padding: 20px; text-align: center;">
            <a href="{App URL}/pages/{Page ID}/checks/{Check ID}" 
               style="background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px;">
                View Full Details
            </a>
        </div>
        
        <!-- Footer -->
        <div style="padding: 20px; text-align: center; color: #6B7280; font-size: 12px;">
            <p>You're receiving this email because you're monitoring this page in {Organization Name}.</p>
            <p>
                <a href="{App URL}/settings/notifications">Manage preferences</a> | 
                <a href="{Unsubscribe URL}">Unsubscribe from this page</a>
            </p>
            <p>© 2025 Pulzifi. All rights reserved.</p>
        </div>
        
    </div>
</body>
</html>
```

### Email Service Implementation

#### Service Interface
Located in: `modules/alert/domain/services/email_service.go`

```go
type EmailService interface {
    SendAlertEmail(ctx context.Context, params *AlertEmailParams) error
    SendDigestEmail(ctx context.Context, params *DigestEmailParams) error // Post-MVP
    VerifyEmail(ctx context.Context, email string, token string) error
}

type AlertEmailParams struct {
    Tenant              string
    Recipient           *Recipient
    Alert               *Alert
    Page                *Page
    Workspace           *Workspace
    CurrentScreenshot   string // URL or base64
    PreviousScreenshot  string // URL or base64
    AppBaseURL          string
    UnsubscribeURL      string
}

type Recipient struct {
    UserID string
    Email  string
    Name   string
}
```

#### SendGrid Implementation
Located in: `modules/alert/infrastructure/email/sendgrid_client.go`

```go
type SendGridClient struct {
    apiKey     string
    fromEmail  string
    fromName   string
    templates  *template.Template
}

func (c *SendGridClient) SendAlertEmail(ctx context.Context, params *AlertEmailParams) error {
    // 1. Render HTML template
    var htmlBuffer bytes.Buffer
    err := c.templates.ExecuteTemplate(&htmlBuffer, "alert_immediate.html", params)
    
    // 2. Create email message
    from := mail.NewEmail(c.fromName, c.fromEmail)
    to := mail.NewEmail(params.Recipient.Name, params.Recipient.Email)
    subject := fmt.Sprintf("[Pulzifi] Change detected: %s - %s", params.Page.Name, params.Alert.Type)
    
    message := mail.NewSingleEmail(from, subject, to, "", htmlBuffer.String())
    message.SetHeader("X-Pulzifi-Tenant", params.Tenant)
    message.SetHeader("X-Pulzifi-Alert-ID", params.Alert.ID)
    
    // 3. Send via SendGrid
    client := sendgrid.NewSendClient(c.apiKey)
    response, err := client.SendWithContext(ctx, message)
    
    return err
}
```

#### AWS SES Implementation (Alternative)
Located in: `modules/alert/infrastructure/email/ses_client.go`

Similar implementation using AWS SES SDK.

### Email Sending with Retry Logic

```go
// In application/create_alert/handler.go

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
    tenant := ctx.Value(TenantKey).(string)
    
    // 1. Create alert in database
    alert, err := h.alertRepo.Create(ctx, tenant, req.Alert)
    if err != nil {
        return nil, err
    }
    
    // 2. Get recipients (users who should be notified)
    recipients, err := h.getNotificationRecipients(ctx, tenant, alert)
    if err != nil {
        h.logger.Error("failed to get recipients", zap.Error(err))
        // Don't fail alert creation if email recipient fetching fails
    }
    
    // 3. Send emails asynchronously (don't block alert creation)
    go h.sendEmailNotifications(context.Background(), tenant, alert, recipients)
    
    return &Response{Alert: alert}, nil
}

func (h *Handler) sendEmailWithRetry(ctx context.Context, tenant string, alert *Alert, recipient *Recipient) {
    maxRetries := 3
    backoff := time.Second
    
    // Create email log record (status: pending)
    emailLog := &EmailLog{
        AlertID:         alert.ID,
        RecipientUserID: recipient.UserID,
        RecipientEmail:  recipient.Email,
        Subject:         fmt.Sprintf("[Pulzifi] Change detected: %s", alert.Title),
        Status:          "pending",
    }
    h.emailLogRepo.Create(ctx, tenant, emailLog)
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := h.emailService.SendAlertEmail(ctx, params)
        
        if err == nil {
            // Success - update log
            emailLog.Status = "sent"
            emailLog.SentAt = time.Now()
            h.emailLogRepo.Update(ctx, tenant, emailLog)
            return
        }
        
        if attempt < maxRetries {
            time.Sleep(backoff)
            backoff *= 2 // Exponential backoff
        } else {
            // Final failure - update log
            emailLog.Status = "failed"
            emailLog.ErrorMessage = err.Error()
            h.emailLogRepo.Update(ctx, tenant, emailLog)
        }
    }
}
```

### Unsubscribe Functionality

#### Unsubscribe Token Generation
```go
// Token format: userID:tenantSchema:pageID
token := encrypt(fmt.Sprintf("%s:%s:%s", userID, tenant, pageID))
unsubscribeURL := fmt.Sprintf("%s/api/v1/unsubscribe?token=%s", appBaseURL, token)
```

#### Unsubscribe Endpoint (Public)
```
GET /api/v1/unsubscribe?token={encrypted_token}

Response:
{
  "message": "You have been unsubscribed from notifications for 'Toyota Homepage' page."
}
```

### Configuration

#### Environment Variables
```env
# Email Service Provider (sendgrid or aws_ses)
EMAIL_PROVIDER=sendgrid

# SendGrid Configuration
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM_EMAIL=notifications@pulzifi.com
SENDGRID_FROM_NAME=Pulzifi Alerts

# AWS SES Configuration (if using SES)
AWS_REGION=us-east-1
AWS_SES_FROM_EMAIL=notifications@pulzifi.com
AWS_SES_FROM_NAME=Pulzifi Alerts

# Application URLs
APP_BASE_URL=https://app.pulzifi.com

# Email Templates Path
EMAIL_TEMPLATES_PATH=./modules/alert/infrastructure/email/templates
```

### Email Delivery Tracking

All sent emails are logged in `<tenant>.email_logs` table:
- `alert_id` - Associated alert
- `recipient_user_id` - User who received email
- `recipient_email` - Email address
- `subject` - Email subject
- `status` - 'pending', 'sent', 'failed', 'bounced'
- `provider` - 'sendgrid' or 'aws_ses'
- `provider_message_id` - Provider's message ID for tracking
- `error_message` - Error details if failed
- `sent_at` - Timestamp when email was sent

### Security Considerations

1. **Email Verification**: Only send to verified email addresses (`email_verified = true`)
2. **Unsubscribe Token Security**: Encrypt with tenant-specific key, include expiration
3. **Rate Limiting**: Limit emails per user/tenant per day (prevent spam)
4. **SPF/DKIM/DMARC**: Configure DNS records for email authentication
5. **PII Protection**: Never log email content, mask emails in logs

### Performance Considerations

1. **Async Sending**: Always send emails in goroutines (don't block alert creation)
2. **Retry Logic**: Exponential backoff for failed sends (max 3 retries)
3. **Template Caching**: Parse templates once at startup
4. **Rate Limiting**: Respect provider rate limits (SendGrid: 100/sec)
5. **Batch Sending**: For digest emails (post-MVP), batch multiple alerts

### Cost Estimation

#### SendGrid Pricing
- **Free Tier**: 100 emails/day
- **Essentials**: $19.95/month - 50,000 emails
- **Pro**: $89.95/month - 100,000 emails

#### AWS SES Pricing
- **First 62,000 emails/month**: FREE (via EC2)
- **Additional emails**: $0.10 per 1,000 emails

#### Recommendation
- **MVP**: Start with SendGrid Free Tier (100 emails/day)
- **Production**: Switch to AWS SES if sending > 100,000 emails/month

### Post-MVP Email Features

1. **Daily Digest Emails**: Summary of all changes in last 24 hours
2. **Weekly Digest Emails**: Summary of all changes in last 7 days
3. **Email Open Tracking**: Track when users open emails
4. **Click Tracking**: Track link clicks in emails
5. **A/B Testing**: Test different email templates
6. **Custom Templates**: Per-organization email branding
7. **Interactive Emails**: Approve/dismiss alerts from email

---

### Insight Module (Tenant-scoped)
```protobuf
service InsightService {
  rpc ListInsightsByPage(ListInsightsRequest) returns (ListInsightsResponse);
  rpc GetInsight(GetInsightRequest) returns (GetInsightResponse);
  rpc CreateInsightRule(CreateRuleRequest) returns (CreateRuleResponse);
  rpc UpdateInsightRule(UpdateRuleRequest) returns (UpdateRuleResponse);
  rpc DeleteInsightRule(DeleteRuleRequest) returns (DeleteRuleResponse);
  rpc SearchTextChanges(SearchRequest) returns (SearchResponse);
}
```

### Report Module (Tenant-scoped)
```protobuf
service ReportService {
  rpc GenerateReport(GenerateReportRequest) returns (GenerateReportResponse);
  rpc ListReportsByPage(ListReportsRequest) returns (ListReportsResponse);
  rpc GetReport(GetReportRequest) returns (GetReportResponse);
  rpc ExportReport(ExportReportRequest) returns (ExportReportResponse);
}
```

### Integration Module (Tenant-scoped)
```protobuf
service IntegrationService {
  rpc CreateIntegration(CreateIntegrationRequest) returns (CreateIntegrationResponse);
  rpc UpdateIntegration(UpdateIntegrationRequest) returns (UpdateIntegrationResponse);
  rpc DeleteIntegration(DeleteIntegrationRequest) returns (DeleteIntegrationResponse);
  rpc ListIntegrations(ListIntegrationsRequest) returns (ListIntegrationsResponse);
  rpc ToggleIntegration(ToggleIntegrationRequest) returns (ToggleIntegrationResponse);
}
```

### Usage Module (Tenant-scoped)
```protobuf
service UsageService {
  rpc GetCurrentUsage(GetUsageRequest) returns (GetUsageResponse);
  rpc GetUsageHistory(GetUsageHistoryRequest) returns (GetUsageHistoryResponse);
  rpc TrackCheckUsage(TrackUsageRequest) returns (TrackUsageResponse);
}
```

---

## Kafka Topics & Events

### Topic: `organization.created`
**Publisher:** organization module  
**Subscribers:** None (for future features like welcome emails)  
**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "organization.created",
  "timestamp": "2025-10-25T10:00:00Z",
  "tenant": "jcsoftdev_inc",
  "data": {
    "organization_id": "uuid",
    "name": "jcsoftdev INC",
    "subdomain": "jcsoftdev-inc",
    "schema_name": "jcsoftdev_inc",
    "owner_user_id": "uuid"
  }
}
```

### Topic: `check.completed`
**Publisher:** monitoring module  
**Subscribers:** alert, insight, usage modules  
**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "check.completed",
  "timestamp": "2025-10-25T10:00:00Z",
  "tenant": "jcsoftdev_inc",
  "data": {
    "check_id": "uuid",
    "page_id": "uuid",
    "workspace_id": "uuid",
    "status": "completed",
    "change_detected": true,
    "change_type": "visual",
    "screenshot_url": "s3://...",
    "html_snapshot_url": "s3://..."
  }
}
```

### Topic: `alert.created`
**Publisher:** alert module  
**Subscribers:** integration module  
**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "alert.created",
  "timestamp": "2025-10-25T10:00:00Z",
  "tenant": "jcsoftdev_inc",
  "data": {
    "alert_id": "uuid",
    "page_id": "uuid",
    "workspace_id": "uuid",
    "type": "Visual",
    "title": "Visual change detected on About Us page",
    "description": "..."
  }
}
```

### Topic: `insight.generated`
**Publisher:** insight module  
**Subscribers:** None (for future features)  
**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "insight.generated",
  "timestamp": "2025-10-25T10:00:00Z",
  "tenant": "jcsoftdev_inc",
  "data": {
    "insight_id": "uuid",
    "page_id": "uuid",
    "check_id": "uuid",
    "insight_type": "Marketing Lens",
    "title": "...",
    "content": "..."
  }
}
```

---

## REST API Endpoints (Gateway Module)

The gateway module provides a REST API for the frontend, aggregating data from multiple gRPC services.

### Auth Endpoints
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`

### Organization Endpoints
- `POST /api/v1/organizations`
- `GET /api/v1/organizations` - List user's organizations
- `GET /api/v1/organizations/:id`
- `PUT /api/v1/organizations/:id`
- `DELETE /api/v1/organizations/:id`
- `POST /api/v1/organizations/:id/members`
- `DELETE /api/v1/organizations/:id/members/:userId`
- `GET /api/v1/organizations/:id/members`

### Workspace Endpoints (Tenant-scoped via subdomain)
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces`
- `GET /api/v1/workspaces/:id`
- `PUT /api/v1/workspaces/:id`
- `DELETE /api/v1/workspaces/:id` (soft delete)
- `POST /api/v1/workspaces/:id/restore`
- `GET /api/v1/workspaces/:id/statistics`

### Page Endpoints (Tenant-scoped)
- `POST /api/v1/workspaces/:workspaceId/pages`
- `GET /api/v1/workspaces/:workspaceId/pages`
- `GET /api/v1/pages/:id`
- `PUT /api/v1/pages/:id`
- `DELETE /api/v1/pages/:id`
- `POST /api/v1/pages/:id/tags`
- `DELETE /api/v1/pages/:id/tags/:tag`
- `PUT /api/v1/pages/:id/monitoring-config`

### Monitoring Endpoints (Tenant-scoped)
- `POST /api/v1/pages/:id/check` (manual trigger)
- `GET /api/v1/pages/:id/checks`
- `GET /api/v1/checks/:id`

### Alert Endpoints (Tenant-scoped)
- `GET /api/v1/alerts`
- `GET /api/v1/alerts/recent`
- `GET /api/v1/alerts/:id`
- `PUT /api/v1/alerts/:id/read`
- `PUT /api/v1/alerts/read-all`
- `GET /api/v1/alerts/unread-count`
- `WS /api/v1/ws/alerts` (WebSocket for real-time)

### User Email Preferences Endpoints
- `GET /api/v1/users/me/preferences` - Get user email preferences
- `PUT /api/v1/users/me/preferences` - Update user email preferences

### Notification Preferences Endpoints (Tenant-scoped)
- `GET /api/v1/workspaces/:workspaceId/notifications/preferences` - Get workspace notification preferences
- `PUT /api/v1/workspaces/:workspaceId/notifications/preferences` - Update workspace notification preferences
- `GET /api/v1/pages/:pageId/notifications/preferences` - Get page notification preferences
- `PUT /api/v1/pages/:pageId/notifications/preferences` - Update page notification preferences
- `GET /api/v1/unsubscribe?token={token}` - Unsubscribe from page notifications (public)

### Insight Endpoints (Tenant-scoped)
- `GET /api/v1/pages/:pageId/insights`
- `GET /api/v1/insights/:id`
- `POST /api/v1/insight-rules`
- `PUT /api/v1/insight-rules/:id`
- `DELETE /api/v1/insight-rules/:id`
- `GET /api/v1/insight-rules`
- `POST /api/v1/insights/search-text-changes`

### Report Endpoints (Tenant-scoped)
- `POST /api/v1/pages/:pageId/reports`
- `GET /api/v1/pages/:pageId/reports`
- `GET /api/v1/reports/:id`
- `POST /api/v1/reports/:id/export`

### Integration Endpoints (Tenant-scoped)
- `POST /api/v1/integrations`
- `GET /api/v1/integrations`
- `GET /api/v1/integrations/:id`
- `PUT /api/v1/integrations/:id`
- `DELETE /api/v1/integrations/:id`
- `PUT /api/v1/integrations/:id/toggle`

### Usage Endpoints (Tenant-scoped)
- `GET /api/v1/usage/current`
- `GET /api/v1/usage/history`

### Dashboard Endpoints (Tenant-scoped - Aggregated)
- `GET /api/v1/dashboard/statistics`
  - Aggregates: workspace count, page count, today's checks, usage stats
- `GET /api/v1/dashboard/changes-by-workspace`
- `GET /api/v1/dashboard/recent-activity`

---

## Non-Functional Requirements

### Performance
- Support concurrent monitoring of 1000+ pages per tenant
- gRPC response time < 100ms for read operations
- REST API response time < 200ms (including gRPC call overhead)
- Background jobs must be horizontally scalable
- Database connection pooling per module

### Security
- mTLS between modules (production)
- JWT with expiration and refresh tokens
- Rate limiting per tenant (at gateway level)
- Input validation in gateway and gRPC services
- SQL injection prevention (parameterized queries)
- Tenant isolation validation (prevent cross-tenant access)
- RBAC: validate user belongs to organization before allowing access

### Scalability
- Stateless gRPC servers (horizontal scaling)
- Independent module deployment
- Job queue for background processing (Asynq/Kafka)
- Database connection pooling per module
- Caching layer (Redis) for:
  - User session data
  - Organization metadata
  - Frequent queries (workspace counts, page counts)
- Read replicas for reporting queries

### Reliability
- Retry mechanism for failed gRPC calls (with exponential backoff)
- Circuit breakers for external services
- Dead letter queue for failed jobs
- Database backups (daily full, hourly incremental)
- Health check endpoints per module
- Graceful shutdown for background workers
- Idempotent event handlers (Kafka consumer groups)

---

## Technology Stack

### Core
- **Language**: Go 1.21+
- **gRPC Framework**: google.golang.org/grpc
- **Database**: PostgreSQL 15+
- **Database Driver**: pgx/v5 or sqlx
- **Migration**: golang-migrate/migrate
- **Config**: viper
- **Logging**: zap or zerolog

### Communication
- **gRPC**: google.golang.org/grpc + protobuf
- **Message Queue**: Kafka (or NATS for simpler setup)
- **Job Queue**: Asynq (Redis-based) for background workers

### Storage & Cache
- **Cache**: Redis 7+
- **Object Storage**: AWS S3 or MinIO (for screenshots, reports)

### External Services
- **Screenshot**: Playwright (as library) or Browserless (as service)
- **AI**: OpenAI API or Anthropic Claude API
- **Email**: SendGrid or AWS SES
- **SMS**: Twilio

### Monitoring & Observability
- **Metrics**: Prometheus + Grafana
- **Tracing**: Jaeger or OpenTelemetry
- **Logging**: ELK stack (Elasticsearch, Logstash, Kibana) or Loki
- **APM**: Optional (Datadog, New Relic)

### Testing
- **Unit Tests**: testify/assert, testify/mock
- **Integration Tests**: testcontainers-go (PostgreSQL, Redis, Kafka)
- **gRPC Tests**: Mock gRPC clients
- **Load Tests**: k6 or Gatling

### Deployment
- **Container**: Docker
- **Orchestration**: Kubernetes (Helm charts per module)
- **Service Mesh**: Istio or Linkerd (optional, for advanced traffic management)
- **CI/CD**: GitHub Actions or GitLab CI
- **IaC**: Terraform (for cloud resources)

---

## Deployment Architecture

### Development Environment
```
docker-compose.yml:
  - PostgreSQL (with multiple schemas)
  - Redis
  - Kafka + Zookeeper
  - MinIO (S3-compatible)
  - All modules as separate services
```

### Production Environment (Kubernetes)
```
Namespace: pulzifi-production
├── Ingress (HTTPS termination, subdomain routing)
├── Gateway Deployment (3 replicas)
├── Auth Module Deployment (3 replicas)
├── Organization Module Deployment (3 replicas)
├── Workspace Module Deployment (3 replicas)
├── Page Module Deployment (3 replicas)
├── Monitoring Module Deployment (5 replicas + workers)
├── Alert Module Deployment (3 replicas)
├── Insight Module Deployment (3 replicas + workers)
├── Report Module Deployment (2 replicas)
├── Integration Module Deployment (2 replicas + workers)
├── Usage Module Deployment (2 replicas + workers)
├── PostgreSQL (StatefulSet with read replicas)
├── Redis Cluster (3 nodes)
├── Kafka Cluster (3 brokers)
└── MinIO Cluster (distributed mode)
```

### Service Discovery
- Kubernetes DNS for inter-module communication
- gRPC load balancing via K8s services

---

## Data Flow Examples

### Example 1: User Creates a Page
1. Frontend → Gateway REST API: `POST /api/v1/workspaces/:id/pages`
2. Gateway extracts tenant from subdomain
3. Gateway validates JWT, extracts user_id
4. Gateway → Page gRPC: CreatePage(tenant, user_id, workspace_id, url, ...)
5. Page Module:
   - Validates user belongs to organization (call Organization gRPC)
   - Validates workspace exists (query tenant schema)
   - Creates page in tenant schema
   - Creates monitoring_config record
6. Page Module → Response to Gateway
7. Gateway → Response to Frontend

### Example 2: Scheduled Check Execution (Event-Driven)
1. Monitoring Worker: Fetch pages due for check (from tenant schemas)
2. For each page:
   - Capture screenshot (Playwright)
   - Save HTML snapshot
   - Upload to S3
   - Detect changes (compare with previous)
   - Save check record in tenant schema
3. Monitoring Module → Kafka: Publish `check.completed` event
4. Alert Module (subscriber):
   - Receive event
   - If change_detected: create alert in tenant schema
   - Publish `alert.created` event
   - Send WebSocket notification
5. Insight Module (subscriber):
   - Receive `check.completed` event
   - Enqueue insight generation job
   - Call AI API asynchronously
   - Save insight in tenant schema
6. Usage Module (subscriber):
   - Receive `check.completed` event
   - Decrement quota in tenant schema
   - Log usage
7. Integration Module (subscriber):
   - Receive `alert.created` event
   - Fetch integrations from tenant schema
   - Send notifications (Slack, Teams, etc.)

### Example 3: User Switches Organization
1. Frontend: User selects different organization from dropdown
2. Frontend changes subdomain: `toyota-corp.pulzifi.com`
3. All subsequent requests go to new subdomain
4. Gateway extracts new tenant: `toyota_corp`
5. Gateway validates user is member of this organization (call Organization gRPC)
6. Gateway passes new tenant in gRPC metadata
7. Modules use new tenant schema for all queries

---

## Next Steps

1. ✅ Database schema design (with multi-tenant per organization)
2. ✅ Project structure setup (hexagonal + vertical slicing)
3. ✅ Shared modules implementation (config, database, middleware, logger)
4. 🔄 Module-by-module implementation:
   - Phase 1: Foundation (shared/)
   - Phase 2: Auth module (public schema)
   - Phase 3: Organization module (public schema + tenant creation)
   - Phase 4: Workspace module (tenant schema)
   - Phase 5: Page module (tenant schema)
   - Phase 6: Monitoring module (tenant schema + workers)
   - Phase 7: Alert module (tenant schema + events)
   - Phase 8: Insight module (tenant schema + AI)
   - Phase 9: Report module (tenant schema)
   - Phase 10: Integration module (tenant schema + notifications)
   - Phase 11: Usage module (tenant schema + billing)
   - Phase 12: Gateway module (REST API + aggregation)
5. Testing strategy (unit, integration, e2e)
6. Deployment configuration (K8s, Helm)
7. Documentation (API docs, architecture diagrams)

---

## Architecture Principles (Critical)

### ✅ ALWAYS
1. **Module Independence**: Each module is deployable independently
2. **No Direct Imports**: Modules communicate only via gRPC or Kafka
3. **Tenant Isolation**: Always validate tenant, use `SET search_path`
4. **Vertical Slicing**: Features are organized by what they do (screaming architecture)
5. **Hexagonal Ports**: Interfaces in domain, implementations in infrastructure
6. **DTOs Per Feature**: Each feature has its own request/response DTOs
7. **Event-Driven**: Publish domain events for async communication
8. **Idempotency**: All event handlers must be idempotent

### ❌ NEVER
1. **Cross-Module Imports**: Never import code from another module
2. **Hardcoded Tenant**: Always extract tenant dynamically
3. **Business Logic in Infrastructure**: Keep domain pure
4. **Shared Structs**: Modules serialize/deserialize their own types
5. **SQL Outside Persistence**: Only repositories execute queries
6. **Shared State**: No global variables between modules

### Multi-Tenant Rules
1. **User can belong to multiple organizations** (multiple tenants)
2. **Tenant is derived from subdomain**: `{org-subdomain}.pulzifi.com`
3. **Public schema**: users, organizations, organization_members, auth tokens
4. **Tenant schema**: all business data (workspaces, pages, checks, etc.)
5. **Authorization**: Check `organization_members` table before allowing access
6. **Context Switching**: User switches tenant by changing subdomain
7. **gRPC Metadata**: Tenant passed in every gRPC call metadata
8. **Kafka Messages**: Tenant included in every event payload

---

## Estimated Timeline

- **Phase 1 (Foundation)**: 1 week
- **Phase 2 (Auth)**: 2 weeks
- **Phase 3 (Organization)**: 1.5 weeks
- **Phase 4-5 (Workspace + Page)**: 3.5 weeks
- **Phase 6 (Monitoring)**: 3 weeks
- **Phase 7-11 (Alert, Insight, Report, Integration, Usage)**: 9 weeks
- **Phase 12 (Gateway)**: 1 week
- **Testing & QA**: 2 weeks
- **Deployment & DevOps**: 2 weeks
- **Documentation**: 1 week

**Total**: ~26 weeks (6.5 months) for full production-ready system
**MVP** (Auth + Org + Workspace + Page + Basic Monitoring): ~10 weeks
