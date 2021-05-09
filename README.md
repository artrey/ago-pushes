# Pushes

1. Run system:

```bash
docker-compose up -d
```

2. Emulate payment event:

```bash
docker-compose run --rm payments --userId 1 --message "Your payment (150$) is successful"
```

3. Read logs:

```bash
docker-compose logs tokens
```

## Optional:

Register new push token via [POST request](./client/requests.http):

```api
POST http://localhost:9999/api/tokens/register
Content-Type: application/json

{
  "userId": 4,
  "pushToken": "TOKEN_FOR_USER4"
}
```
