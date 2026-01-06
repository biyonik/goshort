# 12 kez URL oluştur (limit 10)
for i in {1..12}; do
  echo "Request $i:"
  curl -s -X POST http://localhost:8080/api/urls \
    -H "Content-Type: application/json" \
    -d '{"long_url": "https://google.com"}' | head -1
  echo ""
done