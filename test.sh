# Create a book
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Foo","author":"Bar","isbn":"0123456789"}'

# Get all books
curl http://localhost:8080/books

# Get a specific book
curl http://localhost:8080/books/1

curl http://localhost:8080/books/isbn/0123456789

# Update a book
curl -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Foo (updated)","author":"Bar","isbn":"0123456789"}'

curl -X PUT http://localhost:8080/books/isbn/0123456789 \
  -H "Content-Type: application/json" \
  -d '{"title":"Foo (updated)","author":"Bar","isbn":"0123456789"}'

# Delete a book
curl -X DELETE http://localhost:8080/books/1