curl -X GET http://localhost:8080/task

curl -X POST http://localhost:8080/task/<TASK_ID> \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/file.pdf"}'

curl -X GET http://localhost:8080/task/<TASK_ID>

curl -O http://localhost:8080/download/<TASK_ID>