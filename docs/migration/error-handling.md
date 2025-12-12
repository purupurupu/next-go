# エラーハンドリング仕様書

## 概要

APIは統一されたエラーレスポンス形式を採用し、クライアントが一貫した方法でエラーを処理できるようにしています。

---

## エラーレスポンス形式

### 基本構造

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { },
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### フィールド説明

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| code | string | Yes | 機械可読なエラーコード |
| message | string | Yes | 人間可読なエラーメッセージ |
| details | object | No | 追加情報（バリデーションエラー詳細など） |
| request_id | string | Yes | リクエスト追跡用UUID |
| timestamp | string | Yes | エラー発生日時（ISO 8601形式） |

---

## エラーコード一覧

### 認証・認可エラー

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| AUTHENTICATION_FAILED | 401 | 認証に失敗 |
| TOKEN_EXPIRED | 401 | トークンが期限切れ |
| TOKEN_REVOKED | 401 | トークンが無効化済み |
| INVALID_TOKEN | 401 | トークンが無効 |
| AUTHORIZATION_FAILED | 403 | 権限なし |

### バリデーションエラー

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| VALIDATION_FAILED | 422 | 入力値のバリデーション失敗 |
| PARAMETER_MISSING | 400 | 必須パラメータが不足 |

### リソースエラー

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| RESOURCE_NOT_FOUND | 404 | リソースが見つからない |
| DUPLICATE_RESOURCE | 409 | リソースが重複 |

### ビジネスロジックエラー

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| INVALID_STATE_TRANSITION | 422 | 無効な状態遷移 |
| RESOURCE_LIMIT_EXCEEDED | 422 | リソース上限超過 |
| EDIT_TIME_EXPIRED | 422 | 編集可能時間を超過 |

### システムエラー

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| RATE_LIMIT_EXCEEDED | 429 | レート制限超過 |
| INTERNAL_ERROR | 500 | サーバー内部エラー |

---

## HTTPステータスコード対応表

| ステータスコード | 意味 | 使用場面 |
|-----------------|------|---------|
| 200 | OK | 正常処理完了 |
| 201 | Created | リソース作成成功 |
| 204 | No Content | 削除成功（レスポンスボディなし） |
| 400 | Bad Request | パラメータエラー |
| 401 | Unauthorized | 認証失敗 |
| 403 | Forbidden | 権限なし |
| 404 | Not Found | リソース未検出 |
| 409 | Conflict | リソース競合 |
| 422 | Unprocessable Entity | バリデーション失敗 |
| 429 | Too Many Requests | レート制限 |
| 500 | Internal Server Error | サーバーエラー |

---

## エラー詳細例

### AUTHENTICATION_FAILED (401)

認証に失敗した場合

