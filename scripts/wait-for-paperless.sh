#!/bin/bash
# Wait for Paperless-ngx to be ready

set -e

PAPERLESS_URL="${PAPERLESS_URL:-http://localhost:8000}"
MAX_ATTEMPTS=150
ATTEMPT=0

echo "Waiting for Paperless-ngx at $PAPERLESS_URL to be ready..."

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  if curl -s -f "$PAPERLESS_URL/api/" > /dev/null 2>&1; then
    echo "Paperless-ngx is ready!"

    # Get or create API token
    echo "Getting API token..."

    # Try to get token using docker exec
    if command -v docker &> /dev/null; then
      TOKEN=$(docker exec paperless-test python manage.py shell -c "
from django.contrib.auth.models import User
from rest_framework.authtoken.models import Token
user = User.objects.get(username='admin')
token, created = Token.objects.get_or_create(user=user)
print(token.key)
" 2> /dev/null | tail -1)

      if [ -n "$TOKEN" ]; then
        echo "API Token retrieved successfully"
        echo "Export this token with: export PAPERLESS_TOKEN=\$TOKEN"
        echo ""
        echo "To set the token in your environment, run:"
        echo "  export PAPERLESS_TOKEN='$TOKEN'"
        exit 0
      fi
    fi

    echo "Could not retrieve token automatically. Please get it manually from the Paperless UI."
    echo "Go to http://localhost:8000/admin/authtoken/tokenproxy/ to create/view tokens."
    exit 0
  fi

  ATTEMPT=$((ATTEMPT + 1))
  echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Paperless-ngx not ready yet, waiting..."
  sleep 2
done

echo "Timeout waiting for Paperless-ngx to be ready"
exit 1
