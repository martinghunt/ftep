package ftep

import "regexp"

// AccessionType is the ENA result category inferred from an accession.
type AccessionType string

const (
	AccessionTypeAssembly   AccessionType = "assembly"
	AccessionTypeWGSSet     AccessionType = "wgs_set"
	AccessionTypeStudy      AccessionType = "study"
	AccessionTypeSample     AccessionType = "sample"
	AccessionTypeRun        AccessionType = "run"
	AccessionTypeExperiment AccessionType = "experiment"
)

type accessionRegex struct {
	re        *regexp.Regexp
	typ       AccessionType
	normalize func([]string) string
}

var accessionRegexes = []accessionRegex{
	{re: regexp.MustCompile(`^(GCA_[0-9]{9})(\.[0-9]*)*$`), typ: AccessionTypeAssembly, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^(?:([A-Z]{4})0{8,10}|([A-Z]{6})0{9,11})\.([0-9]+)$`), typ: AccessionTypeWGSSet, normalize: normalizeWGSSetMasterAccession},
	{re: regexp.MustCompile(`^(PRJ(?:E|D|N)[A-Z][0-9]+)$`), typ: AccessionTypeStudy, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^((?:E|D|S)RP[0-9]{6,})$`), typ: AccessionTypeStudy, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^(SAM(?:E|D|N)[A-Z]?[0-9]+)$`), typ: AccessionTypeSample, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^((?:E|D|S)RS[0-9]{6,})$`), typ: AccessionTypeSample, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^((?:E|D|S)RR[0-9]{6,})$`), typ: AccessionTypeRun, normalize: firstAccessionMatch},
	{re: regexp.MustCompile(`^((?:E|D|S)RX[0-9]{6,})$`), typ: AccessionTypeExperiment, normalize: firstAccessionMatch},
}

// IdentifyAccession returns the normalized accession, its type, and whether it
// was recognized. Assembly version suffixes are stripped. WGS master
// accessions are normalized to the ENA wgs_set id.
func IdentifyAccession(accession string) (string, AccessionType, bool) {
	for _, candidate := range accessionRegexes {
		matches := candidate.re.FindStringSubmatch(accession)
		if matches != nil {
			return candidate.normalize(matches), candidate.typ, true
		}
	}

	return "", "", false
}

func firstAccessionMatch(matches []string) string {
	return matches[1]
}

func normalizeWGSSetMasterAccession(matches []string) string {
	prefix := matches[1]
	if prefix == "" {
		prefix = matches[2]
	}

	version := matches[3]
	if len(version) == 1 {
		version = "0" + version
	}
	return prefix + version
}
