# My Own HTTP Server

An HTTP (Hypertext Transfer Protocol) server serves as an intermediary between clients (usually web browsers) and the web resources that those clients request. Its main function is to receive HTTP requests from clients and send them the corresponding responses. In essence, it is responsible for delivering the web content (web pages, images, videos, etc.) that you see in your browser.

### How to run
```bash
cd cmd
go run .
```

### How to test 
```bash
curl -v http://localhost:4221/echo/abc
```

### How to test gzip
```bash
curl -v http://localhost:4221/echo/abc -H "Accept-Encoding: gzip"
```

### How to test file upload
```bash
curl -v -F "file=@{path of file to upload}" http://localhost:4221/files/{name of file}
```

### How to test file download
```bash
curl -v http://localhost:4221/files/{name of file}
``` 