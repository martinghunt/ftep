package ftep

import "regexp"

// AccessionType is the ENA result category inferred from an accession.
type AccessionType string

const (
	AccessionTypeAssembly   AccessionType = "assembly"
	AccessionTypeStudy      AccessionType = "study"
	AccessionTypeSample     AccessionType = "sample"
	AccessionTypeRun        AccessionType = "run"
	AccessionTypeExperiment AccessionType = "experiment"
)

type accessionRegex struct {
	re  *regexp.Regexp
	typ AccessionType
}

var accessionRegexes = []accessionRegex{
	{regexp.MustCompile(`^(GCA_[0-9]{9})(\.[0-9]*)*$`), AccessionTypeAssembly},
	{regexp.MustCompile(`^(PRJ(?:E|D|N)[A-Z][0-9]+)$`), AccessionTypeStudy},
	{regexp.MustCompile(`^((?:E|D|S)RP[0-9]{6,})$`), AccessionTypeStudy},
	{regexp.MustCompile(`^(SAM(?:E|D|N)[A-Z]?[0-9]+)$`), AccessionTypeSample},
	{regexp.MustCompile(`^((?:E|D|S)RS[0-9]{6,})$`), AccessionTypeSample},
	{regexp.MustCompile(`^((?:E|D|S)RR[0-9]{6,})$`), AccessionTypeRun},
	{regexp.MustCompile(`^((?:E|D|S)RX[0-9]{6,})$`), AccessionTypeExperiment},
}

// IdentifyAccession returns the normalized accession, its type, and whether it
// was recognized. Assembly version suffixes are stripped.
func IdentifyAccession(accession string) (string, AccessionType, bool) {
	for _, candidate := range accessionRegexes {
		matches := candidate.re.FindStringSubmatch(accession)
		if matches != nil {
			return matches[1], candidate.typ, true
		}
	}

	return "", "", false
}
