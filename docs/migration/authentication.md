# 認証仕様書

## 概要

このシステムはJWT（JSON Web Token）ベースの認証を採用しています。
現在のRails実装ではDevise + devise-jwtを使用していますが、移行先では任意のJWTライブラリで実装可能です。

## 認証フロー

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant DB

    Note over Client,DB: ユーザー登録フロー
    Client->>API: POST /auth/sign_up
    API->>DB: ユーザー作成
    DB-->>API: ユーザー情報
    API->>API: JWT生成
    API-->>Client: 200 OK + Authorization: Bearer <token>

    Note over Client,DB: ログインフロー
    Client->>API: POST /auth/sign_in
    API->>DB: ユーザー検証
    DB-->>API: ユーザー情報
    API->>API: パスワード検証
    API->>API: JWT生成
    API-->>Client: 200 OK + Authorization: Bearer <token>

    Note over Client,DB: API呼び出しフロー
    Client->>API: GET /api/v1/todos (Authorization: Bearer <token>)
    API->>API: JWT検証
    API->>DB: jwt_denylistsチェック
    DB-->>API: 有効性確認
    API->>DB: データ取得
    DB-->>API: データ
    API-->>Client: 200 OK + データ

    Note over Client,DB: ログアウトフロー
    Client->>API: DELETE /auth/sign_out (Authorization: Bearer <token>)
    API->>API: JWT検証
    API->>DB: jwt_denylistsに追加
    DB-->>API: 完了
    API-->>Client: 200 OK
```

---

## JWTトークン構造

### ヘッダー
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### ペイロード
```json
{
  "sub": "1",           // ユーザーID
  "jti": "abc123...",   // JWT ID（ユニーク識別子）
  "scp": "user",        // スコープ
  "aud": null,          // オーディエンス（オプション）
  "iat": 1705312200,    // 発行日時（Unix timestamp）
  "exp": 1705398600     // 有効期限（Unix timestamp）
}
```

### 署名
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret_key
)
```

---

## トークン設定

| 項目 | 値 | 説明 |
|------|-----|------|
| アルゴリズム | HS256 | HMAC SHA-256 |
| 有効期限 | 24時間 | `exp` claimで管理 |
| 発行者 | - | 設定により変更可能 |
| Secret Key | Rails credentials | `devise_jwt_secret_key` |

### Secret Key管理

**Rails現状**:
```ruby
# config/initializers/devise.rb
config.jwt do |jwt|
  jwt.secret = Rails.application.credentials.devise_jwt_secret_key!
  jwt.dispatch_requests = [
    ['POST', %r{^/auth/sign_in$}]
  ]
  jwt.revocation_requests = [
    ['DELETE', %r{^/auth/sign_out$}]
  ]
  jwt.expiration_time = 1.day.to_i
end
```

**移行先での推奨**:
- 環境変数 `JWT_SECRET` で管理
- 256ビット以上のランダム文字列を使用
- 本番環境では必ずセキュアな値を設定

---

## 認証エンドポイント詳細

### POST /auth/sign_up

**ユーザー登録**

リクエスト:
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123",
    "password_confirmation": "password123",
    "name": "John Doe"
  }
}
```

バリデーション:
- `email`: 必須、ユニーク、メール形式
- `password`: 必須、6文字以上
- `password_confirmation`: `password`と一致
- `name`: 必須、2-50文字

成功レスポンス (201):
```json
{
  "status": {
    "code": 201,
    "message": "Signed up successfully."
  },
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

レスポンスヘッダー:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

エラーレスポンス (422):
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "details": {
      "validation_errors": {
        "email": ["has already been taken"],
        "password": ["is too short (minimum is 6 characters)"]
      }
    }
  }
}
```

---

### POST /auth/sign_in

**ユーザーログイン**

リクエスト:
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123"
  }
}
```

