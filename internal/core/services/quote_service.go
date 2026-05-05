package services

import (
	"bytes"
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type QuoteService struct {
	quoteRepo   ports.QuoteRepository
	clientRepo  ports.ClientRepository
	companyRepo ports.CompanyRepository
}

func NewQuoteService(quoteRepo ports.QuoteRepository, clientRepo ports.ClientRepository, companyRepo ports.CompanyRepository) *QuoteService {
	return &QuoteService{quoteRepo: quoteRepo, clientRepo: clientRepo, companyRepo: companyRepo}
}

func (s *QuoteService) CreateQuote(companyID, userID, clientID, title string, clientSnapshot domain.ClientSnapshot, paymentTerms domain.PaymentTerms, items []domain.QuoteItem) (*domain.Quote, error) {
	now := time.Now()
	quote := &domain.Quote{
		ID:        uuid.NewString(),
		CompanyID: companyID,
		ClientID:  clientID,
		Title:     strings.TrimSpace(title),
		Status:    domain.QuoteStatusDraft,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.hydrateSnapshots(quote, clientSnapshot); err != nil {
		return nil, err
	}
	if err := applyQuoteItems(quote, items); err != nil {
		return nil, err
	}
	if err := applyPaymentTerms(quote, paymentTerms); err != nil {
		return nil, err
	}
	if err := s.quoteRepo.CreateQuote(quote); err != nil {
		return nil, err
	}
	return quote, nil
}

func (s *QuoteService) ListQuotes(companyID string) ([]domain.Quote, error) {
	return s.quoteRepo.GetAllQuotes(companyID)
}

func (s *QuoteService) GetQuote(id, companyID string) (*domain.Quote, error) {
	return s.quoteRepo.GetQuoteByID(id, companyID)
}

func (s *QuoteService) UpdateQuote(id, companyID, clientID, title string, clientSnapshot domain.ClientSnapshot, paymentTerms domain.PaymentTerms, items []domain.QuoteItem) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetQuoteByID(id, companyID)
	if err != nil {
		return nil, err
	}
	if quote.Status == domain.QuoteStatusPublished {
		return nil, errors.New("published quotes cannot be edited")
	}
	quote.ClientID = clientID
	quote.Title = strings.TrimSpace(title)
	quote.UpdatedAt = time.Now()
	if err := s.hydrateSnapshots(quote, clientSnapshot); err != nil {
		return nil, err
	}
	if err := applyQuoteItems(quote, items); err != nil {
		return nil, err
	}
	if err := applyPaymentTerms(quote, paymentTerms); err != nil {
		return nil, err
	}
	if err := s.quoteRepo.UpdateQuote(quote); err != nil {
		return nil, err
	}
	return quote, nil
}

func (s *QuoteService) PublishQuote(id, companyID string) (*domain.Quote, string, error) {
	quote, err := s.quoteRepo.GetQuoteByID(id, companyID)
	if err != nil {
		return nil, "", err
	}
	if len(quote.Items) == 0 {
		return nil, "", errors.New("quote must have at least one item")
	}

	password, err := randomSixDigitPassword()
	if err != nil {
		return nil, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	if quote.ShareToken == "" {
		token, err := randomToken()
		if err != nil {
			return nil, "", err
		}
		quote.ShareToken = token
	}
	quote.SharePasswordHash = string(hash)
	quote.Status = domain.QuoteStatusPublished
	quote.UpdatedAt = time.Now()
	if err := s.quoteRepo.UpdateQuote(quote); err != nil {
		return nil, "", err
	}
	return quote, password, nil
}

func (s *QuoteService) ArchiveQuote(id, companyID string) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetQuoteByID(id, companyID)
	if err != nil {
		return nil, err
	}
	quote.Status = domain.QuoteStatusArchived
	quote.UpdatedAt = time.Now()
	if err := s.quoteRepo.UpdateQuote(quote); err != nil {
		return nil, err
	}
	return quote, nil
}

func (s *QuoteService) GetPublicQuote(token string) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetQuoteByToken(token)
	if err != nil {
		return nil, err
	}
	if quote.Status != domain.QuoteStatusPublished {
		return nil, gorm.ErrRecordNotFound
	}
	return &domain.Quote{
		ID:         quote.ID,
		Title:      quote.Title,
		Status:     quote.Status,
		ShareToken: quote.ShareToken,
	}, nil
}

