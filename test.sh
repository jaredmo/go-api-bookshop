# Create a book
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Slaughterhouse-Five","author":"Kurt Vonnegut Jr.","isbn":"978-0385333849"}'

# Get all books
curl http://localhost:8080/books

# Get a specific book
curl http://localhost:8080/books/1

# Update a book
curl -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Slaughterhouse-Five (TEST UPDATE)","author":"Kurt Vonnegut Jr.","isbn":"978-0385333849"}'

# Delete a book
curl -X DELETE http://localhost:8080/books/1