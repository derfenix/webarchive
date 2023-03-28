# Own Webarchive

Aimed to be a simple, fast and easy-to-use webarchive for personal or home-net usage.

## Supported store formats

* **headers** — save all headers from response
* **pdf** — save page in pdf
* **single_file** — save html and all its resources (css,js,images) into one html file

## Requirements 

* Golang 1.19 or higher
* wkhtmltopdf binary in $PATH (to save pages in pdf)

## Configuration

The service can be configured via environment variables. There is a list of available
variables:

* **DB**
  * **DB_PATH** — path for the database files (default `./db`)
* **LOGGING**
  * **LOGGING_DEBUG** — enable debug logs (default `false`)
* **API**
  * **API_ADDRESS** — address the API server will listen (default `0.0.0.0:5001`)
* **PDF**
  * **PDF_LANDSCAPE** — use landscape page orientation instead of portrait (default `false`)
  * **PDF_GRAYSCALE** — use grayscale filter for the output pdf (default `false`)
  * **PDF_MEDIA_PRINT** — use media type `print` for the request (default `true`)
  * **PDF_ZOOM** — zoom page (default `1.0` i.e. no actual zoom)
  * **PDF_VIEWPORT** — use specified viewport value (default `1920x1080`)
  * **PDF_DPI** — use specified DPI value for the output pdf (default `300`)
  * **PDF_FILENAME** — use specified name for output pdf file (default `page.pdf`)


*Note*: Prefix **WEBARCHIVE_** can be used with the environment variable names 
in case of any conflicts.

## Usage

#### 1. Start the server

```shell
go run ./cmd/server/main.go
```

#### 2. Add a page

```shell
curl -X POST --location "http://localhost:5001/pages" \
    -H "Content-Type: application/json" \
    -d "{
          \"url\": \"https://github.com/wkhtmltopdf/wkhtmltopdf/issues/1937\",
          \"formats\": [
            \"pdf\",
            \"headers\"
          ]
        }" | jq .
```

or

```shell
curl -X POST --location \
  "http://localhost:5001/pages?url=https%3A%2F%2Fgithub.com%2Fwkhtmltopdf%2Fwkhtmltopdf%2Fissues%2F1937&formats=pdf%2Cheaders&description=Foo+Bar"
```

#### 3. Get the page's info

```shell
curl -X GET --location "http://localhost:5001/pages/$page_id" | jq .
```
where `$page_id` — value of the `id` field from previous command response.
If `status` field in response is `success` (or `with_errors`) - the `results` field
will contain all processed formats with ids of the stored files.

#### 4. Open file in browser

```shell
xdg-open "http://localhost:5001/pages/$page_id/file/$file_id"
```
Where  `$page_id` — value of the `id` field from previous command response, and
`$file_id` — the id of interesting file.

#### 5. List all stored pages

```shell
curl -X GET --location "http://localhost:5001/pages" | jq .
```

### Roadmap

- [x] Save page to pdf 
- [x] Save URL headers
- [ ] Save page to the single-page html
- [ ] Save page to html with separate resource files (?)
- [ ] Optional authentication
- [ ] Multi-user access
- [ ] Support PostgreSQL
