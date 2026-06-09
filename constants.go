package ftep

const BasePortalURL = "https://www.ebi.ac.uk/ena/portal/api/"

type searchEndpoint struct {
	mainType string
	result   string
}

var urlSearchData = map[AccessionType]searchEndpoint{
	AccessionTypeAssembly:   {mainType: "search", result: "assembly"},
	AccessionTypeStudy:      {mainType: "search", result: "study"},
	AccessionTypeSample:     {mainType: "search", result: "sample"},
	AccessionTypeRun:        {mainType: "search", result: "read_run"},
	AccessionTypeExperiment: {mainType: "search", result: "read_run"},
}

var assemblySmall = []string{"accession", "sample_accession", "run_accession", "version"}
var assemblyDefault = append(copyStrings(assemblySmall), "scientific_name", "tax_id")

var studySmall = []string{
	"study_accession",
	"secondary_study_accession",
}
var studyDefault = append(copyStrings(studySmall),
	"study_title",
	"project_name",
)
var studyBig = append(copyStrings(studyDefault),
	"study_description",
	"center_name",
	"broker_name",
	"first_public",
	"last_updated",
	"scientific_name",
	"tax_id",
)

var sampleSmall = []string{
	"study_accession",
	"sample_accession",
}
var sampleDefault = []string{
	"secondary_sample_accession",
	"collection_date",
	"country",
}
var sampleBig = append(copyStrings(sampleDefault),
	"center_name",
	"broker_name",
)

var runSmall = []string{
	"study_accession",
	"secondary_study_accession",
	"sample_accession",
	"secondary_sample_accession",
	"run_accession",
}
var runDefault = append(copyStrings(runSmall), "instrument_platform", "library_layout", "fastq_ftp")
var runBig = append(copyStrings(runDefault),
	"center_name",
	"broker_name",
	"read_count",
	"base_count",
	"collection_date",
	"scientific_name",
)

var fieldPresets = map[AccessionType]map[string][]string{
	AccessionTypeAssembly: {
		"SMALL":   assemblySmall,
		"DEFAULT": assemblyDefault,
		"BIG":     assemblyDefault,
	},
	AccessionTypeStudy: {
		"SMALL":   studySmall,
		"DEFAULT": studyDefault,
		"BIG":     studyBig,
	},
	AccessionTypeSample: {
		"SMALL":   sampleSmall,
		"DEFAULT": sampleDefault,
		"BIG":     sampleBig,
	},
	AccessionTypeRun: {
		"SMALL":   runSmall,
		"DEFAULT": runDefault,
		"BIG":     runBig,
	},
	AccessionTypeExperiment: {
		"SMALL":   runSmall,
		"DEFAULT": runDefault,
		"BIG":     runBig,
	},
}

func copyStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}
