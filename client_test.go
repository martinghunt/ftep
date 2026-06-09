package ftep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestQuerySampleAtRunLevel(t *testing.T) {
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
		if got := query.Get("format"); got != "json" {
			t.Fatalf("format = %q, want json", got)
		}
		if got := query.Get("fields"); got != "study_accession,secondary_study_accession,sample_accession,secondary_sample_accession,run_accession,instrument_platform,library_layout,fastq_ftp" {
			t.Fatalf("fields = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"run_accession":"ERR123456","fastq_ftp":""}]`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL + "/", HTTPClient: server.Client()}
	resultType, fields, records, err := client.Query(context.Background(), "SAMN05276490", AccessionTypeSample, []string{"DEFAULT"}, AccessionTypeRun)
	if err != nil {
		t.Fatal(err)
	}
	if resultType != AccessionTypeRun {
		t.Fatalf("resultType = %q, want %q", resultType, AccessionTypeRun)
	}
	if !reflect.DeepEqual(fields, runDefault) {
		t.Fatalf("fields = %#v, want %#v", fields, runDefault)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0]["fastq_ftp"] != nil {
		t.Fatalf("empty string was not normalized to nil: %#v", records[0]["fastq_ftp"])
	}
}

func TestQueryStudy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("result"); got != "study" {
			t.Fatalf("result = %q, want study", got)
		}
		if got := query.Get("query"); got != "study_accession=PRJEB1787 OR secondary_study_accession=PRJEB1787" {
			t.Fatalf("query = %q", got)
		}
		if got := query.Get("format"); got != "json" {
			t.Fatalf("format = %q, want json", got)
		}
		if got := query.Get("fields"); got != "study_accession,secondary_study_accession,study_title,project_name" {
			t.Fatalf("fields = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"study_accession":"PRJEB1787","secondary_study_accession":"ERP001736","study_title":"Tara Oceans"}]`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL + "/", HTTPClient: server.Client()}
	resultType, fields, records, err := client.Query(context.Background(), "PRJEB1787", AccessionTypeStudy, []string{"DEFAULT"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if resultType != AccessionTypeStudy {
		t.Fatalf("resultType = %q, want %q", resultType, AccessionTypeStudy)
	}
	if !reflect.DeepEqual(fields, studyDefault) {
		t.Fatalf("fields = %#v, want %#v", fields, studyDefault)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0]["secondary_study_accession"] != "ERP001736" {
		t.Fatalf("secondary_study_accession = %q", records[0]["secondary_study_accession"])
	}
}

func TestQuerySecondaryStudyAtSampleLevel(t *testing.T) {
	var sawStudyLookup bool
	var sawSampleSearch bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}

		query := r.URL.Query()
		switch query.Get("result") {
		case "study":
			sawStudyLookup = true
			if got := query.Get("query"); got != "study_accession=ERP001736 OR secondary_study_accession=ERP001736" {
				t.Fatalf("study lookup query = %q", got)
			}
			if got := query.Get("fields"); got != "study_accession" {
				t.Fatalf("study lookup fields = %q", got)
			}
			_, _ = w.Write([]byte(`[{"study_accession":"PRJEB1787"}]`))
		case "sample":
			sawSampleSearch = true
			if got := query.Get("query"); got != "study_accession=PRJEB1787" {
				t.Fatalf("sample query = %q", got)
			}
			if got := query.Get("fields"); got != "secondary_sample_accession,collection_date,country" {
				t.Fatalf("sample fields = %q", got)
			}
			_, _ = w.Write([]byte(`[{"secondary_sample_accession":"ERS478017","country":"France"}]`))
		default:
			t.Fatalf("result = %q", query.Get("result"))
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL + "/", HTTPClient: server.Client()}
	resultType, fields, records, err := client.Query(context.Background(), "ERP001736", AccessionTypeStudy, []string{"DEFAULT"}, AccessionTypeSample)
	if err != nil {
		t.Fatal(err)
	}
	if !sawStudyLookup || !sawSampleSearch {
		t.Fatalf("sawStudyLookup=%v sawSampleSearch=%v", sawStudyLookup, sawSampleSearch)
	}
	if resultType != AccessionTypeSample {
		t.Fatalf("resultType = %q, want %q", resultType, AccessionTypeSample)
	}
	if !reflect.DeepEqual(fields, sampleDefault) {
		t.Fatalf("fields = %#v, want %#v", fields, sampleDefault)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
}

func TestResolveSearchLevelRejectsUnsupportedCombination(t *testing.T) {
	_, err := ResolveSearchLevel(AccessionTypeRun, AccessionTypeSample)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAllowedFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/searchFields" {
			t.Fatalf("path = %q, want /searchFields", r.URL.Path)
		}
		if got := r.URL.Query().Get("result"); got != "read_run" {
			t.Fatalf("result = %q, want read_run", got)
		}
		_, _ = w.Write([]byte("field\tdescription\nrun_accession\tRun accession\n"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL + "/", HTTPClient: server.Client()}
	text, err := client.GetAllowedFields(context.Background(), "read_run")
	if err != nil {
		t.Fatal(err)
	}
	if text != "field\tdescription\nrun_accession\tRun accession\n" {
		t.Fatalf("text = %q", text)
	}
}

func TestGetResultTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/results" {
			t.Fatalf("path = %q, want /results", r.URL.Path)
		}
		_, _ = w.Write([]byte("resultId\tdescription\nread_run\tRaw reads\n"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL + "/", HTTPClient: server.Client()}
	text, err := client.GetResultTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if text != "resultId\tdescription\nread_run\tRaw reads\n" {
		t.Fatalf("text = %q", text)
	}
}

func TestSearchRejectsMixedAccessionTypes(t *testing.T) {
	client := &Client{}
	_, err := client.Search(context.Background(), SearchOptions{
		Accessions: []string{"SAMN123456", "ERR123456"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
