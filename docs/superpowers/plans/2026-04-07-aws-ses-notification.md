# AWS SES Notification Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an email notification module that sends transactional emails via AWS SES, triggered by EventBus events, respecting user notification preferences, with Chinese/English/Korean templates.

**Architecture:** New `internal/modules/notification` module following the existing modular pattern. EmailProvider interface (SES + Mock) mirrors the Cobo WalletProvider pattern. NotificationService subscribes to EventBus events, checks user preferences, renders Go html/template files by language, and calls the provider. OTP emails (register, password reset) are sent directly from auth service via a shared EmailSender interface.

**Tech Stack:** Go, AWS SDK v2 (sesv2), Go html/template, existing EventBus, existing SettingsRepository/UserRepository

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `internal/modules/notification/provider/provider.go` | EmailProvider interface + SendInput struct |
| Create | `internal/modules/notification/provider/ses.go` | AWS SES v2 implementation |
| Create | `internal/modules/notification/provider/mock.go` | Mock implementation (logs + records calls) |
| Create | `internal/modules/notification/provider/provider_test.go` | Tests for SES param building |
| Create | `internal/modules/notification/template/renderer.go` | Template renderer (preloads, renders by event+lang) |
| Create | `internal/modules/notification/template/renderer_test.go` | Template rendering tests |
| Create | `internal/modules/notification/template/emails/deposit_confirmed/{en,zh,ko}.html` | Deposit confirmed templates (3 langs) |
| Create | `internal/modules/notification/template/emails/withdraw_completed/{en,zh,ko}.html` | Withdraw completed templates (3 langs) |
| Create | `internal/modules/notification/template/emails/withdraw_failed/{en,zh,ko}.html` | Withdraw failed templates (3 langs) |
| Create | `internal/modules/notification/template/emails/settlement_created/{en,zh,ko}.html` | Settlement created templates (3 langs) |
| Create | `internal/modules/notification/template/emails/password_reset/{en,zh,ko}.html` | Password reset OTP templates (3 langs) |
| Create | `internal/modules/notification/template/emails/register_otp/{en,zh,ko}.html` | Register OTP templates (3 langs) |
| Create | `internal/modules/notification/service/notification_service.go` | Core service: event handlers, pref checks, send logic |
| Create | `internal/modules/notification/service/notification_service_test.go` | Service unit tests |
| Create | `internal/modules/notification/module.go` | Module init, wires provider/renderer/service |
| Modify | `config/config.go` | Add NotificationConfig struct |
| Modify | `config/config.yaml` | Add notification section with defaults |
| Modify | `internal/app/app.go` | Init notification module, register event subscriptions |
| Modify | `internal/modules/auth/service/auth_service.go` | Replace OTP logger.Info with EmailProvider.Send calls |
| Modify | `go.mod` | Add AWS SDK v2 dependencies |

---

### Task 1: Add AWS SDK Dependencies

**Files:**
- Modify: `server/go.mod`

- [ ] **Step 1: Install AWS SDK v2 packages**

```bash
cd /Users/johnny/Work/soveregin/server && go get github.com/aws/aws-sdk-go-v2 github.com/aws/aws-sdk-go-v2/config github.com/aws/aws-sdk-go-v2/service/sesv2
```

- [ ] **Step 2: Tidy modules**

```bash
cd /Users/johnny/Work/soveregin/server && go mod tidy
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add go.mod go.sum && git commit -m "chore: add AWS SDK v2 dependencies for SES integration"
```

---

### Task 2: Add NotificationConfig to Config

**Files:**
- Modify: `server/config/config.go`
- Modify: `server/config/config.yaml`

- [ ] **Step 1: Add NotificationConfig struct and field**

In `server/config/config.go`, add the struct after `LogConfig`:

```go
type NotificationConfig struct {
	UseMock     bool   `yaml:"use_mock"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
	AWSRegion   string `yaml:"aws_region"`
}
```

Add to `Config` struct:

```go
Notification NotificationConfig `yaml:"notification"`
```

- [ ] **Step 2: Add defaults to config.yaml**

Append to `server/config/config.yaml`:

```yaml
notification:
  use_mock: true
  from_address: "noreply@example.com"
  from_name: "Sovereign Fund"
  aws_region: "ap-northeast-2"
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add config/config.go config/config.yaml && git commit -m "feat: add notification config for AWS SES"
```

---

### Task 3: EmailProvider Interface + Mock Implementation

**Files:**
- Create: `server/internal/modules/notification/provider/provider.go`
- Create: `server/internal/modules/notification/provider/mock.go`
- Create: `server/internal/modules/notification/provider/provider_test.go`

- [ ] **Step 1: Write tests for MockProvider**

Create `server/internal/modules/notification/provider/provider_test.go`:

```go
package provider

import (
	"context"
	"testing"
)

func TestMockProviderSend(t *testing.T) {
	mock := NewMockProvider()
	input := SendInput{
		To:      "user@example.com",
		Subject: "Test Subject",
		HTML:    "<p>Hello</p>",
	}

	err := mock.Send(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sent := mock.(*MockProvider).Sent
	if len(sent) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(sent))
	}
	if sent[0].To != "user@example.com" {
		t.Errorf("to = %q, want %q", sent[0].To, "user@example.com")
	}
	if sent[0].Subject != "Test Subject" {
		t.Errorf("subject = %q, want %q", sent[0].Subject, "Test Subject")
	}
}

