package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/martinghunt/ftep"
)

func TestRunSearchWritesTSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "sample" {
			t.Fatalf("result = %q, want sample", got)
		}
		if got := query.Get("query"); got != "sample_accession=SAMN05276490 OR secondary_sample_accession=SAMN05276490" {
			t.Fatalf("query = %q", got)
		}
		if got := query.Get("fields"); got != "secondary_sample_accession,collection_date,country" {
			t.Fatalf("fields = %q", got)
		}
		_, _ = w.Write([]byte(`[{"secondary_sample_accession":"SRS123456","collection_date":"2016-01-01","country":""}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"search", "-a", "SAMN05276490"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "input_accession\tsecondary_sample_accession\tcollection_date\tcountry\n" +
		"SAMN05276490\tSRS123456\t2016-01-01\t.\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunSearchWithLevel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		if got := query.Get("query"); got != "sample_accession=SAMN05276490 OR secondary_sample_accession=SAMN05276490" {
			t.Fatalf("query = %q", got)
		}
		if got := query.Get("fields"); got != "study_accession,secondary_study_accession,sample_accession,secondary_sample_accession,run_accession,instrument_platform,library_layout,fastq_ftp" {
			t.Fatalf("fields = %q", got)
		}
		_, _ = w.Write([]byte(`[{"run_accession":"ERR123456","fastq_ftp":"ftp.sra.ebi.ac.uk/file.fastq.gz"}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"search", "-a", "SAMN05276490", "--level", "run"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "input_accession\tstudy_accession\tsecondary_study_accession\tsample_accession\tsecondary_sample_accession\trun_accession\tinstrument_platform\tlibrary_layout\tfastq_ftp\n" +
		"SAMN05276490\t.\t.\t.\t.\tERR123456\t.\t.\tftp.sra.ebi.ac.uk/file.fastq.gz\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunSearchWritesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		if got := query.Get("query"); got != "run_accession=ERR123456" {
			t.Fatalf("query = %q", got)
		}
		_, _ = w.Write([]byte(`[{"run_accession":"ERR123456","fastq_ftp":"ftp.sra.ebi.ac.uk/file.fastq.gz"}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"search", "-a", "ERR123456", "--outfmt", "json"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var got map[string][]map[string]string
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json output did not unmarshal: %v\n%s", err, stdout)
	}
	if got["ERR123456"][0]["run_accession"] != "ERR123456" {
		t.Fatalf("run_accession = %q", got["ERR123456"][0]["run_accession"])
	}
	if got["ERR123456"][0]["fastq_ftp"] != "ftp.sra.ebi.ac.uk/file.fastq.gz" {
		t.Fatalf("fastq_ftp = %q", got["ERR123456"][0]["fastq_ftp"])
	}
}

func TestRunReadsWritesManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		if got := query.Get("query"); got != "sample_accession=SAMN05276490 OR secondary_sample_accession=SAMN05276490" {
			t.Fatalf("query = %q", got)
		}
		if got := query.Get("fields"); got != "run_accession,fastq_ftp,fastq_md5,fastq_bytes" {
			t.Fatalf("fields = %q", got)
		}
		_, _ = w.Write([]byte(`[{"run_accession":"SRR3675520","fastq_ftp":"ftp.sra.ebi.ac.uk/read_1.fastq.gz;ftp.sra.ebi.ac.uk/read_2.fastq.gz","fastq_md5":"abc;def","fastq_bytes":"10;20"}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"reads", "-a", "SAMN05276490"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "input_accession\trun_accession\tfilename\turl\tmd5\tbytes\n" +
		"SAMN05276490\tSRR3675520\tread_1.fastq.gz\thttps://ftp.sra.ebi.ac.uk/read_1.fastq.gz\tabc\t10\n" +
		"SAMN05276490\tSRR3675520\tread_2.fastq.gz\thttps://ftp.sra.ebi.ac.uk/read_2.fastq.gz\tdef\t20\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunReadsWritesWget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		if got := query.Get("query"); got != "run_accession=ERR123456" {
			t.Fatalf("query = %q", got)
		}
		_, _ = w.Write([]byte(`[{"run_accession":"ERR123456","fastq_ftp":"ftp.sra.ebi.ac.uk/file.fastq.gz","fastq_md5":"abc","fastq_bytes":"10"}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"reads", "-a", "ERR123456", "--format", "wget", "--protocol", "ftp", "--output-dir", "reads"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "wget -c -O 'reads/file.fastq.gz' 'ftp://ftp.sra.ebi.ac.uk/file.fastq.gz'\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunReadsWritesMD5(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"run_accession":"ERR123456","fastq_ftp":"ftp.sra.ebi.ac.uk/file.fastq.gz","fastq_md5":"abc","fastq_bytes":"10"}]`))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"reads", "-a", "ERR123456", "--format", "md5", "--output-dir", "reads"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "abc  reads/file.fastq.gz\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunOpenPrintsURL(t *testing.T) {
	code, stdout := captureStdout(t, func() int {
		return run([]string{"open", "SRR3675520", "--print-url"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "https://www.ebi.ac.uk/ena/browser/view/SRR3675520\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunOpenUsesBrowserOpener(t *testing.T) {
	var openedURL string
	withTestBrowserOpener(t, func(browserURL string) error {
		openedURL = browserURL
		return nil
	})

	code, stdout := captureStdout(t, func() int {
		return run([]string{"open", "-a", "GCA_000195955.2"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}

	const wantURL = "https://www.ebi.ac.uk/ena/browser/view/GCA_000195955.2"
	if openedURL != wantURL {
		t.Fatalf("openedURL = %q, want %q", openedURL, wantURL)
	}
}

func TestRunOpenRejectsInvalidAccession(t *testing.T) {
	code, _ := captureStdout(t, func() int {
		return run([]string{"open", "not-an-accession", "--print-url"})
	})

	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
}

func TestRunGetFieldsListsDataTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/results" {
			t.Fatalf("path = %q, want /results", r.URL.Path)
		}
		_, _ = w.Write([]byte("resultId\tdescription\tprimaryAccessionType\ntls_set\tTargeted locus study contig sets\taccession\nsample\tSamples\tsample_accession\nanalysis\tAnalyses\tanalysis_accession\nread_run\tRaw reads\trun_accession\n"))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"get_fields"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "resultId\tdescription\tprimaryAccessionType\tftep_search\n" +
		"read_run\tRaw reads\trun_accession\tyes\n" +
		"sample\tSamples\tsample_accession\tyes\n" +
		"analysis\tAnalyses\tanalysis_accession\tno\n" +
		"tls_set\tTargeted locus study contig sets\taccession\tno\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestRunGetFieldsForDataType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/searchFields" {
			t.Fatalf("path = %q, want /searchFields", r.URL.Path)
		}
		if got := r.URL.Query().Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		_, _ = w.Write([]byte("columnId\tdescription\nrun_accession\taccession number\n"))
	}))
	defer server.Close()

	withTestClient(t, server)
	code, stdout := captureStdout(t, func() int {
		return run([]string{"get_fields", "read_run"})
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	const want = "columnId\tdescription\nrun_accession\taccession number\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestWriteReadsCurl(t *testing.T) {
	files := []readFile{
		{
			OutputPath: "reads/file.fastq.gz",
			URL:        "https://ftp.sra.ebi.ac.uk/file.fastq.gz",
		},
	}

	var out bytes.Buffer
	if err := writeReads(&out, files, readsFormatCurl); err != nil {
		t.Fatal(err)
	}

	const want = "curl -L --fail --continue-at - --output 'reads/file.fastq.gz' 'https://ftp.sra.ebi.ac.uk/file.fastq.gz'\n"
	if out.String() != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
}

func TestWriteTSVAllFieldsSortsColumnsAndFormatsNil(t *testing.T) {
	results := []ftep.SearchResult{
		{
			InputAccession: "SAMN05276490",
			Records: []ftep.Record{
				{
					"z_field": "last",
					"a_field": "first",
					"m_field": nil,
				},
			},
		},
	}

	var out bytes.Buffer
	if err := writeTSV(&out, results, []string{"ALL"}); err != nil {
		t.Fatal(err)
	}

	const want = "input_accession\ta_field\tm_field\tz_field\n" +
		"SAMN05276490\tfirst\t.\tlast\n"
	if out.String() != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
}

func withTestClient(t *testing.T, server *httptest.Server) {
	t.Helper()

	previous := newClient
	newClient = func() *ftep.Client {
		return &ftep.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}
	}
	t.Cleanup(func() {
		newClient = previous
	})
}

func withTestBrowserOpener(t *testing.T, opener func(string) error) {
	t.Helper()

	previous := openBrowser
	openBrowser = opener
	t.Cleanup(func() {
		openBrowser = previous
	})
}

func captureStdout(t *testing.T, fn func() int) (int, string) {
	t.Helper()

	oldStdout := os.Stdout
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writeEnd
	defer func() {
		os.Stdout = oldStdout
	}()

	code := fn()

	if err := writeEnd.Close(); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if _, err := io.Copy(&stdout, readEnd); err != nil {
		t.Fatal(err)
	}
	if err := readEnd.Close(); err != nil {
		t.Fatal(err)
	}

	return code, stdout.String()
}
