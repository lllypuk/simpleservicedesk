# HTMX Interface Specification

## Overview

This specification outlines the implementation of an HTMX-based web interface for the SimpleServiceDesk project. The interface will provide a modern, interactive user experience while leveraging the existing API infrastructure.

## Objectives

- Add a responsive web interface using HTMX for dynamic interactions
- Maintain separation of concerns between API and web interface
- Reuse existing business logic through service layer refactoring
- Provide role-based access control for different user types
- Ensure accessibility and modern UX patterns

## Architecture Changes

### Service Layer Introduction

To avoid code duplication between API handlers and HTMX handlers, we'll introduce a service layer:

```
internal/
├── application/
│   ├── services/           # NEW: Business logic services
│   │   ├── user_service.go
│   │   ├── ticket_service.go
│   │   ├── category_service.go
│   │   └── organization_service.go
│   ├── handlers/
│   │   ├── api/            # Existing API handlers (refactored)
│   │   └── web/            # NEW: HTMX handlers
│   └── templates/          # NEW: HTML templates
```

### Service Layer Structure

Each service will contain:
- Business logic validation
- Repository interactions
- Domain object creation/updates
- Error handling and mapping

```go
type UserService interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
    GetUser(ctx context.Context, id uuid.UUID) (*User, error)
    UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*User, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    ListUsers(ctx context.Context, filter UserFilter) ([]*User, int64, error)
    UpdateUserRole(ctx context.Context, id uuid.UUID, role Role) error
}
```

## Web Interface Design

### Technology Stack

- **Backend**: Go with Echo framework
- **Frontend**: HTMX 2 (avoid JavaScript)
- **Styling**: PicoCSS
- **Templates**: Go html/template
- **Icons**: Unicode symbols or embedded SVG icons

### Page Structure

#### Authentication Pages
- `/login` - User login form
- `/register` - User registration (if enabled)
- `/logout` - Logout action

#### Dashboard Pages
- `/` - Main dashboard (redirect based on role)
- `/dashboard/admin` - Admin overview
- `/dashboard/agent` - Agent workbench
- `/dashboard/user` - User portal

#### Functional Pages
- `/tickets` - Ticket management
- `/tickets/new` - Create new ticket
- `/tickets/{id}` - View/edit ticket
- `/users` - User management (admin/agent only)
- `/organizations` - Organization management
- `/categories` - Category management
- `/profile` - User profile management

### HTMX Integration Patterns

#### Common Patterns
1. **Partial Page Updates**: Update specific sections without full page reload
2. **Form Submissions**: Submit forms with HTMX and show validation errors
3. **Modal Dialogs**: Load content dynamically in modals
4. **Infinite Scroll**: Load more content as user scrolls
5. **Real-time Updates**: WebSocket or SSE for live ticket updates

#### Example HTMX Handlers

```go
// Web handlers for HTMX responses
func (h *WebHandlers) GetTicketsPartial(c echo.Context) error {
    // Return partial HTML for tickets list
}

func (h *WebHandlers) PostTicketForm(c echo.Context) error {
    // Handle form submission, return updated HTML or errors
}

func (h *WebHandlers) GetTicketModal(c echo.Context) error {
    // Return modal content for ticket details
}
```

## Implementation Plan

### Phase 1: Service Layer Refactoring

1. **Create Service Interfaces**
   - Define service interfaces in `internal/application/services/`
   - Move business logic from handlers to services
   - Keep repository interactions in services

2. **Refactor Existing API Handlers**
   - Move existing handlers to `internal/application/handlers/api/`
   - Update handlers to use services instead of direct repository calls
   - Maintain existing API contracts

3. **Update HTTP Server Setup**
   - Modify `http_server.go` to support both API and web routes
   - Add service layer to dependency injection

### Phase 2: Web Infrastructure

1. **Template System**
   - Create base layout template
   - Implement component-based template structure
   - Add helper functions for common UI patterns

2. **Asset Management**
   - Set up static file serving (CSS, images)
   - Add PicoCSS and HTMX 2.0 CDN links
   - No JavaScript build process required

