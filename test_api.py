import requests
import jwt
from datetime import datetime, timedelta

# Create token
SECRET_KEY = "super_secret_key_change_in_production"
ALGORITHM = "HS256"

to_encode = {"sub": "admin"}
expire = datetime.utcnow() + timedelta(minutes=15)
to_encode.update({"exp": expire})
token = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

res = requests.get('http://127.0.0.1:8000/api/admin/planos', headers={'Authorization': f'Bearer {token}'})
print(res.status_code)
print(res.text)