```json
{
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "Invalid email or password",
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### TOKEN_EXPIRED (401)

JWTトークンが期限切れの場合

```json
{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Token has expired. Please sign in again.",
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### AUTHORIZATION_FAILED (403)

権限がない操作を試みた場合

```json
{
  "error": {
    "code": "AUTHORIZATION_FAILED",
    "message": "You are not authorized to perform this action",
    "details": {
      "resource": "Comment",
      "action": "update"
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### RESOURCE_NOT_FOUND (404)

リソースが見つからない場合

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested resource was not found",
    "details": {
      "resource": "Todo",
      "id": 123
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### VALIDATION_FAILED (422)

バリデーションエラーの場合

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed. Please check your input.",
    "details": {
      "validation_errors": {
        "title": ["can't be blank"],
        "email": ["has already been taken", "is invalid"],
        "password": ["is too short (minimum is 6 characters)"]
      }
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### PARAMETER_MISSING (400)

必須パラメータが不足している場合

```json
{
  "error": {
    "code": "PARAMETER_MISSING",
    "message": "Required parameter is missing",
    "details": {
      "parameter": "todo"
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### DUPLICATE_RESOURCE (409)

リソースが重複している場合

```json
{
  "error": {
    "code": "DUPLICATE_RESOURCE",
    "message": "Resource already exists",
    "details": {
      "resource": "Category",
      "field": "name",
      "value": "Work"
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### EDIT_TIME_EXPIRED (422)

コメントの編集可能時間（15分）を超過した場合

```json
{
  "error": {
    "code": "EDIT_TIME_EXPIRED",
    "message": "Comment can no longer be edited. Edit time limit (15 minutes) has expired.",
    "details": {
      "resource": "Comment",
      "id": 123,
      "created_at": "2025-01-15T10:00:00Z",
      "edit_limit_minutes": 15
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### RATE_LIMIT_EXCEEDED (429)

レート制限を超過した場合

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "details": {
      "limit": 100,
      "remaining": 0,
      "reset_at": "2025-01-15T11:00:00Z"
    },
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

### INTERNAL_ERROR (500)

サーバー内部エラーの場合

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred. Please try again later.",
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

**注意**: 本番環境ではセキュリティのため、詳細なエラー情報を含めない。

---

## エラーハンドリング実装パターン

### カスタムエラークラス階層

```
ApiError (Base)
├── AuthenticationError (401)
├── AuthorizationError (403)
├── NotFoundError (404)
├── ValidationError (422)
├── RateLimitError (429)
└── BusinessLogicError (422)
    ├── DuplicateResourceError
    ├── InvalidStateTransitionError
    └── ResourceLimitExceededError
```

### エラークラス定義例（擬似コード）

```python
class ApiError(Exception):
    def __init__(self, message, code="API_ERROR", status=400, details=None):
        self.message = message
        self.code = code
        self.status = status
        self.details = details or {}

class AuthenticationError(ApiError):
    def __init__(self, message="Authentication failed"):
        super().__init__(
            message=message,
            code="AUTHENTICATION_FAILED",
            status=401
        )

class AuthorizationError(ApiError):
    def __init__(self, message="Not authorized", resource=None, action=None):
        super().__init__(
            message=message,
            code="AUTHORIZATION_FAILED",
            status=403,
            details={"resource": resource, "action": action}
        )

class NotFoundError(ApiError):
    def __init__(self, resource, id):
        super().__init__(
            message=f"{resource} not found",
            code="RESOURCE_NOT_FOUND",
            status=404,
            details={"resource": resource, "id": id}
        )

class ValidationError(ApiError):
    def __init__(self, errors):
        super().__init__(
            message="Validation failed. Please check your input.",
            code="VALIDATION_FAILED",
            status=422,
            details={"validation_errors": errors}
        )

class RateLimitError(ApiError):
    def __init__(self, limit, remaining, reset_at):
        super().__init__(
            message="Too many requests. Please try again later.",
            code="RATE_LIMIT_EXCEEDED",
            status=429,
            details={
                "limit": limit,
                "remaining": remaining,
                "reset_at": reset_at
            }
        )
```

### グローバルエラーハンドラー（擬似コード）

```python
def error_handler(error, request):
    request_id = request.headers.get("X-Request-Id") or generate_uuid()
    timestamp = datetime.now().isoformat()

    if isinstance(error, ApiError):
        return json_response({
            "error": {
                "code": error.code,
                "message": error.message,
                "details": error.details,
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=error.status)

    # ActiveRecord::RecordNotFound相当
    if isinstance(error, RecordNotFoundError):
        return json_response({
            "error": {
                "code": "RESOURCE_NOT_FOUND",
                "message": "The requested resource was not found",
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=404)

    # バリデーションエラー
    if isinstance(error, RecordInvalidError):
        return json_response({
            "error": {
                "code": "VALIDATION_FAILED",
                "message": "Validation failed. Please check your input.",
                "details": {
                    "validation_errors": format_validation_errors(error.errors)
                },
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=422)

    # JWTエラー
    if isinstance(error, (JWTDecodeError, JWTExpiredError)):
        return json_response({
            "error": {
                "code": "AUTHENTICATION_FAILED",
                "message": "Invalid or expired token",
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=401)

    # その他のエラー
    log_error(error)  # エラーログ記録

    # 本番環境では詳細を隠す
    if is_production():
        return json_response({
            "error": {
                "code": "INTERNAL_ERROR",
                "message": "An unexpected error occurred",
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=500)
    else:
        return json_response({
            "error": {
                "code": "INTERNAL_ERROR",
                "message": str(error),
                "details": {"trace": format_traceback(error)},
                "request_id": request_id,
                "timestamp": timestamp
            }
        }, status=500)
```

---

## クライアント側のエラー処理

### フロントエンドでの処理例

```typescript
interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
    request_id: string;
    timestamp: string;
  };
}

async function apiRequest<T>(url: string, options: RequestInit): Promise<T> {
  const response = await fetch(url, options);

  if (!response.ok) {
    const error: ApiError = await response.json();

    switch (error.error.code) {
      case 'AUTHENTICATION_FAILED':
      case 'TOKEN_EXPIRED':
      case 'TOKEN_REVOKED':
        // ログイン画面にリダイレクト
        redirectToLogin();
        break;

      case 'VALIDATION_FAILED':
        // バリデーションエラーを表示
        const errors = error.error.details?.validation_errors;
        showValidationErrors(errors);
        break;

      case 'RESOURCE_NOT_FOUND':
        // 404ページを表示
        show404Page();
        break;

      case 'RATE_LIMIT_EXCEEDED':
        // リトライ案内を表示
        const resetAt = error.error.details?.reset_at;
        showRateLimitMessage(resetAt);
        break;

      default:
        // 一般的なエラーメッセージを表示
        showErrorToast(error.error.message);
    }

    throw error;
  }

  return response.json();
}
```

---

## ログ出力

### エラーログ形式

```json
{
  "level": "error",
  "timestamp": "2025-01-15T10:30:00Z",
  "request_id": "abc123",
  "method": "POST",
  "path": "/api/v1/todos",
  "user_id": 1,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "class": "ValidationError"
  },
  "params": {
    "todo": {
      "title": "",
      "priority": "invalid"
    }
  }
}
```

### センシティブ情報のフィルタリング

以下のパラメータはログに出力しない：
- password
- password_confirmation
- token
- authorization
- secret
- api_key

---

## テストケース

### 認証エラー

1. トークンなし → 401 AUTHENTICATION_FAILED
2. 無効なトークン → 401 INVALID_TOKEN
3. 期限切れトークン → 401 TOKEN_EXPIRED
4. 無効化されたトークン → 401 TOKEN_REVOKED

### 認可エラー

1. 他ユーザーのリソースへのアクセス → 404 RESOURCE_NOT_FOUND
2. コメント編集権限なし → 403 AUTHORIZATION_FAILED

### バリデーションエラー

1. 空のtitle → 422 VALIDATION_FAILED
2. 重複したemail → 422 VALIDATION_FAILED
3. 無効なカラーコード → 422 VALIDATION_FAILED

### リソースエラー

1. 存在しないTodo取得 → 404 RESOURCE_NOT_FOUND
2. 重複したカテゴリ名 → 409 DUPLICATE_RESOURCE

### ビジネスロジックエラー

1. 15分経過後のコメント編集 → 422 EDIT_TIME_EXPIRED