成功レスポンス (200):
```json
{
  "status": {
    "code": 200,
    "message": "Logged in successfully."
  },
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

レスポンスヘッダー:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

エラーレスポンス (401):
```json
{
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "Invalid email or password"
  }
}
```

---

### DELETE /auth/sign_out

**ログアウト（トークン無効化）**

リクエストヘッダー:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

成功レスポンス (200):
```json
{
  "status": {
    "code": 200,
    "message": "Logged out successfully."
  }
}
```

処理内容:
1. JWTからjtiを抽出
2. `jwt_denylists`テーブルにjtiとexpを保存
3. 以降、同じトークンは無効として扱う

---

## トークン無効化（Revocation）

### jwt_denylistsテーブル

無効化されたトークンを管理するテーブル。

```sql
CREATE TABLE jwt_denylists (
  id BIGSERIAL PRIMARY KEY,
  jti VARCHAR NOT NULL,
  exp TIMESTAMP,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE INDEX index_jwt_denylists_on_jti ON jwt_denylists(jti);
```

### 無効化チェックフロー

```mermaid
flowchart TD
    A[リクエスト受信] --> B{Authorizationヘッダー存在?}
    B -->|No| C[401 Unauthorized]
    B -->|Yes| D[JWT抽出]
    D --> E{署名検証}
    E -->|Invalid| C
    E -->|Valid| F{有効期限チェック}
    F -->|Expired| C
    F -->|Valid| G{jwt_denylistsにjti存在?}
    G -->|Yes| C
    G -->|No| H[認証成功]
    H --> I[current_user設定]
```

### 期限切れトークンのクリーンアップ

```sql
-- 定期実行推奨（1日1回など）
DELETE FROM jwt_denylists WHERE exp < NOW();
```

---

## パスワードハッシュ化

### 現在の実装（Devise/bcrypt）

```ruby
# パスワードハッシュ化
BCrypt::Password.create(password, cost: 12)

# パスワード検証
BCrypt::Password.new(encrypted_password) == password
```

### 移行先での推奨

**Rust**:
```rust
// bcryptクレート使用
use bcrypt::{hash, verify, DEFAULT_COST};

let hashed = hash(password, DEFAULT_COST)?;
let valid = verify(password, &hashed)?;
```

**Go**:
```go
// golang.org/x/crypto/bcrypt使用
import "golang.org/x/crypto/bcrypt"

hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword(hashed, []byte(password))
```

---

## 認証ミドルウェア

### 処理フロー

1. `Authorization`ヘッダーからトークンを抽出
2. `Bearer `プレフィックスを除去
3. JWT署名を検証
4. ペイロードから`sub`（ユーザーID）を取得
5. `jwt_denylists`でトークン無効化をチェック
6. ユーザーをDBから取得
7. リクエストコンテキストに`current_user`を設定

### 擬似コード

```python
def authenticate(request):
    # 1. ヘッダー取得
    auth_header = request.headers.get("Authorization")
    if not auth_header:
        raise AuthenticationError("Missing authorization header")

    # 2. トークン抽出
    if not auth_header.startswith("Bearer "):
        raise AuthenticationError("Invalid authorization format")
    token = auth_header[7:]  # "Bearer "を除去

    # 3. JWT検証
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
    except jwt.ExpiredSignatureError:
        raise AuthenticationError("Token has expired")
    except jwt.InvalidTokenError:
        raise AuthenticationError("Invalid token")

    # 4. ユーザーID取得
    user_id = payload.get("sub")
    jti = payload.get("jti")

    # 5. 無効化チェック
    if db.jwt_denylists.exists(jti=jti):
        raise AuthenticationError("Token has been revoked")

    # 6. ユーザー取得
    user = db.users.find(id=user_id)
    if not user:
        raise AuthenticationError("User not found")

    # 7. コンテキストに設定
    request.context.current_user = user
    return user
```

---

## セキュリティ考慮事項

### 必須事項

1. **HTTPS必須**: 本番環境ではTLS/SSL必須
2. **Secret Keyの管理**: 環境変数で管理、ソースコードにハードコードしない
3. **トークン有効期限**: 24時間以内を推奨
4. **パスワードポリシー**: 最低6文字、より強力なポリシーを推奨

### 推奨事項

1. **Rate Limiting**: ログインエンドポイントにレート制限を設定
2. **アカウントロック**: 連続ログイン失敗時のアカウントロック
3. **Refresh Token**: 長期セッションが必要な場合は検討
4. **監査ログ**: ログイン/ログアウトの記録

### CORS設定

```yaml
# 許可するオリジン
origins:
  - http://localhost:3000  # 開発環境フロントエンド

# 許可するメソッド
methods:
  - GET
  - POST
  - PUT
  - PATCH
  - DELETE
  - OPTIONS
  - HEAD

# 許可するヘッダー
headers:
  - Authorization
  - Content-Type

# 公開するヘッダー
expose:
  - Authorization

# クレデンシャル許可
credentials: true
```

---

## フロントエンド連携

### トークンの保存

```javascript
// ログイン成功時
const response = await fetch('/auth/sign_in', { ... });
const token = response.headers.get('Authorization');
localStorage.setItem('token', token);
```

### APIリクエスト

```javascript
const token = localStorage.getItem('token');
const response = await fetch('/api/v1/todos', {
  headers: {
    'Authorization': token,
    'Content-Type': 'application/json'
  }
});
```

### ログアウト

```javascript
const token = localStorage.getItem('token');
await fetch('/auth/sign_out', {
  method: 'DELETE',
  headers: { 'Authorization': token }
});
localStorage.removeItem('token');
```

---

## テストケース

### 認証が必要なエンドポイントのテスト

1. トークンなしでリクエスト → 401
2. 無効なトークンでリクエスト → 401
3. 期限切れトークンでリクエスト → 401
4. 無効化されたトークンでリクエスト → 401
5. 有効なトークンでリクエスト → 200/データ返却

### ログインテスト

1. 正しい認証情報 → 200 + トークン
2. 間違ったパスワード → 401
3. 存在しないメール → 401
4. 空のリクエスト → 400/422

### 登録テスト

1. 有効なデータ → 201 + トークン
2. 重複メール → 422
3. パスワード不一致 → 422
4. 短すぎるパスワード → 422
