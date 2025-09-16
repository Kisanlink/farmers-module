package parsers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/xuri/excelize/v2"
)

// FileParser interface for parsing different file formats
type FileParser interface {
	ParseCSV(data []byte) ([]*requests.FarmerBulkData, error)
	ParseExcel(data []byte) ([]*requests.FarmerBulkData, error)
	ParseJSON(data []byte) ([]*requests.FarmerBulkData, error)
	GenerateCSVTemplate(includeExample bool) ([]byte, error)
	GenerateExcelTemplate(includeExample bool) ([]byte, error)
}

// FileParserImpl implements FileParser
type FileParserImpl struct {
	config *ParserConfig
}

// ParserConfig contains configuration for file parsing
type ParserConfig struct {
	MaxRecords        int
	RequiredFields    []string
	AllowedDelimiters []rune
	DateFormats       []string
	DefaultCountry    string
}

// NewFileParser creates a new file parser with default configuration
func NewFileParser() FileParser {
	config := &ParserConfig{
		MaxRecords:        10000,
		RequiredFields:    []string{"first_name", "last_name", "phone_number"},
		AllowedDelimiters: []rune{',', ';', '\t'},
		DateFormats: []string{
			"2006-01-02",
			"02/01/2006",
			"02-01-2006",
			"2006/01/02",
		},
		DefaultCountry: "India",
	}

	return &FileParserImpl{
		config: config,
	}
}

// ParseCSV parses CSV data and returns farmer bulk data
func (p *FileParserImpl) ParseCSV(data []byte) ([]*requests.FarmerBulkData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	// Detect delimiter
	delimiter, err := p.detectDelimiter(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect CSV delimiter: %w", err)
	}

	// Create CSV reader
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = delimiter
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found in CSV")
	}

	// First row should be headers
	headers := p.normalizeHeaders(records[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("no headers found in CSV")
	}

	// Validate required headers
	if err := p.validateHeaders(headers); err != nil {
		return nil, fmt.Errorf("invalid CSV headers: %w", err)
	}

	// Parse data rows
	var farmers []*requests.FarmerBulkData
	for i, record := range records[1:] {
		if len(record) == 0 {
			continue // Skip empty rows
		}

		// Check max records limit
		if len(farmers) >= p.config.MaxRecords {
			return nil, fmt.Errorf("exceeded maximum record limit of %d", p.config.MaxRecords)
		}

		farmer, err := p.parseCSVRecord(headers, record, i+1)
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", i+2, err) // +2 because we skip header
		}

		if farmer != nil {
			farmers = append(farmers, farmer)
		}
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("no valid farmer records found")
	}

	return farmers, nil
}

// ParseExcel parses Excel data and returns farmer bulk data
func (p *FileParserImpl) ParseExcel(data []byte) ([]*requests.FarmerBulkData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty Excel data")
	}

	// Create a temporary file-like reader
	file, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get first sheet name
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]

	// Get all rows from the first sheet
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in Excel sheet")
	}

	// First row should be headers
	if len(rows[0]) == 0 {
		return nil, fmt.Errorf("no headers found in Excel sheet")
	}

	headers := p.normalizeHeaders(rows[0])

	// Validate required headers
	if err := p.validateHeaders(headers); err != nil {
		return nil, fmt.Errorf("invalid Excel headers: %w", err)
	}

	// Parse data rows
	var farmers []*requests.FarmerBulkData
	for i, row := range rows[1:] {
		if len(row) == 0 {
			continue // Skip empty rows
		}

		// Check max records limit
		if len(farmers) >= p.config.MaxRecords {
			return nil, fmt.Errorf("exceeded maximum record limit of %d", p.config.MaxRecords)
		}

		farmer, err := p.parseCSVRecord(headers, row, i+1) // Reuse CSV parsing logic
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", i+2, err)
		}

		if farmer != nil {
			farmers = append(farmers, farmer)
		}
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("no valid farmer records found")
	}

	return farmers, nil
}