3. **Authentication Middleware**
   - Implement session-based authentication for web interface
   - Add role-based access control middleware
   - Create login/logout flows

### Phase 3: Core Web Pages

1. **Authentication Pages**
   - Login form with validation
   - Registration form (if enabled)
   - Password reset functionality

2. **Dashboard Implementation**
   - Role-specific dashboards
   - Key metrics and widgets
   - Navigation structure

3. **Basic CRUD Operations**
   - Ticket list/create/edit pages
   - User management pages
   - Organization management

### Phase 4: Advanced HTMX Features

1. **Dynamic Interactions**
   - Real-time form validation using HTMX
   - Dependent dropdowns with HTMX triggers
   - Auto-save functionality with HTMX

2. **Enhanced UX**
   - Modal dialogs using HTMX and CSS
   - Inline editing with HTMX
   - Progressive enhancement without JavaScript

3. **Performance Optimization**
   - Lazy loading with HTMX infinite scroll
   - Efficient pagination using HTMX
   - Server-side caching strategies

## Directory Structure

```
internal/
├── application/
│   ├── services/
│   │   ├── interfaces.go           # Service interfaces
│   │   ├── user_service.go
│   │   ├── ticket_service.go
│   │   ├── category_service.go
│   │   └── organization_service.go
│   ├── handlers/
│   │   ├── api/                    # Existing API handlers (moved)
│   │   │   ├── users/
│   │   │   ├── tickets/
│   │   │   ├── categories/
│   │   │   └── organizations/
│   │   └── web/                    # New web handlers
│   │       ├── auth.go
│   │       ├── dashboard.go
│   │       ├── tickets.go
│   │       ├── users.go
│   │       ├── categories.go
│   │       └── organizations.go
│   ├── templates/
│   │   ├── layouts/
│   │   │   ├── base.html
│   │   │   └── auth.html
│   │   ├── components/
│   │   │   ├── navigation.html
│   │   │   ├── forms.html
│   │   │   └── modals.html
│   │   ├── pages/
│   │   │   ├── dashboard/
│   │   │   ├── tickets/
│   │   │   ├── users/
│   │   │   └── auth/
│   │   └── partials/               # HTMX partial responses
│   └── middleware/
│       ├── auth.go
│       └── session.go
├── web/                            # Static assets
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   └── assets.go                   # Embedded assets
```

## Template Architecture

### Base Layout

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - SimpleServiceDesk</title>
    <script src="https://unpkg.com/htmx.org@2.0.0"></script>
    <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@latest/css/pico.min.css">
    <link href="/static/css/custom.css" rel="stylesheet">
</head>
<body>
    {{template "navigation" .}}

    <main class="container">
        {{template "content" .}}
    </main>

    {{template "modals" .}}
</body>
</html>
```

### Component Templates

- **Navigation**: Role-based navigation menu
- **Forms**: Reusable form components with validation
- **Tables**: Data tables with sorting and filtering
- **Cards**: Information display cards
- **Modals**: Dialog components

## Authentication & Authorization

### Session Management

```go
type SessionStore interface {
    Get(ctx context.Context, sessionID string) (*Session, error)
    Set(ctx context.Context, session *Session) error
    Delete(ctx context.Context, sessionID string) error
}

type Session struct {
    ID       string
    UserID   uuid.UUID
    Role     Role
    Email    string
    Name     string
    ExpireAt time.Time
}
```

### Middleware Chain

1. **Session Middleware**: Extract and validate session
2. **Auth Middleware**: Ensure user is authenticated
3. **Role Middleware**: Check role-based permissions
4. **CSRF Middleware**: Protect against CSRF attacks

## Error Handling

### Web-Specific Error Responses

- **Validation Errors**: Inline form validation with HTMX
- **Authorization Errors**: Redirect to login or show access denied
- **Server Errors**: User-friendly error pages
- **HTMX Errors**: Return partial error content

### Error Templates

```html
<!-- Inline validation error -->
<small id="error-{field}" role="alert" style="color: var(--pico-del-color);">
    {error_message}
