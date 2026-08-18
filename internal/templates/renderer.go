package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Locale represents the target language for notifications.
type Locale string

const (
	LocaleRU Locale = "ru"
	LocaleEN Locale = "en"
)

// Base layout and styling for all notification emails.
const emailBaseLayout = `<!DOCTYPE html>
<html lang="{{ .Lang }}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }}</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      background-color: #f4f6f8;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      color: #1e293b;
      line-height: 1.6;
    }
    .wrapper {
      width: 100%;
      table-layout: fixed;
      background-color: #f4f6f8;
      padding: 40px 0;
    }
    .container {
      max-width: 580px;
      margin: 0 auto;
      background: #ffffff;
      border-radius: 8px;
      overflow: hidden;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
      border: 1px solid #e2e8f0;
    }
    .header {
      background: #1e3a8a;
      padding: 24px 32px;
      text-align: left;
    }
    .header h1 {
      margin: 0;
      color: #ffffff;
      font-size: 20px;
      font-weight: 600;
      letter-spacing: -0.02em;
    }
    .content {
      padding: 32px;
    }
    .title {
      font-size: 18px;
      font-weight: 600;
      color: #0f172a;
      margin-top: 0;
      margin-bottom: 16px;
    }
    .badge {
      display: inline-block;
      padding: 4px 10px;
      border-radius: 4px;
      font-size: 13px;
      font-weight: 500;
      background-color: #e0f2fe;
      color: #0369a1;
      margin-bottom: 16px;
    }
    .info-table {
      width: 100%;
      margin: 20px 0;
      border-collapse: collapse;
    }
    .info-table td {
      padding: 10px 12px;
      border-bottom: 1px solid #f1f5f9;
      font-size: 14px;
    }
    .info-table td.label {
      color: #64748b;
      width: 35%;
      font-weight: 500;
    }
    .info-table td.value {
      color: #0f172a;
      font-weight: 600;
    }
    .button-container {
      margin: 28px 0 12px 0;
      text-align: center;
    }
    .btn {
      display: inline-block;
      padding: 12px 28px;
      background-color: #2563eb;
      color: #ffffff !important;
      text-decoration: none;
      border-radius: 6px;
      font-size: 14px;
      font-weight: 600;
    }
    .footer {
      padding: 20px 32px;
      background: #f8fafc;
      border-top: 1px solid #e2e8f0;
      font-size: 12px;
      color: #94a3b8;
      text-align: center;
    }
  </style>
</head>
<body>
  <table class="wrapper" role="presentation">
    <tr>
      <td align="center">
        <div class="container">
          <div class="header">
            <h1>Appraisal CRM · {{ .BrandSubtitle }}</h1>
          </div>
          <div class="content">
            <h2 class="title">{{ .Title }}</h2>
            {{ if .Badge }}<div class="badge">{{ .Badge }}</div>{{ end }}
            <p>{{ .Message }}</p>
            {{ if .Details }}
            <table class="info-table">
              {{ range .Details }}
              <tr>
                <td class="label">{{ .Label }}</td>
                <td class="value">{{ .Value }}</td>
              </tr>
              {{ end }}
            </table>
            {{ end }}
            {{ if .ActionURL }}
            <div class="button-container">
              <a href="{{ .ActionURL }}" class="btn" target="_blank">{{ .ActionText }}</a>
            </div>
            {{ end }}
          </div>
          <div class="footer">
            {{ .FooterText }}
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

type DetailItem struct {
	Label string
	Value string
}

type emailData struct {
	Lang          string
	BrandSubtitle string
	Title         string
	Badge         string
	Message       string
	Details       []DetailItem
	ActionURL     string
	ActionText    string
	FooterText    string
}

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	tmpl, err := template.New("email").Parse(emailBaseLayout)
	if err != nil {
		return nil, fmt.Errorf("parse base email template: %w", err)
	}
	return &Renderer{tmpl: tmpl}, nil
}

func (r *Renderer) render(data emailData) (string, error) {
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func normalizeLocale(locale Locale) Locale {
	if locale == LocaleEN {
		return LocaleEN
	}
	return LocaleRU
}

// 1. Request Created -> Client
func (r *Renderer) RenderRequestCreatedClient(locale Locale, reqID, email, phone string, objType, address *string, clientPortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Your appraisal request is received (№ %s)", shortID)
		details := []DetailItem{
			{Label: "Request ID", Value: reqID},
		}
		if objType != nil && *objType != "" {
			details = append(details, DetailItem{Label: "Property Type", Value: formatObjectTypeEN(*objType)})
		}
		if address != nil && *address != "" {
			details = append(details, DetailItem{Label: "Address", Value: *address})
		}

		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         "Appraisal Request Successfully Registered",
			Badge:         "Status: New",
			Message:       "Thank you for your request. Our team has received your order and will contact you shortly to coordinate details.",
			Details:       details,
			ActionURL:     fmt.Sprintf("%s/", clientPortalURL),
			ActionText:    "Open Client Portal",
			FooterText:    "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Ваша заявка на оценку принята (№ %s)", shortID)
	details := []DetailItem{
		{Label: "Номер заявки", Value: reqID},
	}
	if objType != nil && *objType != "" {
		details = append(details, DetailItem{Label: "Тип объекта", Value: formatObjectTypeRU(*objType)})
	}
	if address != nil && *address != "" {
		details = append(details, DetailItem{Label: "Адрес", Value: *address})
	}

	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Заявка на оценку успешно зарегистрирована",
		Badge:         "Статус: Новая",
		Message:       "Спасибо за обращение! Ваша заявка принята в работу. Наш специалист свяжется с вами в ближайшее время для согласования деталей.",
		Details:       details,
		ActionURL:     fmt.Sprintf("%s/", clientPortalURL),
		ActionText:    "Перейти в личный кабинет",
		FooterText:    "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 2. Request Created -> Staff / Appraiser
func (r *Renderer) RenderRequestCreatedStaff(locale Locale, reqID, email, phone string, objType, address *string, officePortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("New appraisal request (№ %s)", shortID)
		details := []DetailItem{
			{Label: "Request ID", Value: reqID},
			{Label: "Client Email", Value: email},
			{Label: "Client Phone", Value: phone},
		}
		if objType != nil && *objType != "" {
			details = append(details, DetailItem{Label: "Property Type", Value: formatObjectTypeEN(*objType)})
		}
		if address != nil && *address != "" {
			details = append(details, DetailItem{Label: "Address", Value: *address})
		}

		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Back Office",
			Title:         "New Appraisal Request Received",
			Badge:         "Action Required",
			Message:       "A client submitted a new valuation request via the client portal. Please review and assign an inspector.",
			Details:       details,
			ActionURL:     fmt.Sprintf("%s/requests", officePortalURL),
			ActionText:    "View Requests List",
			FooterText:    "This is an internal system notification. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Новая заявка на оценку (№ %s)", shortID)
	details := []DetailItem{
		{Label: "Номер заявки", Value: reqID},
		{Label: "Email клиента", Value: email},
		{Label: "Телефон клиента", Value: phone},
	}
	if objType != nil && *objType != "" {
		details = append(details, DetailItem{Label: "Тип объекта", Value: formatObjectTypeRU(*objType)})
	}
	if address != nil && *address != "" {
		details = append(details, DetailItem{Label: "Адрес", Value: *address})
	}

	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Поступила новая заявка на оценку",
		Badge:         "Требует обработки",
		Message:       "Клиент подал заявку через форму на портале. Необходимо взять заявку в работу и согласовать выезд осмотрщика.",
		Details:       details,
		ActionURL:     fmt.Sprintf("%s/requests", officePortalURL),
		ActionText:    "Открыть список заявок",
		FooterText:    "Это служебное уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 3. Inspection Scheduled -> Client
func (r *Renderer) RenderInspectionScheduledClient(locale Locale, reqID string, clientPortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Property inspection scheduled for request № %s", shortID)
		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         "Inspection Visit Scheduled",
			Badge:         "Status: Inspection Scheduled",
			Message:       "An inspector has been assigned to your request and will reach out to you shortly to agree on the exact time of the visit.",
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
			},
			ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
			ActionText: "Check Request Status",
			FooterText: "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Назначен осмотр объекта по заявке № %s", shortID)
	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Осмотр объекта назначен",
		Badge:         "Статус: Осмотр назначен",
		Message:       "По вашей заявке назначен выезд специалиста-осмотрщика. Осмотрщик свяжется с вами для согласования удобного времени проведения осмотра.",
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
		},
		ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
		ActionText: "Проверить статус заявки",
		FooterText: "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 4. Inspection Scheduled -> Inspector (Staff)
func (r *Renderer) RenderInspectionScheduledStaff(locale Locale, reqID string, officePortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("New inspection assignment (request № %s)", shortID)
		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Back Office",
			Title:         "Inspection Assignment",
			Badge:         "Scheduled",
			Message:       "You have been assigned to conduct an on-site property inspection. Please check the assignment card for property details and photo requirements.",
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
			},
			ActionURL:  fmt.Sprintf("%s/inspections", officePortalURL),
			ActionText: "Open Inspection Assignment",
			FooterText: "This is an internal system notification. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Вам назначено проведение осмотра (заявка № %s)", shortID)
	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Новое задание на осмотр",
		Badge:         "Назначен выезд",
		Message:       "Вам поручено проведение осмотра объекта. Пожалуйста, ознакомьтесь с деталями заявки и проведите фотофиксацию.",
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
		},
		ActionURL:  fmt.Sprintf("%s/inspections", officePortalURL),
		ActionText: "Открыть задание на осмотр",
		FooterText: "Это служебное уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 5. Inspection Completed -> Client / Appraiser
func (r *Renderer) RenderInspectionCompleted(locale Locale, reqID string, photoCount int, portalURL string, isClient bool) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Inspection for request № %s completed", shortID)
		title := "Inspection Visit Completed"
		badge := "Status: Inspection Completed"
		msg := "The property inspection has finished. Photos and field notes have been submitted to the appraiser."
		actionText := "Open Request"
		if isClient {
			msg = "The property inspection was successfully carried out. Our appraiser is now working on the valuation report."
			actionText = "Open Client Portal"
		}

		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         title,
			Badge:         badge,
			Message:       msg,
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
				{Label: "Uploaded Photos", Value: fmt.Sprintf("%d files", photoCount)},
			},
			ActionURL:  portalURL,
			ActionText: actionText,
			FooterText: "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Осмотр по заявке № %s успешно завершен", shortID)
	message := "Осмотр объекта завершен. Фотографии и материалы переданы оценщику."
	actionText := "Открыть заявку"
	if isClient {
		message = "Осмотр вашего объекта успешно проведен. Эксперт приступает к составлению отчета об оценке."
		actionText = "Перейти в личный кабинет"
	}

	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Осмотр завершен",
		Badge:         "Статус: Осмотр выполнен",
		Message:       message,
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
			{Label: "Загружено фото", Value: fmt.Sprintf("%d шт.", photoCount)},
		},
		ActionURL:  portalURL,
		ActionText: actionText,
		FooterText: "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 6. Appraisal Started -> Client
func (r *Renderer) RenderAppraisalStarted(locale Locale, reqID string, clientPortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Appraisal calculation in progress (request № %s)", shortID)
		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         "Valuation in Progress",
			Badge:         "Status: Appraisal",
			Message:       "The appraiser is conducting market comparable analysis and final valuation calculations. Your report will be ready soon.",
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
			},
			ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
			ActionText: "Track Request Status",
			FooterText: "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Идет подготовка отчета об оценке (заявка № %s)", shortID)
	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Оценка в процессе",
		Badge:         "Статус: Оценка",
		Message:       "Оценщик проводит анализ рынка и расчёт стоимости вашего объекта. Скоро отчет будет готов.",
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
		},
		ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
		ActionText: "Следить за статусом",
		FooterText: "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 7. Report Ready -> Client (BR-013)
func (r *Renderer) RenderReportReadyClient(locale Locale, reqID string, clientPortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Your appraisal report is ready! (request № %s)", shortID)
		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         "Appraisal Report is Ready for Download",
			Badge:         "Status: Report Sent",
			Message:       "Your official property valuation report has been approved and is available for download in PDF format.",
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
			},
			ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
			ActionText: "Download Report",
			FooterText: "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Отчет об оценке готов! (заявка № %s)", shortID)
	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Отчет об оценке готов к скачиванию",
		Badge:         "Статус: Отчет отправлен",
		Message:       "Итоговый отчет об оценке успешно сформирован и проверен. Вы можете скачать PDF-документ в личном кабинете.",
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
		},
		ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
		ActionText: "Скачать отчет в кабинете",
		FooterText: "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

// 8. Request Closed -> Client
func (r *Renderer) RenderRequestClosed(locale Locale, reqID string, clientPortalURL string) (string, string, error) {
	loc := normalizeLocale(locale)
	shortID := formatShortID(reqID)

	if loc == LocaleEN {
		subject := fmt.Sprintf("Request № %s is closed", shortID)
		body, err := r.render(emailData{
			Lang:          "en",
			BrandSubtitle: "Property Valuation",
			Title:         "Request Successfully Closed",
			Badge:         "Status: Closed",
			Message:       "All stages of your appraisal request are completed. Thank you for choosing our service!",
			Details: []DetailItem{
				{Label: "Request ID", Value: reqID},
			},
			ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
			ActionText: "Open Client Portal",
			FooterText: "This is an automated notification from Appraisal CRM. Please do not reply to this email.",
		})
		return subject, body, err
	}

	subject := fmt.Sprintf("Заявка № %s закрыта", shortID)
	body, err := r.render(emailData{
		Lang:          "ru",
		BrandSubtitle: "Авангард Оценка",
		Title:         "Заявка успешно закрыта",
		Badge:         "Статус: Закрыта",
		Message:       "Все этапы оценки по вашей заявке завершены. Благодарим за выбор нашей компании!",
		Details: []DetailItem{
			{Label: "Номер заявки", Value: reqID},
		},
		ActionURL:  fmt.Sprintf("%s/", clientPortalURL),
		ActionText: "Перейти в личный кабинет",
		FooterText: "Это автоматическое уведомление системы Appraisal CRM. Пожалуйста, не отвечайте на это письмо.",
	})
	return subject, body, err
}

func formatShortID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatObjectTypeRU(ot string) string {
	switch ot {
	case "apartment":
		return "Квартира"
	case "house":
		return "Жилой дом"
	case "land":
		return "Земельный участок"
	case "commercial":
		return "Коммерческая недвижимость"
	case "vehicle":
		return "Транспортное средство"
	default:
		return ot
	}
}

func formatObjectTypeEN(ot string) string {
	switch ot {
	case "apartment":
		return "Apartment"
	case "house":
		return "House"
	case "land":
		return "Land plot"
	case "commercial":
		return "Commercial real estate"
	case "vehicle":
		return "Vehicle"
	default:
		return ot
	}
}