// ParseJSON parses JSON data and returns farmer bulk data
func (p *FileParserImpl) ParseJSON(data []byte) ([]*requests.FarmerBulkData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	var farmers []*requests.FarmerBulkData

	// Try to parse as array first
	err := json.Unmarshal(data, &farmers)
	if err != nil {
		// Try to parse as single object
		var singleFarmer requests.FarmerBulkData
		if err2 := json.Unmarshal(data, &singleFarmer); err2 != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		farmers = []*requests.FarmerBulkData{&singleFarmer}
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("no farmer records found in JSON")
	}

	// Check max records limit
	if len(farmers) > p.config.MaxRecords {
		return nil, fmt.Errorf("exceeded maximum record limit of %d", p.config.MaxRecords)
	}

	// Validate each farmer record
	for i, farmer := range farmers {
		if err := p.validateFarmerData(farmer, i); err != nil {
			return nil, fmt.Errorf("validation error for record %d: %w", i, err)
		}

		// Set defaults
		p.setFarmerDefaults(farmer)
	}

	return farmers, nil
}

// GenerateCSVTemplate generates a CSV template with headers
func (p *FileParserImpl) GenerateCSVTemplate(includeExample bool) ([]byte, error) {
	headers := []string{
		"first_name",
		"last_name",
		"phone_number",
		"email",
		"date_of_birth",
		"gender",
		"street_address",
		"city",
		"state",
		"postal_code",
		"land_ownership_type",
		"external_id",
	}

	var rows [][]string
	rows = append(rows, headers)

	if includeExample {
		example := []string{
			"John",
			"Doe",
			"9876543210",
			"john.doe@example.com",
			"1990-01-15",
			"male",
			"123 Farm Street",
			"Mumbai",
			"Maharashtra",
			"400001",
			"owned",
			"FARMER001",
		}
		rows = append(rows, example)
	}

	// Convert to CSV format
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return []byte(buf.String()), nil
}

// GenerateExcelTemplate generates an Excel template
func (p *FileParserImpl) GenerateExcelTemplate(includeExample bool) ([]byte, error) {
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	sheetName := "Farmers"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %w", err)
	}

	// Set headers
	headers := []string{
		"first_name",
		"last_name",
		"phone_number",
		"email",
		"date_of_birth",
		"gender",
		"street_address",
		"city",
		"state",
		"postal_code",
		"land_ownership_type",
		"external_id",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		_ = file.SetCellValue(sheetName, cell, header)
	}

	// Add example data if requested
	if includeExample {
		example := []interface{}{
			"John",
			"Doe",
			"9876543210",
			"john.doe@example.com",
			"1990-01-15",
			"male",
			"123 Farm Street",
			"Mumbai",
			"Maharashtra",
			"400001",
			"owned",
			"FARMER001",
		}

		for i, value := range example {
			cell := fmt.Sprintf("%c2", 'A'+i)
			_ = file.SetCellValue(sheetName, cell, value)
		}
	}

	// Set active sheet
	file.SetActiveSheet(index)
	_ = file.DeleteSheet("Sheet1") // Remove default sheet

	// Save to buffer
	var buf strings.Builder
	if err := file.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return []byte(buf.String()), nil
}

// Helper methods

func (p *FileParserImpl) detectDelimiter(data []byte) (rune, error) {
	sample := string(data[:min(1000, len(data))]) // Use first 1000 chars for detection

	counts := make(map[rune]int)
	for _, delimiter := range p.config.AllowedDelimiters {
		counts[delimiter] = strings.Count(sample, string(delimiter))
	}

	// Find the most common delimiter
	maxCount := 0
	var bestDelimiter rune = ','

	for delimiter, count := range counts {
		if count > maxCount {
			maxCount = count
			bestDelimiter = delimiter
		}
	}

	if maxCount == 0 {
		return ',', nil // Default to comma
	}

	return bestDelimiter, nil
}

func (p *FileParserImpl) normalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, header := range headers {
		// Convert to lowercase and replace spaces/special chars with underscores
		normalized[i] = strings.ToLower(strings.TrimSpace(header))
		normalized[i] = strings.ReplaceAll(normalized[i], " ", "_")
		normalized[i] = strings.ReplaceAll(normalized[i], "-", "_")
		normalized[i] = strings.ReplaceAll(normalized[i], ".", "_")
	}
	return normalized
}

