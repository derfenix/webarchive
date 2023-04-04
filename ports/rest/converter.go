package rest

import (
	"fmt"
	"html"

	"github.com/derfenix/webarchive/api/openapi"
	"github.com/derfenix/webarchive/entity"
)

func PageToRestWithResults(page *entity.Page) openapi.PageWithResults {
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
			results := make([]openapi.Result, len(page.Results.Results()))

			for i := range results {
				result := &(page.Results.Results())[i]

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

func PageToRest(page *entity.Page) openapi.Page {
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

func StatusToRest(s entity.Status) openapi.Status {
	switch s {
	case entity.StatusNew:
		return openapi.StatusNew
	case entity.StatusProcessing:
		return openapi.StatusProcessing
	case entity.StatusDone:
		return openapi.StatusDone
	case entity.StatusFailed:
		return openapi.StatusFailed
	case entity.StatusWithErrors:
		return openapi.StatusWithErrors
	default:
		return ""
	}
}

func FormatFromRest(format []openapi.Format) ([]entity.Format, error) {
	var formats []entity.Format

	switch {
	case len(format) == 0 || (len(format) == 1 && format[0] == openapi.FormatAll):
		formats = entity.AllFormats

	default:
		formats = make([]entity.Format, len(format))
		for i, format := range format {
			switch format {
			case openapi.FormatPdf:
				formats[i] = entity.FormatPDF

			case openapi.FormatHeaders:
				formats[i] = entity.FormatHeaders

			case openapi.FormatSingleFile:
				formats[i] = entity.FormatSingleFile

			default:
				return nil, fmt.Errorf("invalid format value %s", format)
			}
		}
	}

	return formats, nil
}

func FormatToRest(format entity.Format) openapi.Format {
	switch format {
	case entity.FormatPDF:
		return openapi.FormatPdf
	case entity.FormatSingleFile:
		return openapi.FormatSingleFile
	case entity.FormatHeaders:
		return openapi.FormatHeaders
	default:
		return ""
	}
}
