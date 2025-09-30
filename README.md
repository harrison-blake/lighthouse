## Lighthouse
A go exectuable that takes content in markdown format, injects it into html templates to build staic webpages.
Implements a go worker pool to take advantage of concurrency. Custom markdown parser that seperates frontmatter and content.
Currently being used to build my portfolio [here](https//www.harrisonblake.net).

Hosted on an all Amazon stack
- S3 Bucket for storage
- Route53 for DNS services
- CloudFront for edge distribution

## Development
```
go install

lighthouse build
or
lighthouse build --serve (for localhost:8080 access)

or

go run . build

go test ./... (run all tests in module)
```

## Local Deploy

sample .envrc file
```
export S3_BUCKET=
export CLOUDFRONT_ID_SUB=
export CLOUDFRONT_ID_ROOT=

make deploy
```

## CI/CD w/ Github actions
add same secrets to github repo under `actions secrets`

AWS CLI creds needed as well.
```
AWS_SECRET_ACCESS_KEY
AWS_ACCESS_KEY_ID
```
Current workflow deploys on any merge into main or push to main.

