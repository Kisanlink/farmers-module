package parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileParser_ParseCSV(t *testing.T) {
	parser := NewFileParser()

	tests := []struct {
		name        string
		csvData     string
		expectCount int
		expectError bool
	}{
		{
			name: "Valid CSV with required fields",
			csvData: `first_name,last_name,phone_number,email
John,Doe,9876543210,john@example.com
Jane,Smith,9876543211,jane@example.com`,
			expectCount: 2,
			expectError: false,
		},
		{
			name: "CSV with missing required field",
			csvData: `first_name,last_name,email
John,Doe,john@example.com`,
			expectCount: 0,
			expectError: true,
		},
		{
			name:        "Empty CSV",
			csvData:     ``,
			expectCount: 0,
			expectError: true,
		},
		{
			name:        "CSV with headers only",
			csvData:     `first_name,last_name,phone_number`,
			expectCount: 0,
			expectError: true,
		},
		{
			name: "CSV with semicolon delimiter",
			csvData: `first_name;last_name;phone_number;email
John;Doe;9876543210;john@example.com`,
			expectCount: 1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			farmers, err := parser.ParseCSV([]byte(tt.csvData))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, farmers, tt.expectCount)

				if len(farmers) > 0 {
					farmer := farmers[0]
					assert.NotEmpty(t, farmer.FirstName)
					assert.NotEmpty(t, farmer.LastName)
					assert.NotEmpty(t, farmer.PhoneNumber)
				}
			}
		})
	}
}

func TestFileParser_ParseJSON(t *testing.T) {
	parser := NewFileParser()

	tests := []struct {
		name        string
		jsonData    string
		expectCount int
		expectError bool
	}{
		{
			name: "Valid JSON array",
			jsonData: `[
				{
					"first_name": "John",
					"last_name": "Doe",
					"phone_number": "9876543210",
					"email": "john@example.com"
				},
				{
					"first_name": "Jane",
					"last_name": "Smith",
					"phone_number": "9876543211",
					"email": "jane@example.com"
				}
			]`,
			expectCount: 2,
			expectError: false,
		},
		{
			name: "Valid single JSON object",
			jsonData: `{
				"first_name": "John",
				"last_name": "Doe",
				"phone_number": "9876543210",
				"email": "john@example.com"
			}`,
			expectCount: 1,
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			jsonData:    `{invalid json}`,
			expectCount: 0,
			expectError: true,
		},
		{
			name: "JSON with missing required fields",
			jsonData: `[{
				"first_name": "John",
				"email": "john@example.com"
			}]`,
			expectCount: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			farmers, err := parser.ParseJSON([]byte(tt.jsonData))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, farmers, tt.expectCount)

				if len(farmers) > 0 {
					farmer := farmers[0]
					assert.NotEmpty(t, farmer.FirstName)
					assert.NotEmpty(t, farmer.LastName)
					assert.NotEmpty(t, farmer.PhoneNumber)
				}
			}
		})
	}
}

func TestFileParser_GenerateCSVTemplate(t *testing.T) {
	parser := NewFileParser()

	tests := []struct {
		name           string
		includeExample bool
	}{
		{
			name:           "Template without example",
			includeExample: false,
		},
		{
			name:           "Template with example",
			includeExample: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := parser.GenerateCSVTemplate(tt.includeExample)
			require.NoError(t, err)
			assert.NotEmpty(t, content)

			// Check that content contains headers
			contentStr := string(content)
			assert.Contains(t, contentStr, "first_name")
			assert.Contains(t, contentStr, "last_name")
			assert.Contains(t, contentStr, "phone_number")

			if tt.includeExample {
				// Should contain example data
				lines := len(strings.Split(contentStr, "\n"))
				assert.GreaterOrEqual(t, lines, 2) // Header + example row
			}
		})
	}
}

func TestFileParser_PhoneNumberValidation(t *testing.T) {
	parser := NewFileParser().(*FileParserImpl)

	tests := []struct {
		name   string
		phone  string
		expect bool
	}{
		{"Valid 10-digit starting with 9", "9876543210", true},
		{"Valid 10-digit starting with 8", "8876543210", true},
		{"Valid 10-digit starting with 7", "7876543210", true},
		{"Valid 10-digit starting with 6", "6876543210", true},
		{"Invalid starting with 5", "5876543210", false},
		{"Invalid starting with 1", "1876543210", false},
		{"Invalid 9 digits", "987654321", false},
		{"Invalid 11 digits", "98765432100", false},
		{"Invalid with letters", "987654321a", false},
		{"Empty phone", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isValidPhoneNumber(tt.phone)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestFileParser_EmailValidation(t *testing.T) {
	parser := NewFileParser().(*FileParserImpl)

	tests := []struct {
		name   string
		email  string
		expect bool
	}{
		{"Valid email", "test@example.com", true},
		{"Valid email with subdomain", "test@mail.example.com", true},
		{"Invalid without @", "testexample.com", false},
		{"Invalid without domain", "test@", false},
		{"Invalid without dot", "test@example", false},
		{"Empty email", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isValidEmail(tt.email)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestFileParser_NormalizePhoneNumber(t *testing.T) {
	parser := NewFileParser().(*FileParserImpl)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"10-digit number", "9876543210", "9876543210"},
		{"Number with spaces", "987 654 3210", "9876543210"},
		{"Number with dashes", "987-654-3210", "9876543210"},
		{"Number with +91 prefix", "+919876543210", "9876543210"},
		{"Number with 91 prefix", "919876543210", "9876543210"},
		{"Number with 091 prefix", "0919876543210", "9876543210"},
		{"Number with parentheses", "(987) 654-3210", "9876543210"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.normalizePhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileParser_DetectDelimiter(t *testing.T) {
	parser := NewFileParser().(*FileParserImpl)

	tests := []struct {
		name     string
		data     string
		expected rune
	}{
		{"Comma separated", "a,b,c\n1,2,3", ','},
		{"Semicolon separated", "a;b;c\n1;2;3", ';'},
		{"Tab separated", "a\tb\tc\n1\t2\t3", '\t'},
		{"Mixed delimiters - comma wins", "a,b,c;d\n1,2,3;4", ','},
		{"No clear delimiter", "abc\n123", ','},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.detectDelimiter([]byte(tt.data))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
