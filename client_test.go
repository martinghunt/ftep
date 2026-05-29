package ftep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestQuerySampleToRun(t *testing.T) {
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
	fields, records, err := client.Query(context.Background(), "SAMN05276490", AccessionTypeSample, []string{"DEFAULT"}, true)
	if err != nil {
		t.Fatal(err)
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

func TestSearchRejectsMixedAccessionTypes(t *testing.T) {
	client := &Client{}
	_, err := client.Search(context.Background(), SearchOptions{
		Accessions: []string{"SAMN123456", "ERR123456"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
