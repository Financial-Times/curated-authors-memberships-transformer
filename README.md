# curated-authors-transformer

[![CircleCI](https://circleci.com/gh/Financial-Times/curated-authors-transformer.svg?style=svg)](https://circleci.com/gh/Financial-Times/curated-authors-transformer)

Retrieves author data curated by editorial people and transforms it to People according to UP JSON model.
The authors data is specified by a Google spreadsheet which is accessible through [Bertha API](https://github.com/ft-interactive/bertha/wiki/Tutorial).
The service exposes endpoints for getting all the curated authors UUIDs and for getting authors by uuid.

# How to run

## Locally: 

`go get github.com/Financial-Times/curated-authors-transformer`

`$GOPATH/bin/ ./curated-authors-transformer  --bertha-source-url=<BERTHA_SOURCE_URL> --port=8080`                

```
export|set PORT=8080
export|set BERTHA_SOURCE_URL="http://bertha.ig.ft.com/view/publish/gss/1wEdVRLtayZ6-XBfYM3vKAGaOV64cNJD3L8MlLM8-uFY/TestAuthors"
$GOPATH/bin/curated-authors-transformer
```

## With Docker:

`docker build -t coco/curated-authors-transformer .`

`docker run -ti --env BERTHA_SOURCE_URL=<bertha_url> coco/curated-authors-transformer`

#Endpoints

* `/transformers/authors/__count` - number of available authors to be transformed 
* `/transformers/authors/__ids` - Get the list of authors UUIDs available to be transformed 
* `/transformers/authors/{uuid}` - Get author data of the given uuid