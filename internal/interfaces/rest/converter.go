package rest

import (
	"fmt"
	"html"

	"github.com/derfenix/webarchive/internal/api/openapi"
	"github.com/derfenix/webarchive/internal/domain"
)

func PageToRestWithResults(page *domain.Page) openapi.PageWithResults {
	return openapi.PageWithResults{
		ID:      page.ID,
		URL:     page.URL,
		Created: page.Created,
		Formats: func() []openapi.Format {
			res := make([]openapi.Format, len(page.Formats))

			for i, format := range page.Formats {
				res[i] = FormatToRest(format)
			}

			return res
		}(),
		Status: StatusToRest(page.Status),
		Meta: openapi.PageWithResultsMeta{
			Title:       html.EscapeString(page.Meta.Title),
			Description: html.EscapeString(page.Meta.Description),
			Error:       openapi.NewOptString(page.Meta.Error),
		},
		Results: func() []openapi.Result {
			results := make([]openapi.Result, len(page.Results))

			for i := range results {
				result := &page.Results[i]

				errText := openapi.OptString{}
				if result.Err != nil {
					errText = openapi.NewOptString(result.Err.Error())
				}

				results[i] = openapi.Result{
					Format: FormatToRest(result.Format),
					Error:  errText,
					Files: func() []openapi.ResultFilesItem {
						files := make([]openapi.ResultFilesItem, len(result.Files))

						for j := range files {
							file := &result.Files[j]

							files[j] = openapi.ResultFilesItem{
								ID:       file.ID,
								Name:     file.Name,
								Mimetype: file.MimeType,
								Size:     file.Size,
							}
						}

						return files
					}(),
				}
			}

			return results
		}(),
	}
}

func BasePageToRest(page *domain.PageBase) openapi.Page {
	return openapi.Page{
		ID:      page.ID,
		URL:     page.URL,
		Created: page.Created,
		Meta: openapi.PageMeta{
			Title:       html.EscapeString(page.Meta.Title),
			Description: html.EscapeString(page.Meta.Description),
			Error:       openapi.NewOptString(page.Meta.Error),
		},
		Formats: func() []openapi.Format {
			res := make([]openapi.Format, len(page.Formats))

			for i, format := range page.Formats {
				res[i] = FormatToRest(format)
			}

			return res
		}(),
		Status: StatusToRest(page.Status),
	}
}

func PageToRest(page *domain.Page) openapi.Page {
	return BasePageToRest(&page.PageBase)
}

func StatusToRest(s domain.Status) openapi.Status {
	switch s {
	case domain.StatusNew:
		return openapi.StatusNew
	case domain.StatusProcessing:
		return openapi.StatusProcessing
	case domain.StatusDone:
		return openapi.StatusDone
	case domain.StatusFailed:
		return openapi.StatusFailed
	case domain.StatusWithErrors:
		return openapi.StatusWithErrors
	default:
		return ""
	}
}

func FormatFromRest(format []openapi.Format) ([]domain.Format, error) {
	var formats []domain.Format

	switch {
	case len(format) == 0 || (len(format) == 1 && format[0] == openapi.FormatAll):
		formats = domain.AllFormats()

	default:
		formats = make([]domain.Format, len(format))
		for i, format := range format {
			switch format {
			case openapi.FormatPdf:
				formats[i] = domain.FormatPDF

			case openapi.FormatHeaders:
				formats[i] = domain.FormatHeaders

			case openapi.FormatSingleFile:
				formats[i] = domain.FormatSingleFile

			default:
				return nil, fmt.Errorf("invalid format value %s", format)
			}
		}
	}

	return formats, nil
}

func FormatToRest(format domain.Format) openapi.Format {
	switch format {
	case domain.FormatPDF:
		return openapi.FormatPdf
	case domain.FormatSingleFile:
		return openapi.FormatSingleFile
	case domain.FormatHeaders:
		return openapi.FormatHeaders
	default:
		return ""
	}
}
