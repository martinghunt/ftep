package ftep

import "testing"

func TestIdentifyAccession(t *testing.T) {
	tests := []struct {
		accession string
		fixed     string
		typ       AccessionType
		ok        bool
	}{
		{"GCA_123456789", "GCA_123456789", AccessionTypeAssembly, true},
		{"GCA_123456789.1", "GCA_123456789", AccessionTypeAssembly, true},
		{"GCA_12345678.1", "", "", false},
		{"G123456.1", "", "", false},
		{"SAMN123456", "SAMN123456", AccessionTypeSample, true},
		{"ERS123456", "ERS123456", AccessionTypeSample, true},
		{"ERS12345", "", "", false},
		{"ERR123456", "ERR123456", AccessionTypeRun, true},
		{"ERR12345", "", "", false},
		{"ERX123456", "ERX123456", AccessionTypeExperiment, true},
	}

	for _, tt := range tests {
		fixed, typ, ok := IdentifyAccession(tt.accession)
		if fixed != tt.fixed || typ != tt.typ || ok != tt.ok {
			t.Fatalf("IdentifyAccession(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.accession, fixed, typ, ok, tt.fixed, tt.typ, tt.ok)
		}
	}
}
