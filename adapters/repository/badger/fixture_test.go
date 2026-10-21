package badger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/derfenix/webarchive/entity"
)

const fixturePath = "testdata/page_v1.msgpack"

var (
	fixturePageID = uuid.MustParse("11111111-2222-3333-4444-555555555555")
	fixtureFileID = uuid.MustParse("66666666-7777-8888-9999-aaaaaaaaaaaa")
	fixtureCreated = time.Date(2024, 5, 17, 10, 20, 30, 123456000, time.UTC)
)

// fixturePage is a deterministic page covering everything the storage format
// has to keep: formats, status, meta and results with files.
func fixturePage() *entity.Page {
	page := entity.NewPage("https://example.com/page", "fixture page", entity.FormatPDF, entity.FormatHeaders)
	page.ID = fixturePageID
	page.Created = fixtureCreated
	page.Status = entity.StatusWithErrors
	page.Meta = entity.Meta{
		Title:       "Fixture",
		Description: "Fixture page",
		Encoding:    "utf-8",
	}
	page.Results = entity.ResultsRO{
		{
			Format: entity.FormatPDF,
			Files: []entity.File{{
				ID:       fixtureFileID,
				Name:     "page.pdf",
				MimeType: "application/pdf",
				Size:     15,
				Data:     []byte("%PDF-1.4 fixure"),
				Created:  fixtureCreated,
			}},
		},
		{
			Format: entity.FormatHeaders,
			Err:    fmt.Errorf("process: boom"),
		},
	}

	return page
}

// TestMarshalFixture pins the on-disk format. Run with UPDATE_FIXTURE=1 to
// regenerate the golden file after a deliberate format change.
func TestMarshalFixture(t *testing.T) {
	t.Parallel()

	data, err := marshal(fixturePage())
	require.NoError(t, err)

	if os.Getenv("UPDATE_FIXTURE") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(fixturePath), 0o755))
		require.NoError(t, os.WriteFile(fixturePath, data, 0o600))
	}

	stored, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "fixture is missing, generate it with UPDATE_FIXTURE=1")

	if !bytes.Equal(stored, data) {
		t.Fatalf("storage format changed: stored %d bytes, marshaled %d bytes", len(stored), len(data))
	}
}

func TestUnmarshalFixture(t *testing.T) {
	t.Parallel()

	stored, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	var page entity.Page

	require.NoError(t, unmarshal(stored, &page))

	assert.Equal(t, fixturePageID, page.ID)
	assert.Equal(t, "https://example.com/page", page.URL)
	assert.Equal(t, "fixture page", page.Description)
	assert.True(t, fixtureCreated.Equal(page.Created), "created: want %s, got %s", fixtureCreated, page.Created)
	assert.Equal(t, entity.StatusWithErrors, page.Status)
	assert.Equal(t, []entity.Format{entity.FormatPDF, entity.FormatHeaders}, page.Formats)
	assert.Equal(t, entity.Meta{Title: "Fixture", Description: "Fixture page", Encoding: "utf-8"}, page.Meta)

	require.Len(t, page.Results, 2)

	assert.Equal(t, entity.FormatPDF, page.Results[0].Format)
	require.Len(t, page.Results[0].Files, 1)
	assert.Equal(t, fixtureFileID, page.Results[0].Files[0].ID)
	assert.Equal(t, "page.pdf", page.Results[0].Files[0].Name)
	assert.Equal(t, "application/pdf", page.Results[0].Files[0].MimeType)
	assert.Equal(t, []byte("%PDF-1.4 fixure"), page.Results[0].Files[0].Data)

	assert.Equal(t, entity.FormatHeaders, page.Results[1].Format)

	// Discovered behaviour, not an aspiration: the error interface does not
	// survive the round-trip, so a failed format is persisted without a reason.
	t.Logf("result error after round-trip: %v", page.Results[1].Err)
}
