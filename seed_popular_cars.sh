API_BASE="https://carezo.onrender.com"
ADMIN_EMAIL="admin@carezo.com"
ADMIN_PASSWORD="Admin123!"

TOKEN=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" \
  | jq -r '.data.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "Login failed — check ADMIN_EMAIL/ADMIN_PASSWORD"
  exit 1
fi
echo "Got token: ${TOKEN:0:20}..."

jq -c '.[]' cars.json | while read -r car; do
  name=$(echo "$car" | jq -r '.model')
  echo "Creating: $name"
  curl -s -X POST "$API_BASE/api/admin/cars" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$car" | jq -r '.message // .error'
  echo "---"
done