func (s *QuoteService) VerifyPublicQuotePassword(token, password string) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetQuoteByToken(token)
	if err != nil {
		return nil, err
	}
	if quote.Status != domain.QuoteStatusPublished {
		return nil, gorm.ErrRecordNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(quote.SharePasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return quote, nil
}

func (s *QuoteService) GenerateQuotePDF(id, companyID string) ([]byte, string, error) {
	quote, err := s.quoteRepo.GetQuoteByID(id, companyID)
	if err != nil {
		return nil, "", err
	}
	return buildQuotePDF(quote), fmt.Sprintf("orcamento-%s.pdf", quote.ID), nil
}

func (s *QuoteService) hydrateSnapshots(quote *domain.Quote, override domain.ClientSnapshot) error {
	company, err := s.companyRepo.GetCompanyByID(quote.CompanyID)
	if err != nil {
		return err
	}
	quote.ClientSnapshot = domain.ClientSnapshot{}
	if quote.ClientID != "" {
		client, err := s.clientRepo.GetClientByID(quote.ClientID, quote.CompanyID)
		if err != nil {
			return err
		}
		quote.ClientSnapshot = domain.ClientSnapshot{Name: client.Name, Phone: client.Phone}
	}
	if strings.TrimSpace(override.Name) != "" {
		quote.ClientSnapshot.Name = strings.TrimSpace(override.Name)
	}
	if strings.TrimSpace(override.Phone) != "" {
		quote.ClientSnapshot.Phone = strings.TrimSpace(override.Phone)
	}
	if strings.TrimSpace(override.Email) != "" {
		quote.ClientSnapshot.Email = strings.TrimSpace(override.Email)
	}
	if quote.ClientSnapshot.Name == "" {
		return errors.New("client name is required")
	}
	quote.CompanySnapshot = domain.CompanySnapshot{Name: company.Name, Email: company.Email, Phone: company.Phone, Address: company.Address, LogoURL: company.PublicAvatar}
	return nil
}

func applyQuoteItems(quote *domain.Quote, items []domain.QuoteItem) error {
	if quote.Title == "" {
		return errors.New("title is required")
	}
	if len(items) == 0 {
		return errors.New("at least one item is required")
	}
	now := time.Now()
	total := int64(0)
	quote.Items = make([]domain.QuoteItem, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return errors.New("item name is required")
		}
		if item.Quantity <= 0 {
			return errors.New("item quantity must be greater than zero")
		}
		if item.UnitPrice < 0 {
			return errors.New("item unit_price must be greater than or equal to zero")
		}
		unit, err := normalizeQuoteItemUnit(item.Unit)
		if err != nil {
			return err
		}
		if item.ID == "" {
			item.ID = uuid.NewString()
		}
		item.QuoteID = quote.ID
		item.Name = name
		item.Unit = unit
		item.Total = int64(math.Round(item.Quantity * float64(item.UnitPrice)))
		item.CreatedAt = now
		item.UpdatedAt = now
		total += item.Total
		quote.Items = append(quote.Items, item)
	}
	quote.Subtotal = total
	quote.Total = total
	return nil
}

func applyPaymentTerms(quote *domain.Quote, terms domain.PaymentTerms) error {
	if !terms.InstallmentsEnabled {
		quote.PaymentTerms = domain.PaymentTerms{CashAmount: quote.Total}
		return nil
	}
	if terms.InstallmentCount < 2 {
		return errors.New("installment_count must be greater than 1")
	}
	if terms.InstallmentCount > 36 {
		return errors.New("installment_count must be up to 36")
	}

	interestType := strings.TrimSpace(strings.ToLower(terms.InterestType))
	if interestType == "" {
		interestType = "none"
	}
	if interestType != "none" && interestType != "with_interest" {
		return errors.New("interest_type must be none or with_interest")
	}

	totalWithInterest := quote.Total
	interestRate := 0.0
	if interestType == "with_interest" {
		if terms.InterestRate < 0 {
			return errors.New("interest_rate must be greater than or equal to zero")
		}
		interestRate = terms.InterestRate
		totalWithInterest = int64(math.Round(float64(quote.Total) * (1 + interestRate/100)))
	}

	quote.PaymentTerms = domain.PaymentTerms{
		InstallmentsEnabled: true,
		InstallmentCount:    terms.InstallmentCount,
		InterestType:        interestType,
		InterestRate:        interestRate,
		InstallmentAmount:   int64(math.Ceil(float64(totalWithInterest) / float64(terms.InstallmentCount))),
		CashAmount:          quote.Total,
	}
	return nil
}

