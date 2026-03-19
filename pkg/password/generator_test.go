package password

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
		err    bool
	}{
		{"正常长度24", 24, 24, false},
		{"最小长度4", 4, 4, false},
		{"长度10", 10, 10, false},
		{"长度过小", 3, 0, true},
		{"长度0", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.length)
			if (err != nil) != tt.err {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !tt.err && len(got) != tt.want {
				t.Errorf("Generate() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestGenerateContainsAllCharTypes(t *testing.T) {
	// 生成100个密码，确保每个都包含所有字符类型
	for i := 0; i < 100; i++ {
		pwd, err := Generate(24)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false

		for _, c := range pwd {
			switch {
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= 'a' && c <= 'z':
				hasLower = true
			case c >= '0' && c <= '9':
				hasDigit = true
			case strings.Contains("#_-@~", string(c)):
				hasSpecial = true
			}
		}

		if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
			t.Errorf("密码缺少字符类型: upper=%v, lower=%v, digit=%v, special=%v, pwd=%s",
				hasUpper, hasLower, hasDigit, hasSpecial, pwd)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// 生成1000个密码，确保没有重复
	passwords := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		pwd, err := Generate(24)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if passwords[pwd] {
			t.Errorf("生成重复密码: %s", pwd)
		}
		passwords[pwd] = true
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"有效密码", "Abc123#xyz", true},
		{"缺少大写", "abc123#xyz", false},
		{"缺少小写", "ABC123#XYZ", false},
		{"缺少数字", "Abcdef#xyz", false},
		{"缺少特殊字符", "Abc123xyz", false},
		{"包含所有类型", "Ab1#", true},
		{"空密码", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validate(tt.password); got != tt.want {
				t.Errorf("validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