func TestMockProviderMultipleSends(t *testing.T) {
	mock := NewMockProvider()

	for i := 0; i < 3; i++ {
		err := mock.Send(context.Background(), SendInput{
			To:      "user@example.com",
			Subject: "Test",
			HTML:    "<p>Hello</p>",
		})
		if err != nil {
			t.Fatalf("send %d: unexpected error: %v", i, err)
		}
	}

	sent := mock.(*MockProvider).Sent
	if len(sent) != 3 {
		t.Fatalf("expected 3 sent emails, got %d", len(sent))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/provider/ -v`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Create EmailProvider interface**

Create `server/internal/modules/notification/provider/provider.go`:

```go
package provider

import "context"

// EmailProvider abstracts email sending.
// Implementations: SESProvider (production), MockProvider (dev/test).
type EmailProvider interface {
	Send(ctx context.Context, input SendInput) error
}

type SendInput struct {
	To      string
	Subject string
	HTML    string
}
```

- [ ] **Step 4: Create MockProvider**

Create `server/internal/modules/notification/provider/mock.go`:

```go
package provider

import (
	"context"
	"log/slog"
)

type MockProvider struct {
	Sent []SendInput
}

func NewMockProvider() EmailProvider {
	return &MockProvider{}
}

func (m *MockProvider) Send(_ context.Context, input SendInput) error {
	m.Sent = append(m.Sent, input)
	slog.Info("mock email sent",
		slog.String("to", input.To),
		slog.String("subject", input.Subject),
	)
	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/provider/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/provider/ && git commit -m "feat: add EmailProvider interface and MockProvider"
```

---

### Task 4: SES Provider Implementation

**Files:**
- Create: `server/internal/modules/notification/provider/ses.go`

- [ ] **Step 1: Create SES provider**

Create `server/internal/modules/notification/provider/ses.go`:

```go
package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type SESProvider struct {
	client      *sesv2.Client
	fromAddress string
}

func NewSESProvider(ctx context.Context, region, fromName, fromAddress string) (*SESProvider, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &SESProvider{
		client:      sesv2.NewFromConfig(cfg),
		fromAddress: fmt.Sprintf("%s <%s>", fromName, fromAddress),
	}, nil
}

func (s *SESProvider) Send(ctx context.Context, input SendInput) error {
	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{input.To},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(input.Subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(input.HTML),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ses send email to %s: %w", input.To, err)
	}
	return nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./internal/modules/notification/provider/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/provider/ses.go && git commit -m "feat: add SES email provider implementation"
```

---

### Task 5: Template Renderer

**Files:**
- Create: `server/internal/modules/notification/template/renderer.go`
- Create: `server/internal/modules/notification/template/renderer_test.go`
- Create: `server/internal/modules/notification/template/emails/deposit_confirmed/en.html` (test fixture)

- [ ] **Step 1: Create a minimal test template for testing**

Create `server/internal/modules/notification/template/emails/deposit_confirmed/en.html`:

```html
{{define "subject"}}Your deposit of {{.Amount}} {{.Currency}} has been confirmed{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<body>
<h2>Deposit Confirmed</h2>
<p>{{.Amount}} {{.Currency}} has arrived in your account.</p>
</body>
</html>
{{end}}
```

- [ ] **Step 2: Write renderer tests**

Create `server/internal/modules/notification/template/renderer_test.go`:

```go
package template

import (
	"testing"
)

func TestRendererRenderSuccess(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	data := map[string]string{
		"Amount":   "100.00",
		"Currency": "USDT",
	}

	subject, html, err := r.Render("deposit_confirmed", "en", data)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if subject == "" {
		t.Error("subject should not be empty")
	}
	if html == "" {
		t.Error("html should not be empty")
	}

	wantSubject := "Your deposit of 100.00 USDT has been confirmed"
	if subject != wantSubject {
		t.Errorf("subject = %q, want %q", subject, wantSubject)
	}
}

func TestRendererFallbackToEnglish(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	data := map[string]string{
		"Amount":   "50.00",
		"Currency": "USDT",
	}

	subject, _, err := r.Render("deposit_confirmed", "ja", data)
	if err != nil {
		t.Fatalf("Render with fallback: %v", err)
	}

	if subject == "" {
		t.Error("should fallback to English template")
	}
}

func TestRendererUnknownEvent(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	_, _, err = r.Render("nonexistent_event", "en", nil)
	if err == nil {
		t.Error("expected error for unknown event type")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/template/ -v`
Expected: FAIL — Renderer not defined yet

- [ ] **Step 4: Implement Renderer**

Create `server/internal/modules/notification/template/renderer.go`:

```go
package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Renderer struct {
	templates map[string]map[string]*template.Template // [eventType][lang] -> template
}

func NewRenderer(emailsDir string) (*Renderer, error) {
	templates := make(map[string]map[string]*template.Template)

	entries, err := os.ReadDir(emailsDir)
	if err != nil {
		return nil, fmt.Errorf("read emails dir %s: %w", emailsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		eventType := entry.Name()
		templates[eventType] = make(map[string]*template.Template)

		langFiles, err := filepath.Glob(filepath.Join(emailsDir, eventType, "*.html"))
		if err != nil {
			return nil, fmt.Errorf("glob %s templates: %w", eventType, err)
		}

		for _, f := range langFiles {
			lang := strings.TrimSuffix(filepath.Base(f), ".html")
			tmpl, err := template.ParseFiles(f)
			if err != nil {
				return nil, fmt.Errorf("parse template %s: %w", f, err)
			}
			templates[eventType][lang] = tmpl
		}
	}

	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Render(eventType, lang string, data any) (string, string, error) {
	langMap, ok := r.templates[eventType]
	if !ok {
		return "", "", fmt.Errorf("unknown email event type: %s", eventType)
	}

	tmpl, ok := langMap[lang]
	if !ok {
		tmpl, ok = langMap["en"]
		if !ok {
			return "", "", fmt.Errorf("no template for event %s lang %s (no en fallback)", eventType, lang)
		}
	}

	var subjectBuf, bodyBuf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&subjectBuf, "subject", data); err != nil {
		return "", "", fmt.Errorf("render subject for %s/%s: %w", eventType, lang, err)
	}

	if err := tmpl.ExecuteTemplate(&bodyBuf, "body", data); err != nil {
		return "", "", fmt.Errorf("render body for %s/%s: %w", eventType, lang, err)
	}

	return strings.TrimSpace(subjectBuf.String()), strings.TrimSpace(bodyBuf.String()), nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/template/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/template/ && git commit -m "feat: add email template renderer with language fallback"
```

---

### Task 6: All Email Templates (5 events x 3 languages + register_otp)

**Files:**
- Create: 18 HTML template files across 6 event directories

Each template uses inline CSS, single-column layout with brand header and footer.

- [ ] **Step 1: Create deposit_confirmed templates**

Create `server/internal/modules/notification/template/emails/deposit_confirmed/zh.html`:

```html
{{define "subject"}}您的充值 {{.Amount}} {{.Currency}} 已到账{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">充值已确认</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">您的 <strong>{{.Amount}} {{.Currency}}</strong> 已成功到账。</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">网络</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">交易哈希</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Update `server/internal/modules/notification/template/emails/deposit_confirmed/en.html` (replace the minimal test version):

```html
{{define "subject"}}Your deposit of {{.Amount}} {{.Currency}} has been confirmed{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">Deposit Confirmed</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Your deposit of <strong>{{.Amount}} {{.Currency}}</strong> has arrived in your account.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">Network</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">TxHash</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/deposit_confirmed/ko.html`:

```html
{{define "subject"}}{{.Amount}} {{.Currency}} 입금이 확인되었습니다{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">입금 확인</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;"><strong>{{.Amount}} {{.Currency}}</strong>이(가) 계좌에 입금되었습니다.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">네트워크</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">TxHash</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 2: Create withdraw_completed templates**

Create `server/internal/modules/notification/template/emails/withdraw_completed/en.html`:

```html
{{define "subject"}}Your withdrawal of {{.Amount}} {{.Currency}} is complete{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">Withdrawal Complete</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Your withdrawal of <strong>{{.Amount}} {{.Currency}}</strong> has been processed.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">Network</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">To Address</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;border-bottom:1px solid #eee;">{{.ToAddress}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">TxHash</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/withdraw_completed/zh.html`:

```html
{{define "subject"}}您的提现 {{.Amount}} {{.Currency}} 已完成{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">提现完成</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">您的 <strong>{{.Amount}} {{.Currency}}</strong> 提现已处理完成。</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">网络</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">目标地址</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;border-bottom:1px solid #eee;">{{.ToAddress}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">交易哈希</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/withdraw_completed/ko.html`:

```html
{{define "subject"}}{{.Amount}} {{.Currency}} 출금이 완료되었습니다{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">출금 완료</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;"><strong>{{.Amount}} {{.Currency}}</strong> 출금이 처리되었습니다.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">네트워크</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.Network}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">수신 주소</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;border-bottom:1px solid #eee;">{{.ToAddress}}</td></tr>
      <tr><td style="padding:8px 0;color:#666;">TxHash</td><td style="padding:8px 0;text-align:right;word-break:break-all;font-size:13px;">{{.TxHash}}</td></tr>
    </table>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 3: Create withdraw_failed templates**

Create `server/internal/modules/notification/template/emails/withdraw_failed/en.html`:

```html
{{define "subject"}}Your withdrawal of {{.Amount}} {{.Currency}} has failed{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#c0392b;margin:0 0 16px;">Withdrawal Failed</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Your withdrawal of <strong>{{.Amount}} {{.Currency}}</strong> could not be processed.</p>
    <p style="color:#333;font-size:16px;line-height:1.6;">Reason: {{.Reason}}</p>
    <p style="color:#333;font-size:14px;line-height:1.6;">The funds have been returned to your available balance. Please try again or contact support.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/withdraw_failed/zh.html`:

```html
{{define "subject"}}您的提现 {{.Amount}} {{.Currency}} 失败{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#c0392b;margin:0 0 16px;">提现失败</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">您的 <strong>{{.Amount}} {{.Currency}}</strong> 提现未能处理。</p>
    <p style="color:#333;font-size:16px;line-height:1.6;">原因：{{.Reason}}</p>
    <p style="color:#333;font-size:14px;line-height:1.6;">资金已退回您的可用余额，请重试或联系客服。</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/withdraw_failed/ko.html`:

```html
{{define "subject"}}{{.Amount}} {{.Currency}} 출금이 실패했습니다{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#c0392b;margin:0 0 16px;">출금 실패</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;"><strong>{{.Amount}} {{.Currency}}</strong> 출금을 처리할 수 없습니다.</p>
    <p style="color:#333;font-size:16px;line-height:1.6;">사유: {{.Reason}}</p>
    <p style="color:#333;font-size:14px;line-height:1.6;">자금이 가용 잔액으로 반환되었습니다. 다시 시도하거나 고객지원에 문의해 주세요.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 4: Create settlement_created templates**

Create `server/internal/modules/notification/template/emails/settlement_created/en.html`:

```html
{{define "subject"}}Daily Settlement Report — {{.Date}}{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">Daily Settlement</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Here is your settlement summary for <strong>{{.Date}}</strong>.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">Your Share</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;font-weight:bold;">{{.UserShare}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">Total PnL</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.TotalPnL}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;">Fee Rate</td><td style="padding:8px 0;text-align:right;">{{.FeeRate}}</td></tr>
    </table>
    <p style="color:#666;font-size:14px;line-height:1.6;">Earnings have been added to your earnings account.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/settlement_created/zh.html`:

```html
{{define "subject"}}每日结算报告 — {{.Date}}{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">每日结算</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">以下是 <strong>{{.Date}}</strong> 的结算摘要。</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">您的份额</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;font-weight:bold;">{{.UserShare}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">总损益</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.TotalPnL}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;">费率</td><td style="padding:8px 0;text-align:right;">{{.FeeRate}}</td></tr>
    </table>
    <p style="color:#666;font-size:14px;line-height:1.6;">收益已添加到您的收益账户。</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/settlement_created/ko.html`:

```html
{{define "subject"}}일일 정산 보고서 — {{.Date}}{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">일일 정산</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;"><strong>{{.Date}}</strong> 정산 요약입니다.</p>
    <table style="width:100%;border-collapse:collapse;margin:16px 0;">
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">회원 배분</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;font-weight:bold;">{{.UserShare}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;border-bottom:1px solid #eee;">총 손익</td><td style="padding:8px 0;text-align:right;border-bottom:1px solid #eee;">{{.TotalPnL}} USDT</td></tr>
      <tr><td style="padding:8px 0;color:#666;">수수료율</td><td style="padding:8px 0;text-align:right;">{{.FeeRate}}</td></tr>
    </table>
    <p style="color:#666;font-size:14px;line-height:1.6;">수익이 수익 계좌에 추가되었습니다.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 5: Create password_reset templates**

Create `server/internal/modules/notification/template/emails/password_reset/en.html`:

```html
{{define "subject"}}Your password reset code{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">Password Reset</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Use the following code to reset your password:</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">This code expires in {{.ExpiresIn}}. If you did not request this, please ignore this email.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/password_reset/zh.html`:

```html
{{define "subject"}}您的密码重置验证码{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">密码重置</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">请使用以下验证码重置您的密码：</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">验证码将在 {{.ExpiresIn}} 后过期。如果您没有请求此操作，请忽略此邮件。</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/password_reset/ko.html`:

```html
{{define "subject"}}비밀번호 재설정 코드{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">비밀번호 재설정</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">다음 코드를 사용하여 비밀번호를 재설정하세요:</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">이 코드는 {{.ExpiresIn}} 후에 만료됩니다. 요청하지 않은 경우 이 이메일을 무시하세요.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 6: Create register_otp templates**

Create `server/internal/modules/notification/template/emails/register_otp/en.html`:

```html
{{define "subject"}}Your registration verification code{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">Verify Your Email</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">Use the following code to complete your registration:</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">This code expires in {{.ExpiresIn}}. If you did not request this, please ignore this email.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">This is an automated message. Please do not reply.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/register_otp/zh.html`:

```html
{{define "subject"}}您的注册验证码{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">验证您的邮箱</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">请使用以下验证码完成注册：</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">验证码将在 {{.ExpiresIn}} 后过期。如果您没有请求此操作，请忽略此邮件。</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

Create `server/internal/modules/notification/template/emails/register_otp/ko.html`:

```html
{{define "subject"}}회원가입 인증 코드{{end}}
{{define "body"}}
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;background:#ffffff;">
  <tr><td style="background-color:#1a1a2e;padding:24px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">Sovereign Fund</h1>
  </td></tr>
  <tr><td style="padding:32px 24px;">
    <h2 style="color:#1a1a2e;margin:0 0 16px;">이메일 인증</h2>
    <p style="color:#333;font-size:16px;line-height:1.6;">다음 코드를 사용하여 회원가입을 완료하세요:</p>
    <div style="background-color:#f0f0f0;padding:16px;text-align:center;margin:16px 0;border-radius:8px;">
      <span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#1a1a2e;">{{.OTPCode}}</span>
    </div>
    <p style="color:#999;font-size:14px;">이 코드는 {{.ExpiresIn}} 후에 만료됩니다. 요청하지 않은 경우 이 이메일을 무시하세요.</p>
  </td></tr>
  <tr><td style="padding:16px 24px;background-color:#f9f9f9;text-align:center;color:#999;font-size:12px;">
    <p style="margin:0;">이 메일은 자동으로 발송된 메일입니다. 회신하지 마십시오.</p>
  </td></tr>
</table>
</body>
</html>
{{end}}
```

- [ ] **Step 7: Run renderer tests to verify templates load**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/template/ -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/template/ && git commit -m "feat: add email templates for all notification events (en/zh/ko)"
```

---

### Task 7: NotificationService

**Files:**
- Create: `server/internal/modules/notification/service/notification_service.go`
- Create: `server/internal/modules/notification/service/notification_service_test.go`

- [ ] **Step 1: Write service tests**

Create `server/internal/modules/notification/service/notification_service_test.go`:

```go
package service

import (
	"context"
	"testing"

	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	settingsmodel "github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
)

// stubUserRepo implements the UserFinder interface for testing.
type stubUserRepo struct {
	user *authmodel.User
	err  error
}

func (s *stubUserRepo) FindByID(_ context.Context, _ string) (*authmodel.User, error) {
	return s.user, s.err
}

// stubSettingsRepo implements the PrefFinder interface for testing.
type stubSettingsRepo struct {
	pref *settingsmodel.NotificationPref
	err  error
}

func (s *stubSettingsRepo) FindNotificationPref(_ context.Context, _ string) (*settingsmodel.NotificationPref, error) {
	return s.pref, s.err
}

func newTestService(t *testing.T, user *authmodel.User, pref *settingsmodel.NotificationPref) (*notificationService, *provider.MockProvider) {
	t.Helper()
	mock := &provider.MockProvider{}

	svc := &notificationService{
		emailProvider: mock,
		userRepo:      &stubUserRepo{user: user},
		settingsRepo:  &stubSettingsRepo{pref: pref},
		renderer:      nil, // set below
	}
	return svc, mock
}

func TestHandleDepositConfirmedSendsEmail(t *testing.T) {
	user := &authmodel.User{ID: "u1", Email: "test@example.com", Language: "en"}
	pref := &settingsmodel.NotificationPref{UserID: "u1", EmailDeposit: true}
	svc, mock := newTestService(t, user, pref)

	// Use a real renderer with test templates
	r, err := newTestRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	svc.renderer = r

	event := events.Event{
		Type:    events.DepositConfirmed,
		Payload: map[string]string{"user_id": "u1", "currency": "USDT", "amount": "100.00"},
	}

	err = svc.HandleDepositConfirmed(context.Background(), event)
	if err != nil {
		t.Fatalf("HandleDepositConfirmed: %v", err)
	}

	if len(mock.Sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mock.Sent))
	}
	if mock.Sent[0].To != "test@example.com" {
		t.Errorf("to = %q, want %q", mock.Sent[0].To, "test@example.com")
	}
}

func TestHandleDepositConfirmedSkipsWhenDisabled(t *testing.T) {
	user := &authmodel.User{ID: "u1", Email: "test@example.com", Language: "en"}
	pref := &settingsmodel.NotificationPref{UserID: "u1", EmailDeposit: false}
	svc, mock := newTestService(t, user, pref)

	event := events.Event{
		Type:    events.DepositConfirmed,
		Payload: map[string]string{"user_id": "u1", "currency": "USDT", "amount": "100.00"},
	}

	err := svc.HandleDepositConfirmed(context.Background(), event)
	if err != nil {
		t.Fatalf("HandleDepositConfirmed: %v", err)
	}

	if len(mock.Sent) != 0 {
		t.Errorf("expected 0 emails sent (disabled), got %d", len(mock.Sent))
	}
}

func TestHandleDepositConfirmedSkipsEmptyEmail(t *testing.T) {
	user := &authmodel.User{ID: "u1", Email: "", Language: "en"}
	pref := &settingsmodel.NotificationPref{UserID: "u1", EmailDeposit: true}
	svc, mock := newTestService(t, user, pref)

	event := events.Event{
		Type:    events.DepositConfirmed,
		Payload: map[string]string{"user_id": "u1", "currency": "USDT", "amount": "100.00"},
	}

	err := svc.HandleDepositConfirmed(context.Background(), event)
	if err != nil {
		t.Fatalf("HandleDepositConfirmed: %v", err)
	}

	if len(mock.Sent) != 0 {
		t.Errorf("expected 0 emails sent (empty email), got %d", len(mock.Sent))
	}
}

// newTestRenderer creates a renderer pointing to the actual template directory
// relative to this test file. Adjust if tests are run from a different cwd.
func newTestRenderer() (renderer, error) {
	// Walk up from service/ to notification/ then into template/emails
	return newRendererFromDir("../template/emails")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/service/ -v`
Expected: FAIL — service package does not exist yet

- [ ] **Step 3: Implement NotificationService**

Create `server/internal/modules/notification/service/notification_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"log/slog"

	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	authrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	tmpl "github.com/sovereign-fund/sovereign/internal/modules/notification/template"
	settingsmodel "github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	settingsrepo "github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

// UserFinder is the subset of UserRepository needed by this service.
type UserFinder interface {
	FindByID(ctx context.Context, id string) (*authmodel.User, error)
}

// PrefFinder is the subset of SettingsRepository needed by this service.
type PrefFinder interface {
	FindNotificationPref(ctx context.Context, userID string) (*settingsmodel.NotificationPref, error)
}

// renderer is the subset of template.Renderer used by this service.
type renderer interface {
	Render(eventType, lang string, data any) (subject string, html string, err error)
}

// newRendererFromDir wraps tmpl.NewRenderer for testability.
var newRendererFromDir = func(dir string) (renderer, error) {
	return tmpl.NewRenderer(dir)
}

type NotificationService interface {
	HandleDepositConfirmed(ctx context.Context, event events.Event) error
	HandleWithdrawCompleted(ctx context.Context, event events.Event) error
	HandleWithdrawFailed(ctx context.Context, event events.Event) error
	HandleSettlementCreated(ctx context.Context, event events.Event) error
	HandlePasswordReset(ctx context.Context, event events.Event) error
	SendOTP(ctx context.Context, email, lang, templateName, code, expiresIn string) error
}

type notificationService struct {
	emailProvider provider.EmailProvider
	userRepo      UserFinder
	settingsRepo  PrefFinder
	renderer      renderer
	logger        *slog.Logger
}

func NewNotificationService(
	emailProvider provider.EmailProvider,
	userRepo authrepo.UserRepository,
	settingsRepo settingsrepo.SettingsRepository,
	templateDir string,
	logger *slog.Logger,
) (NotificationService, error) {
	r, err := newRendererFromDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("init template renderer: %w", err)
	}

	return &notificationService{
		emailProvider: emailProvider,
		userRepo:      userRepo,
		settingsRepo:  settingsRepo,
		renderer:      r,
		logger:        logger,
	}, nil
}

func (s *notificationService) HandleDepositConfirmed(ctx context.Context, event events.Event) error {
	payload := extractPayload(event)
	if payload == nil {
		return nil
	}
	return s.sendIfEnabled(ctx, payload["user_id"], "deposit_confirmed", func(p *settingsmodel.NotificationPref) bool {
		return p.EmailDeposit
	}, map[string]string{
		"Amount":   payload["amount"],
		"Currency": payload["currency"],
		"Network":  payload["network"],
		"TxHash":   payload["tx_hash"],
	})
}

func (s *notificationService) HandleWithdrawCompleted(ctx context.Context, event events.Event) error {
	payload := extractPayload(event)
	if payload == nil {
		return nil
	}
	return s.sendIfEnabled(ctx, payload["user_id"], "withdraw_completed", func(p *settingsmodel.NotificationPref) bool {
		return p.EmailWithdraw
	}, map[string]string{
		"Amount":    payload["amount"],
		"Currency":  payload["currency"],
		"Network":   payload["network"],
		"TxHash":    payload["tx_hash"],
		"ToAddress": payload["to_address"],
	})
}

func (s *notificationService) HandleWithdrawFailed(ctx context.Context, event events.Event) error {
	payload := extractPayload(event)
	if payload == nil {
		return nil
	}
	return s.sendIfEnabled(ctx, payload["user_id"], "withdraw_failed", func(p *settingsmodel.NotificationPref) bool {
		return p.EmailWithdraw
	}, map[string]string{
		"Amount":   payload["amount"],
		"Currency": payload["currency"],
		"Reason":   payload["reason"],
	})
}

func (s *notificationService) HandleSettlementCreated(ctx context.Context, event events.Event) error {
	payload := extractPayload(event)
	if payload == nil {
		return nil
	}
	return s.sendIfEnabled(ctx, payload["user_id"], "settlement_created", func(p *settingsmodel.NotificationPref) bool {
		return p.EmailSettlement
	}, map[string]string{
		"Date":      payload["period"],
		"TotalPnL":  payload["total_pnl"],
		"UserShare": payload["net_return"],
		"FeeRate":   payload["fee_rate"],
	})
}

func (s *notificationService) HandlePasswordReset(ctx context.Context, event events.Event) error {
	payload := extractPayload(event)
	if payload == nil {
		return nil
	}
	return s.sendAlways(ctx, payload["email"], payload["lang"], "password_reset", map[string]string{
		"OTPCode":   payload["otp_code"],
		"ExpiresIn": payload["expires_in"],
	})
}

func (s *notificationService) SendOTP(ctx context.Context, email, lang, templateName, code, expiresIn string) error {
	return s.sendAlways(ctx, email, lang, templateName, map[string]string{
		"OTPCode":   code,
		"ExpiresIn": expiresIn,
	})
}

func (s *notificationService) sendIfEnabled(
	ctx context.Context,
	userID, eventType string,
	isEnabled func(*settingsmodel.NotificationPref) bool,
	data map[string]string,
) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("notification: find user failed", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil
	}

	if user.Email == "" {
		s.logger.Warn("notification: user has no email", slog.String("user_id", userID))
		return nil
	}

	pref, err := s.settingsRepo.FindNotificationPref(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Default prefs have email enabled
			pref = &settingsmodel.NotificationPref{
				EmailTrade:      true,
				EmailDeposit:    true,
				EmailWithdraw:   true,
				EmailSettlement: true,
			}
		} else {
			s.logger.Error("notification: find pref failed", slog.String("user_id", userID), slog.String("error", err.Error()))
			return nil
		}
	}

	if !isEnabled(pref) {
		return nil
	}

	return s.sendAlways(ctx, user.Email, user.Language, eventType, data)
}

func (s *notificationService) sendAlways(ctx context.Context, email, lang, eventType string, data map[string]string) error {
	if email == "" {
		s.logger.Warn("notification: empty email, skipping", slog.String("event", eventType))
		return nil
	}

	subject, html, err := s.renderer.Render(eventType, lang, data)
	if err != nil {
		s.logger.Error("notification: render template failed",
			slog.String("event", eventType),
			slog.String("lang", lang),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if err := s.emailProvider.Send(ctx, provider.SendInput{
		To:      email,
		Subject: subject,
		HTML:    html,
	}); err != nil {
		s.logger.Error("notification: send email failed",
			slog.String("event", eventType),
			slog.String("to", email),
			slog.String("error", err.Error()),
		)
	}

	return nil
}

func extractPayload(event events.Event) map[string]string {
	payload, ok := event.Payload.(map[string]string)
	if !ok {
		return nil
	}
	return payload
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/notification/service/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/service/ && git commit -m "feat: add NotificationService with event handlers and preference checks"
```

---

### Task 8: Notification Module Init

**Files:**
- Create: `server/internal/modules/notification/module.go`

- [ ] **Step 1: Create module.go**

Create `server/internal/modules/notification/module.go`:

```go
package notification

import (
	"context"
	"fmt"
	"log/slog"

	authrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/service"
	settingsrepo "github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	"github.com/sovereign-fund/sovereign/config"
)

type Module struct {
	Service service.NotificationService
}

func NewModule(
	cfg config.NotificationConfig,
	userRepo authrepo.UserRepository,
	settingsRepo settingsrepo.SettingsRepository,
	templateDir string,
	logger *slog.Logger,
) (*Module, error) {
	var emailProvider provider.EmailProvider

	if cfg.UseMock {
		emailProvider = provider.NewMockProvider()
		logger.Info("notification: using mock email provider")
	} else {
		var err error
		emailProvider, err = provider.NewSESProvider(context.Background(), cfg.AWSRegion, cfg.FromName, cfg.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("init SES provider: %w", err)
		}
		logger.Info("notification: using AWS SES provider",
			slog.String("region", cfg.AWSRegion),
			slog.String("from", cfg.FromAddress),
		)
	}

	svc, err := service.NewNotificationService(emailProvider, userRepo, settingsRepo, templateDir, logger)
	if err != nil {
		return nil, fmt.Errorf("init notification service: %w", err)
	}

	return &Module{Service: svc}, nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./internal/modules/notification/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/notification/module.go && git commit -m "feat: add notification module init with provider selection"
```

---

### Task 9: Wire Module into App + Register Event Subscriptions

**Files:**
- Modify: `server/internal/app/app.go`

- [ ] **Step 1: Add notification module to App struct and initialization**

In `server/internal/app/app.go`:

Add import:
```go
"github.com/sovereign-fund/sovereign/internal/modules/notification"
```

Add field to `App` struct:
```go
NotificationModule *notification.Module
```

In the `New` function, after the `SettingsModule` initialization and before the `return &App{` block, add:

```go
	// Notification module
	notifMod, err := notification.NewModule(
		cfg.Notification,
		authMod.UserRepo(),
		settingsMod.Repo(),
		"internal/modules/notification/template/emails",
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("init notification module: %w", err)
	}

	// Subscribe notification handlers to events
	bus.Subscribe(events.DepositConfirmed, notifMod.Service.HandleDepositConfirmed)
	bus.Subscribe(events.WithdrawCompleted, notifMod.Service.HandleWithdrawCompleted)
	bus.Subscribe(events.WithdrawFailed, notifMod.Service.HandleWithdrawFailed)
	bus.Subscribe(events.SettlementCreated, notifMod.Service.HandleSettlementCreated)
```

Add `NotificationModule: notifMod,` to the returned `App` struct literal.

**Note:** This step requires that `auth.Module` exposes `UserRepo()` and `settings.Module` exposes `Repo()` methods. If these don't exist, add them:

In `server/internal/modules/auth/module.go`, add:
```go
func (m *Module) UserRepo() repository.UserRepository {
	return m.userRepo
}
```
(Store userRepo as field during NewModule if not already done.)

In `server/internal/modules/settings/module.go`, add:
```go
func (m *Module) Repo() repository.SettingsRepository {
	return m.settingsRepo
}
```
(Store settingsRepo as field during NewModule if not already done.)

- [ ] **Step 2: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/app/app.go internal/modules/auth/module.go internal/modules/settings/module.go && git commit -m "feat: wire notification module into app with event subscriptions"
```

---

### Task 10: Replace Auth OTP Logger with Email Sending

**Files:**
- Modify: `server/internal/modules/auth/service/auth_service.go`
- Modify: `server/internal/modules/auth/module.go`

- [ ] **Step 1: Add NotificationService dependency to authService**

In `server/internal/modules/auth/service/auth_service.go`:

Add to `AuthService` constructor parameters and `authService` struct:
```go
notifSvc notification.NotificationService
```

Import the notification service package:
```go
notifSvc "github.com/sovereign-fund/sovereign/internal/modules/notification/service"
```

- [ ] **Step 2: Replace register OTP log with email send**

In `SendRegisterOTP` (around line 84-93), replace the `// TODO: replace with email service` block:

```go
	// TODO: replace with email service
	s.logger.Info("========== REGISTER OTP ==========",
		slog.String("email", req.Email),
		slog.String("code", otp),
	)
```

With:

```go
	if err := s.notifSvc.SendOTP(ctx, req.Email, "", "register_otp", otp, "5 minutes"); err != nil {
		s.logger.Error("failed to send register OTP email", slog.String("email", req.Email), slog.String("error", err.Error()))
	}
```

- [ ] **Step 3: Replace password reset OTP log with email send**

In `ForgotPassword` (around line 297-301), replace the `// TODO: replace with email service` block:

```go
	// TODO: replace with email service
	s.logger.Info("========== RESET PASSWORD OTP ==========",
		slog.String("email", email),
		slog.String("code", otp),
	)
```

With:

```go
	user, _ := s.userRepo.FindByEmail(ctx, email)
	lang := "en"
	if user != nil {
		lang = user.Language
	}
	if err := s.notifSvc.SendOTP(ctx, email, lang, "password_reset", otp, "5 minutes"); err != nil {
		s.logger.Error("failed to send password reset email", slog.String("email", email), slog.String("error", err.Error()))
	}
```

- [ ] **Step 4: Update auth module to pass NotificationService**

In `server/internal/modules/auth/module.go`, update `NewModule` to accept and pass `NotificationService` to the auth service constructor.

- [ ] **Step 5: Update app.go to pass notification service to auth module**

In `server/internal/app/app.go`, ensure the auth module is initialized after the notification module, and pass `notifMod.Service` to `auth.NewModule`.

**Important:** This may require reordering the module initialization in `app.go`. The notification module must be created before the auth module since auth now depends on it.

- [ ] **Step 6: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 7: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/auth/ internal/app/app.go && git commit -m "feat: replace OTP logger with email sending via notification service"
```

---

### Task 11: Enrich Event Payloads for Email Templates

The existing event payloads are missing some fields needed by the email templates (e.g., `DepositConfirmed` lacks `network` and `tx_hash`, `WithdrawCompleted` lacks `amount`, `currency`, `network`, `to_address`, `tx_hash`).

**Files:**
- Modify: `server/internal/modules/wallet/service/wallet_service.go`

- [ ] **Step 1: Enrich DepositConfirmed payload**

At line ~596 in `wallet_service.go`, the `DepositConfirmed` event currently sends:
```go
Payload: map[string]string{
    "user_id":  userID,
    "currency": payload.Currency,
    "amount":   payload.Amount.String(),
},
```

Add network and tx_hash:
```go
Payload: map[string]string{
    "user_id":  userID,
    "currency": payload.Currency,
    "amount":   payload.Amount.String(),
    "network":  payload.Network,
    "tx_hash":  payload.TxHash,
},
```

- [ ] **Step 2: Enrich WithdrawCompleted payload**

At line ~468, the `WithdrawCompleted` event sends only `user_id` and `transaction_id`. Add the fields needed by the email template. This requires reading the transaction from the database in context — the `tx` variable is already available at that point:

```go
Payload: map[string]string{
    "user_id":        tx.UserID,
    "transaction_id": tx.ID,
    "amount":         tx.Amount.String(),
    "currency":       tx.Currency,
    "network":        tx.Network,
    "to_address":     tx.Address,
    "tx_hash":        tx.TxHash,
},
```

- [ ] **Step 3: Enrich WithdrawFailed payload**

At line ~488, similarly enrich:

```go
Payload: map[string]string{
    "user_id":        tx.UserID,
    "transaction_id": tx.ID,
    "amount":         tx.Amount.String(),
    "currency":       tx.Currency,
    "reason":         "Transaction failed on chain",
},
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/wallet/service/wallet_service.go && git commit -m "feat: enrich event payloads with fields needed for email templates"
```

---

### Task 12: Full Build Verification + Run All Tests

- [ ] **Step 1: Build entire project**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 2: Run all tests**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./... -v -count=1`
Expected: All tests PASS

- [ ] **Step 3: Run vet**

Run: `cd /Users/johnny/Work/soveregin/server && go vet ./...`
Expected: No issues

- [ ] **Step 4: Commit any remaining fixes**

If any fixes were needed, commit them:

```bash
cd /Users/johnny/Work/soveregin/server && git add -A && git commit -m "fix: resolve build/test issues in notification module"
```
