# Mocky

Mocky es un microservicio para **simular APIs** con validaciones y respuestas fijas o dinámicas usando plantillas `{{ ... }}`.
La forma de **registrar** mocks es por HTTP en `POST /v1/prototypes`.
La forma de **consumir** los mocks registrados es invocando el **mismo path** pero **anteponiendo** el prefijo `/v1/mocky`.

> Ejemplo: si defines un endpoint `/v1/users`, lo consumirás en `/v1/mocky/v1/users`.

---

## 🗺️ Flujo de uso

1. **Crear** (o actualizar) un mock:
   `POST /v1/prototypes` con un JSON que describe `request` y `response`.

2. **Consumir** el mock:
   Llama al endpoint real con prefijo `/v1/mocky`:
   `/{PREFIX}{urlPathDefinido}` → `"/v1/mocky" + "/v1/users"`

---

## 📦 Estructura de un mock

```json
{
  "request": {
    "method": "POST",
    "urlPath": "/v1/users",
    "headers": {
      "Content-Type": "application/json"
    },
    "path_params": {
      "user_id": "^[0-9a-fA-F]{8}\\-[0-9a-fA-F]{4}\\-[4][0-9a-fA-F]{3}\\-[89abAB][0-9a-fA-F]{3}\\-[0-9a-fA-F]{12}$"
    },
    "bodySchema": {
      "name": "CreateUserSchema",
      "type_schema": "object",
      "aditional_properties": false,
      "properties": [
        { "name": "name", "is_required": true, "type": "string", "min_length": 1, "max_length": 100 },
        { "name": "email", "is_required": true, "type": "string", "min_length": 5, "max_length": 255, "format": "email" }
      ]
    }
  },
  "response": {
    "statusCode": 201,
    "headers": {
      "X-Mocky": "yes"
    },
    "body": {
      "data": {
        "id": "{{random.UUID}}",
        "name": "{{body.name}}",
        "email": "{{body.email}}"
      },
      "success": true
    }
  }
}
```

### Campos clave

* `request.method` – Verbo HTTP (GET/POST/PUT/DELETE, etc.)

* `request.urlPath` – Path que **definirás** y luego consumirás con el prefijo `/v1/mocky`.

* `request.headers` – Coincidencia exacta clave→valor.

* `request.path_params` – Validación por **regex** de parámetros embebidos en el path (tu router debe extraerlos).

* `request.bodySchema` – Reglas de validación del body:

  * `type_schema`: `"object" | "array" | "string" | "number" | "integer" | "boolean"`
  * `properties`: arreglo de campos, cada uno con:
    `name`, `is_required`, `type`, `min_length`, `max_length`, `format` (como `"email"`), `pattern` (regex)
  * `aditional_properties: false` rechaza campos extra.

* `response.statusCode` – **HTTP status** a devolver (opcional, default 200).

* `response.headers` – Headers de salida.

* `response.body` – JSON de respuesta (soporta plantillas `{{ ... }}`).

---

## 🧩 Plantillas `{{ ... }}`

Puedes usar valores del request o generadores aleatorios:

* `{{path.<name>}}` → parámetro de ruta (ej. `{{path.user_id}}`)
* `{{query.<name>}}` → query string (ej. `?limit=50`)
* `{{headers.<Name>}}` → headers (respeta el nombre tal como llega)
* `{{body.<field>}}` → campos del body (soporta anidación `body.user.email`)

**Generadores integrados:**

* `{{random.UUID}}`
* `{{random.Email}}`
* `{{random.Name}}`
* `{{random.Phone}}`
* `{{random.Date(format:'2006-01-02', startDate:'1990-01-01', endDate:'2000-12-31')}}`

  * args opcionales: `format`, `startDate`, `endDate`

---

## 🚀 Crear mocks (POST `/v1/prototypes`)

### Ejemplo 1 — Signup por **código** (respuesta fija)

```bash
curl -X POST http://localhost:8080/v1/prototypes \
  -H 'Content-Type: application/json' \
  -d '{
    "request": {
      "method": "POST",
      "urlPath": "/v1/signup",
      "headers": { "Content-Type": "application/json" },
      "bodySchema": {
        "name": "SignupSchema",
        "type_schema": "object",
        "aditional_properties": false,
        "properties": [
          { "name": "code", "is_required": true, "type": "string", "min_length": 6, "max_length": 6, "pattern": "^[0-9]{6}$" }
        ]
      }
    },
    "response": {
      "statusCode": 201,
      "body": {
        "data": {
          "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
          "email": "example@gmail.com",
          "image_profile": "unaimagen.com/imagen"
        },
        "status_code": 201,
        "success": true
      }
    }
  }'
```

