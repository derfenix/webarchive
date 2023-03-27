# Own Webarchive

Aimed to be a simple, fast and easy-to-use webarchive for personal or home-net usage.

## Requirements 

* Golang 1.19 or higher
* wkhtmltopdf binary in $PATH (to save pages in pdf)

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
            \"all\"
          ]
        }" | jq .
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
xdg-open http://localhost:5001/pages/$page_id/file/$file_id
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
- [ ] Extend configuration