func (p *FileParserImpl) validateHeaders(headers []string) error {
	headerSet := make(map[string]bool)
	for _, header := range headers {
		headerSet[header] = true
	}

	var missingFields []string
	for _, required := range p.config.RequiredFields {
		if !headerSet[required] {
			missingFields = append(missingFields, required)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %v", missingFields)
	}

	return nil
}

func (p *FileParserImpl) parseCSVRecord(headers []string, record []string, rowNum int) (*requests.FarmerBulkData, error) {
	if len(record) != len(headers) {
		// Pad or trim record to match headers
		if len(record) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, record)
			record = padded
		} else {
			record = record[:len(headers)]
		}
	}

	farmer := &requests.FarmerBulkData{
		CustomFields: make(map[string]string),
	}

	// Map headers to struct fields
	for i, header := range headers {
		value := strings.TrimSpace(record[i])
		if value == "" {
			continue
		}

		switch header {
		case "first_name":
			farmer.FirstName = value
		case "last_name":
			farmer.LastName = value
		case "phone_number":
			farmer.PhoneNumber = p.normalizePhoneNumber(value)
		case "email":
			farmer.Email = value
		case "date_of_birth":
			farmer.DateOfBirth = p.normalizeDate(value)
		case "gender":
			farmer.Gender = strings.ToLower(value)
		case "street_address":
			farmer.StreetAddress = value
		case "city":
			farmer.City = value
		case "state":
			farmer.State = value
		case "postal_code":
			farmer.PostalCode = value
		case "country":
			farmer.Country = value
		case "land_ownership_type":
			farmer.LandOwnershipType = value
		case "external_id":
			farmer.ExternalID = value
		case "password":
			farmer.Password = value
		default:
			// Store as custom field
			farmer.CustomFields[header] = value
		}
	}

	// Validate required fields
	if err := p.validateFarmerData(farmer, rowNum); err != nil {
		return nil, err
	}

	// Set defaults
	p.setFarmerDefaults(farmer)

	return farmer, nil
}

func (p *FileParserImpl) validateFarmerData(farmer *requests.FarmerBulkData, rowNum int) error {
	var errors []string

	if farmer.FirstName == "" {
		errors = append(errors, "first_name is required")
	}

	if farmer.LastName == "" {
		errors = append(errors, "last_name is required")
	}

	if farmer.PhoneNumber == "" {
		errors = append(errors, "phone_number is required")
	} else if !p.isValidPhoneNumber(farmer.PhoneNumber) {
		errors = append(errors, "invalid phone_number format")
	}

	if farmer.Email != "" && !p.isValidEmail(farmer.Email) {
		errors = append(errors, "invalid email format")
	}

	if farmer.Gender != "" && !p.isValidGender(farmer.Gender) {
		errors = append(errors, "invalid gender (must be male, female, or other)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}

	return nil
}

func (p *FileParserImpl) setFarmerDefaults(farmer *requests.FarmerBulkData) {
	if farmer.Country == "" {
		farmer.Country = p.config.DefaultCountry
	}

	if farmer.ExternalID == "" {
		// Generate external ID if not provided
		farmer.ExternalID = fmt.Sprintf("FARMER_%s_%d",
			strings.ReplaceAll(farmer.PhoneNumber, " ", ""),
			len(farmer.FirstName)+len(farmer.LastName))
	}
}

func (p *FileParserImpl) normalizePhoneNumber(phone string) string {
	// Remove non-digit characters
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Handle Indian phone numbers
	if len(digits) == 10 {
		return digits
	} else if len(digits) == 12 && strings.HasPrefix(digits, "91") {
		return digits[2:]
	} else if len(digits) == 13 && strings.HasPrefix(digits, "091") {
		return digits[3:]
	}

	return digits
}

func (p *FileParserImpl) normalizeDate(date string) string {
	if date == "" {
		return ""
	}

	// Handle YYYYMMDD format
	if _, err := strconv.Atoi(date); err == nil && len(date) == 8 {
		return date[:4] + "-" + date[4:6] + "-" + date[6:8]
	}

	return date // Return as-is if can't parse
}

func (p *FileParserImpl) isValidPhoneNumber(phone string) bool {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Indian mobile numbers are 10 digits and start with 6-9
	if len(digits) == 10 {
		firstDigit := digits[0]
		return firstDigit >= '6' && firstDigit <= '9'
	}

	return false
}

func (p *FileParserImpl) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (p *FileParserImpl) isValidGender(gender string) bool {
	gender = strings.ToLower(gender)
	return gender == "male" || gender == "female" || gender == "other" || gender == "m" || gender == "f"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
