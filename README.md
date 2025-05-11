# url-shortner

## Build using docker 

docker build -t url-shortener -f url-shortener/Dockerfile .

## Run 

docker run -p 8081:8080 url-shortener

## Test

```
curl -X POST -d "{\"long_url\": \"https://www.google.com\"}" http://localhost:8081/shorten
curl http://localhost:8081/shortened_url
```