</small>

<!-- Page-level error -->
<article style="background-color: var(--pico-del-background-color); border: 1px solid var(--pico-del-color);">
    {error_message}
</article>
```

## Testing Strategy

### Web Handler Testing

```go
func TestWebTicketCreate(t *testing.T) {
    // Setup service mocks
    // Create test request with form data
    // Assert HTML response contains expected elements
    // Verify service calls
}
```

### Integration Testing

- **E2E Tests**: Full user workflows using browser automation
- **Template Tests**: Verify template rendering with various data
- **HTMX Tests**: Test dynamic interactions and partial updates

## Security Considerations

### Input Validation

- All form inputs must be validated server-side
- HTML escaping for user-generated content
- CSRF protection for state-changing operations
- No client-side JavaScript vulnerabilities

### Content Security Policy

```
Content-Security-Policy: default-src 'self';
                        script-src 'self' https://unpkg.com/htmx.org;
                        style-src 'self' https://unpkg.com/@picocss/pico 'unsafe-inline';
```

### Session Security

- Secure session cookies
- Session timeout and renewal
- Protection against session fixation

## Performance Considerations

### Optimization Strategies

1. **Template Caching**: Cache compiled templates
2. **Minimal Assets**: PicoCSS provides small footprint, HTMX 2.0 is lightweight
3. **Database Query Optimization**: Efficient pagination and filtering
4. **CDN Usage**: Serve PicoCSS and HTMX from CDN

### Monitoring

- Page load times (faster without JavaScript frameworks)
- HTMX request performance
- Database query performance
- Error rates
- Client-side rendering performance

## Deployment

### Build Process

```makefile
.PHONY: build-web
build-web:
	# Embed static assets
	# Build Go binary with templates
	# No CSS build process needed (using PicoCSS CDN)

.PHONY: dev-web
dev-web:
	# Start with hot reload
	# Watch for template changes
	# No CSS build needed
```

### Environment Configuration

```env
WEB_ENABLED=true
SESSION_SECRET=your-secret-key
STATIC_FILES_PATH=/static
TEMPLATE_PATH=internal/application/templates
HTMX_VERSION=2.0.0
PICOCSS_VERSION=latest
```

## Migration Strategy

### Gradual Implementation

1. **Start with service layer refactoring** - No breaking changes
2. **Add web routes alongside API** - Both interfaces work simultaneously
3. **Implement core pages first** - Essential functionality
4. **Add advanced features incrementally** - Enhanced UX features

### Rollback Plan

- API remains fully functional throughout migration
- Web interface can be disabled via configuration
- Database schema remains unchanged

## Future Enhancements

### Potential Additions

- **Real-time Notifications**: Server-Sent Events with HTMX for live updates
- **Mobile App Support**: API-first design enables mobile app development
- **Theme Customization**: PicoCSS theme variants and custom CSS variables
- **Advanced Reporting**: Server-rendered dashboards with HTMX interactions
- **Integration APIs**: Third-party service integrations

### Scalability Considerations

- **Horizontal Scaling**: Session store externalization
- **Caching Layer**: Redis for session and template caching
- **Minimal Asset Load**: PicoCSS and HTMX from CDN reduce server load
- **Load Balancing**: Multi-instance deployment support
- **Server-Side Rendering**: Better SEO and performance than client-side apps

## Conclusion

This specification provides a comprehensive plan for adding HTMX-based web interface to the SimpleServiceDesk project. The approach ensures:

- **Code Reusability**: Service layer prevents duplication
- **Maintainability**: Clear separation of concerns
- **Scalability**: Architecture supports future growth
- **User Experience**: Modern, accessible interface without JavaScript complexity
- **Security**: Proper authentication, authorization, and reduced attack surface
- **Performance**: Faster loading and better SEO with server-side rendering

The implementation should be done incrementally, ensuring that existing API functionality remains unaffected while gradually introducing the new web interface capabilities.