var allowedQuoteItemUnits = map[string]string{
	"kg": "Kg",
	"un": "UN",
	"pc": "PÇ",
	"pç": "PÇ",
	"m2": "m²",
	"m²": "m²",
	"m3": "m³",
	"m³": "m³",
}

func normalizeQuoteItemUnit(unit string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(unit))
	if normalized == "" {
		return "UN", nil
	}
	if value, ok := allowedQuoteItemUnits[normalized]; ok {
		return value, nil
	}
	return "", errors.New("item unit must be one of Kg, UN, PÇ, m² or m³")
}

func randomSixDigitPassword() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000), nil
}

func randomToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func centsToBRL(value int64) string {
	return "R$ " + strconv.FormatInt(value/100, 10) + "," + fmt.Sprintf("%02d", value%100)
}

func buildQuotePDF(quote *domain.Quote) []byte {
	canvas := newPDFCanvas()
	canvas.rect(0, 0, 595, 842, "0.953 0.957 0.965")
	canvas.rect(36, 36, 523, 770, "1 1 1")
	canvas.line(36, 676, 559, 676, "0.898 0.906 0.922", 1)

	titleX := 54.0
	if quote.CompanySnapshot.LogoURL != "" {
		canvas.rect(54, 720, 54, 54, "0.953 0.957 0.965")
		canvas.text(64, 748, "LOGO", 10, "F2", "0.420 0.447 0.502")
		titleX = 124
	}
	canvas.text(titleX, 760, truncatePDFText(quote.Title, 44), 22, "F2", "0.067 0.094 0.153")
	canvas.text(titleX, 736, truncatePDFText(quote.CompanySnapshot.Name, 52), 12, "F1", "0.420 0.447 0.502")
	canvas.text(titleX, 716, "Orcamento gerado em "+time.Now().Format("02/01/2006"), 9, "F1", "0.420 0.447 0.502")

	canvas.text(410, 760, "Contato", 11, "F2", "0.067 0.094 0.153")
	canvas.text(410, 740, truncatePDFText(quote.CompanySnapshot.Email, 27), 9, "F1", "0.294 0.333 0.388")
	canvas.text(410, 724, truncatePDFText(quote.CompanySnapshot.Phone, 27), 9, "F1", "0.294 0.333 0.388")
	canvas.text(410, 708, truncatePDFText(quote.CompanySnapshot.Address, 27), 9, "F1", "0.294 0.333 0.388")

	canvas.rect(54, 594, 487, 58, "0.976 0.980 0.984")
	canvas.text(70, 632, "Cliente", 11, "F2", "0.067 0.094 0.153")
	canvas.text(70, 614, truncatePDFText(quote.ClientSnapshot.Name, 54), 11, "F1", "0.294 0.333 0.388")
	canvas.text(300, 614, truncatePDFText(quote.ClientSnapshot.Phone, 30), 11, "F1", "0.294 0.333 0.388")
	canvas.text(70, 600, truncatePDFText(quote.ClientSnapshot.Email, 78), 9, "F1", "0.420 0.447 0.502")

	tableX := 54.0
	tableY := 550.0
	tableW := 487.0
	rowH := 28.0
	canvas.rect(tableX, tableY, tableW, rowH, "0.953 0.957 0.965")
	canvas.text(tableX+14, tableY+10, "Item", 10, "F2", "0.220 0.255 0.318")
	canvas.text(tableX+292, tableY+10, "Qtd.", 10, "F2", "0.220 0.255 0.318")
	canvas.text(tableX+350, tableY+10, "Preco", 10, "F2", "0.220 0.255 0.318")
	canvas.text(tableX+430, tableY+10, "Total", 10, "F2", "0.220 0.255 0.318")

	y := tableY - rowH
	for i, item := range quote.Items {
		if y < 112 {
			canvas.text(tableX+14, y+10, "Itens adicionais omitidos no PDF. Consulte o link publico.", 9, "F1", "0.420 0.447 0.502")
			break
		}
		if i%2 == 0 {
			canvas.rect(tableX, y, tableW, rowH, "1 1 1")
		} else {
			canvas.rect(tableX, y, tableW, rowH, "0.988 0.992 0.996")
		}
		canvas.line(tableX, y, tableX+tableW, y, "0.898 0.906 0.922", 0.5)
		canvas.text(tableX+14, y+10, truncatePDFText(item.Name, 42), 10, "F1", "0.067 0.094 0.153")
		canvas.text(tableX+292, y+10, fmt.Sprintf("%.2f", item.Quantity), 10, "F1", "0.294 0.333 0.388")
		canvas.textRight(tableX+408, y+10, centsToBRL(item.UnitPrice), 10, "F1", "0.294 0.333 0.388")
		canvas.textRight(tableX+526, y+10, centsToBRL(item.Total), 10, "F2", "0.067 0.094 0.153")
		y -= rowH
	}

	canvas.line(54, 92, 541, 92, "0.898 0.906 0.922", 1)
	canvas.text(392, 62, "Total", 16, "F1", "0.067 0.094 0.153")
	canvas.textRight(541, 62, centsToBRL(quote.Total), 18, "F2", "0.067 0.094 0.153")

	return renderPDF(canvas.String())
}