**Consumirlo** (nota el prefijo `/v1/mocky`):

```bash
curl -X POST http://localhost:8080/v1/mocky/v1/signup \
  -H 'Content-Type: application/json' \
  -d '{"code":"123456"}'
```

---

### Ejemplo 2 — Signup por **email** (eco del email del body)

```bash
curl -X POST http://localhost:8080/v1/prototypes \
  -H 'Content-Type: application/json' \
  -d '{
    "request": {
      "method": "POST",
      "urlPath": "/v1/signup",
      "headers": { "Content-Type": "application/json" },
      "bodySchema": {
        "name": "SignupSchema",
        "type_schema": "object",
        "aditional_properties": false,
        "properties": [
          { "name": "email", "is_required": true, "type": "string", "min_length": 5, "max_length": 255, "format": "email" }
        ]
      }
    },
    "response": {
      "statusCode": 201,
      "body": {
        "data": {
          "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
          "email": "{{body.email}}",
          "image_profile": "unaimagen.com/imagen"
        },
        "status_code": 201,
        "success": true
      }
    }
  }'
```

**Consumirlo**:

```bash
curl -X POST http://localhost:8080/v1/mocky/v1/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafa@example.com"}'
```

---

### Ejemplo 3 — Users con `path/query/headers` y datos aleatorios

```bash
curl -X POST http://localhost:8080/v1/prototypes \
  -H 'Content-Type: application/json' \
  -d '{
    "request": {
      "method": "POST",
      "urlPath": "/v1/users",
      "headers": { "Content-Type": "application/json", "Authorization": "^Bearer\\s.+$" },
      "path_params": {
        "user_id": "^[0-9a-fA-F]{8}\\-[0-9a-fA-F]{4}\\-[4][0-9a-fA-F]{3}\\-[89abAB][0-9a-fA-F]{3}\\-[0-9a-fA-F]{12}$"
      },
      "bodySchema": {
        "name": "CreateUserSchema",
        "type_schema": "object",
        "aditional_properties": false,
        "properties": [
          { "name": "name", "is_required": true, "type": "string", "min_length": 1, "max_length": 100 },
          { "name": "email", "is_required": true, "type": "string", "min_length": 5, "max_length": 255, "format": "email" }
        ]
      }
    },
    "response": {
      "statusCode": 201,
      "body": {
        "id": "{{random.UUID}}",
        "user_id": "{{path.user_id}}",
        "name": "{{body.name}}",
        "email": "{{body.email}}",
        "profile": {
          "primary_email": "{{random.Email}}",
          "birthdate": "{{random.Date(format:\'2006-01-02\', startDate:\'1990-01-01\', endDate:\'2000-12-31\')}}"
        },
        "echo": {
          "auth": "{{headers.Authorization}}",
          "limit": "{{query.limit}}"
        }
      }
    }
  }'
```

**Consumirlo**:

```bash
curl -X POST 'http://localhost:8080/v1/mocky/v1/users?limit=50' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer XYZ' \
  -d '{"name":"Rafa","email":"rafa@example.com"}'
```

> ⚠️ Si tu ruta real incluye un segmento dinámico (p. ej. `/v1/users/:user_id`), tu gateway/router debe mapear el `user_id` del path a `path_params` para que Mocky valide la **regex** configurada.

---

## ✅ Buenas prácticas

* Mantén mocks **idempotentes** en desarrollo (respuestas deterministas) a menos que estés probando aleatoriedad.
* Usa `statusCode` en `response` para separar claramente el **HTTP status** del contenido del body.
* `aditional_properties: false` ayuda a detectar campos extra en payloads.
* En `headers` de `request`, si necesitas flexibilidad, usa **regex** (como se ve arriba con `Authorization`).

---

## 🧪 Errores comunes (y cómo diagnosticarlos)

* **400 – body inválido**: revisa `type`, `min_length`, `format` o `pattern`.
* **400/404 – path param inválido**: tu `user_id` no cumple la **regex** definida.
* **Header faltante o diferente**: confirma coincidencia exacta o ajusta a regex.
* **Placeholders sin resolver**: valida el prefijo correcto `body.|query.|headers.|path.` y que el campo exista.

Ejemplo de error:

```json
{
  "error": "body validation failed",
  "details": [
    {"field":"email","message":"invalid format: email"}
  ]
}
```

---

## 🛠️ Troubleshooting rápido

* **No sale por `/v1/mocky/...`**: recuerda que **siempre** debes anteponer `/v1/mocky` para consumir lo que registraste en `/v1/prototypes`.
* **Regex de headers**: si esperas valores variables, usa regex (p. ej. `^Bearer\\s.+$`).
* **Fechas**: asegúrate que `startDate <= endDate` y `format` válido (Go layout).


