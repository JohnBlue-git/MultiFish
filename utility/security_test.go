package utility

import (
	"testing"
)

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{
			name:     "normal password",
			password: "secretPassword123",
			want:     "******",
		},
		{
			name:     "empty password",
			password: "",
			want:     "",
		},
		{
			name:     "short password",
			password: "pw",
			want:     "******",
		},
		{
			name:     "already masked",
			password: "******",
			want:     "******",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskPassword(tt.password)
			if got != tt.want {
				t.Errorf("MaskPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaskPasswordConsistency(t *testing.T) {
	// Ensure that the same password always produces the same mask
	password := "testPassword"
	mask1 := MaskPassword(password)
	mask2 := MaskPassword(password)
	
	if mask1 != mask2 {
		t.Errorf("MaskPassword() inconsistent: got %v and %v", mask1, mask2)
	}
	
	// Ensure different passwords produce the same mask (security through obscurity)
	password2 := "differentPassword"
	mask3 := MaskPassword(password2)
	
	if mask1 != mask3 {
		t.Errorf("MaskPassword() should produce same mask for different passwords: got %v and %v", mask1, mask3)
	}
}