type pdfCanvas struct {
	bytes.Buffer
}

func newPDFCanvas() *pdfCanvas {
	return &pdfCanvas{}
}

func (c *pdfCanvas) rect(x, y, w, h float64, color string) {
	c.WriteString(fmt.Sprintf("q %s rg %.2f %.2f %.2f %.2f re f Q\n", color, x, y, w, h))
}

func (c *pdfCanvas) line(x1, y1, x2, y2 float64, color string, width float64) {
	c.WriteString(fmt.Sprintf("q %s RG %.2f w %.2f %.2f m %.2f %.2f l S Q\n", color, width, x1, y1, x2, y2))
}

func (c *pdfCanvas) text(x, y float64, value string, size int, font, color string) {
	c.WriteString(fmt.Sprintf("BT %s rg /%s %d Tf %.2f %.2f Td (%s) Tj ET\n", color, font, size, x, y, pdfEscape(pdfASCII(value))))
}

func (c *pdfCanvas) textRight(x, y float64, value string, size int, font, color string) {
	width := float64(len(pdfASCII(value))) * float64(size) * 0.52
	c.text(x-width, y, value, size, font, color)
}

func renderPDF(content string) []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		out.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj))
	}
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xref))
	return out.Bytes()
}

func truncatePDFText(value string, limit int) string {
	value = pdfASCII(strings.TrimSpace(value))
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return strings.TrimSpace(value[:limit-3]) + "..."
}

func pdfASCII(value string) string {
	replacements := map[string]string{
		"á": "a", "à": "a", "ã": "a", "â": "a", "ä": "a",
		"Á": "A", "À": "A", "Ã": "A", "Â": "A", "Ä": "A",
		"é": "e", "è": "e", "ê": "e", "ë": "e",
		"É": "E", "È": "E", "Ê": "E", "Ë": "E",
		"í": "i", "ì": "i", "î": "i", "ï": "i",
		"Í": "I", "Ì": "I", "Î": "I", "Ï": "I",
		"ó": "o", "ò": "o", "õ": "o", "ô": "o", "ö": "o",
		"Ó": "O", "Ò": "O", "Õ": "O", "Ô": "O", "Ö": "O",
		"ú": "u", "ù": "u", "û": "u", "ü": "u",
		"Ú": "U", "Ù": "U", "Û": "U", "Ü": "U",
		"ç": "c", "Ç": "C",
	}
	for old, replacement := range replacements {
		value = strings.ReplaceAll(value, old, replacement)
	}
	var builder strings.Builder
	for _, r := range value {
		if r >= 32 && r <= 126 {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	return strings.ReplaceAll(value, ")", "\\)")
}
