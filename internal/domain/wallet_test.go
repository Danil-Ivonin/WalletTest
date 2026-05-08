package domain

import "testing"

func TestValidateOperationType(t *testing.T) {
	tests := []struct {
		name string
		op   OperationType
		want error
	}{
		{name: "deposit is valid", op: OperationTypeDeposit, want: nil},
		{name: "withdraw is valid", op: OperationTypeWithdraw, want: nil},
		{name: "empty operation is invalid", op: OperationType(""), want: ErrInvalidOperationType},
		{name: "unknown operation is invalid", op: OperationType("TRANSFER"), want: ErrInvalidOperationType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateOperationType(tt.op); got != tt.want {
				t.Fatalf("ValidateOperationType(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}
