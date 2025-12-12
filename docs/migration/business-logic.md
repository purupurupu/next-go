# ビジネスロジック仕様書

## 概要

このドキュメントでは、アプリケーションの各モデルに実装されているバリデーションルール、コールバック、ビジネスロジックを詳細に記述します。

---

## 1. User モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| name | 必須 | can't be blank |
| name | 2-50文字 | is too short (minimum is 2 characters) / is too long (maximum is 50 characters) |
| email | 必須 | can't be blank |
| email | ユニーク（大文字小文字区別なし） | has already been taken |
| email | メール形式 | is invalid |
| password | 必須（作成時） | can't be blank |
| password | 6文字以上 | is too short (minimum is 6 characters) |
| password_confirmation | passwordと一致 | doesn't match Password |

### パスワードハッシュ化

- **アルゴリズム**: bcrypt
- **コスト**: 12（Deviseデフォルト）
- **保存先**: `encrypted_password`カラム

---

## 2. Todo モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| title | 必須 | can't be blank |
| completed | boolean型のみ | is not included in the list |
| due_date | 過去日付禁止（設定時のみ） | can't be in the past |
| files | ファイルサイズ10MB以下 | is too large (maximum is 10 MB) |
| files | 許可MIMEタイプのみ | has an invalid content type |

### 許可されるファイルタイプ

```ruby
ALLOWED_CONTENT_TYPES = [
  # 画像
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  # ドキュメント
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  # テキスト
  'text/plain',
  'text/csv'
]
```

### Enum定義

**priority（優先度）**
| 値 | 名前 | 日本語 |
|----|------|--------|
| 0 | low | 低 |
| 1 | medium | 中（デフォルト） |
| 2 | high | 高 |

**status（ステータス）**
| 値 | 名前 | 日本語 |
|----|------|--------|
| 0 | pending | 未着手（デフォルト） |
| 1 | in_progress | 進行中 |
| 2 | completed | 完了 |

### コールバック

#### before_create: set_position

新規Todo作成時に自動的に最後の位置を設定。

```python
def set_position(todo):
    max_position = db.todos.filter(user_id=todo.user_id).max("position") or 0
    todo.position = max_position + 1
```

#### after_create: record_creation

Todo作成時に初期履歴を記録。

```python
def record_creation(todo):
    TodoHistory.create(
        todo_id=todo.id,
        user_id=todo.user_id,  # current_userから取得
        field_name="todo",
        old_value=None,
        new_value=todo.title,
        action="created"
    )
```

#### around_update: track_changes_with_user

Todo更新時に変更内容を履歴として記録。

**追跡対象フィールド**:
- title
- status
- priority
- due_date
- completed
- category_id
- description

```python
TRACKED_FIELDS = ['title', 'status', 'priority', 'due_date', 'completed', 'category_id', 'description']

def track_changes(todo, current_user):
    changes = []

    for field in TRACKED_FIELDS:
        old_value = todo._original_values.get(field)
        new_value = getattr(todo, field)

        if old_value != new_value:
            action = determine_action(field, old_value, new_value)
            changes.append({
                'field_name': field,
                'old_value': str(old_value) if old_value is not None else None,
                'new_value': str(new_value) if new_value is not None else None,
                'action': action
            })

    # 変更を保存後に履歴を記録
    for change in changes:
        TodoHistory.create(
            todo_id=todo.id,
            user_id=current_user.id,
            **change
        )

def determine_action(field, old_value, new_value):
    if field == 'status':
        return 'status_changed'
    elif field == 'priority':
        return 'priority_changed'
    else:
        return 'updated'
```

### スコープ

```python
# position昇順でソート
def ordered(query):
    return query.order_by("position", "asc")
```

---

## 3. Category モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| name | 必須 | can't be blank |
| name | 50文字以下 | is too long (maximum is 50 characters) |
| name | ユーザー内ユニーク（大文字小文字区別なし） | has already been taken |
| color | 必須 | can't be blank |
| color | HEX形式（#RGB または #RRGGBB） | must be a valid hex color |

### カラーコード検証パターン

```regex
^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$
```

有効例: `#FF5733`, `#fff`, `#ABC123`
無効例: `FF5733`, `#FFFF`, `red`

### コールバック

#### before_validation: normalize_color

カラーコードを大文字に正規化。

```python
def normalize_color(category):
    if category.color:
        category.color = category.color.upper()
```

### カウンターキャッシュ

`todos_count`フィールドは、関連するTodoの数を自動的に追跡。

```python
# Todoが作成されたとき
category.todos_count += 1

# Todoが削除されたとき
category.todos_count -= 1

# Todoのcategory_idが変更されたとき
old_category.todos_count -= 1
new_category.todos_count += 1
```

### 削除時の動作

カテゴリ削除時、関連するTodoの`category_id`は`NULL`に設定（nullify）。

```python
# dependent: :nullify
def on_delete_category(category):
    db.todos.filter(category_id=category.id).update(category_id=None)
```

---

## 4. Tag モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| name | 必須 | can't be blank |
| name | 30文字以下 | is too long (maximum is 30 characters) |
| name | ユーザー内ユニーク（大文字小文字区別なし） | has already been taken |
| color | HEX形式（オプション） | must be a valid hex color |

### コールバック

#### before_validation: normalize_name

タグ名を小文字に正規化し、前後の空白を除去。

```python
def normalize_name(tag):
    if tag.name:
        tag.name = tag.name.strip().lower()
```

#### before_validation: normalize_color

カラーコードを大文字に正規化（colorが設定されている場合）。

```python
def normalize_color(tag):
    if tag.color:
        tag.color = tag.color.upper()
```

### スコープ

```python
# 名前順でソート
def ordered(query):
    return query.order_by("name", "asc")
```

---

## 5. TodoTag モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| todo_id + tag_id | ユニーク | has already been taken |

### 重要な制約

同じTodoに同じTagを2回以上紐付けることはできない。

---

## 6. Comment モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| content | 必須 | can't be blank |
| content | 1000文字以下 | is too long (maximum is 1000 characters) |

### スコープ

```python
# アクティブなコメントのみ（デフォルト）
def active(query):
    return query.filter(deleted_at=None)

# 削除済みコメント
def deleted(query):
    return query.filter(deleted_at__isnot=None)

# 作成日時昇順
def chronological(query):
    return query.order_by("created_at", "asc")

# 作成日時降順
def recent(query):
    return query.order_by("created_at", "desc")
```

### ソフトデリート

コメントは物理削除ではなく、`deleted_at`に日時を設定する論理削除を採用。

```python
def soft_delete(comment):
    comment.deleted_at = datetime.now()
    comment.save()

def restore(comment):
    comment.deleted_at = None
    comment.save()
```

### 編集可能判定

コメントは以下の条件を満たす場合のみ編集可能：

1. 削除されていない（`deleted_at`がNULL）
2. 作成から15分以内

```python
EDIT_TIME_LIMIT = timedelta(minutes=15)

def is_editable(comment):
    if comment.deleted_at is not None:
        return False

    time_since_creation = datetime.now() - comment.created_at
    return time_since_creation <= EDIT_TIME_LIMIT
```

### 所有者判定

```python
def owned_by(comment, user):
    return comment.user_id == user.id
```

### ポリモーフィック関連

コメントは`commentable_type`と`commentable_id`でポリモーフィック関連を実現。

```python
# 現在の関連先
commentable_type = "Todo"  # モデル名
commentable_id = 123       # レコードID
```

将来的に他のモデル（例: Project, Issue）にもコメントを追加可能。

---

## 7. TodoHistory モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| field_name | 必須 | can't be blank |
| action | 必須 | can't be blank |

### Enum定義

**action（アクション）**
| 値 | 名前 | 説明 |
|----|------|------|
| 0 | created | 作成 |
| 1 | updated | 更新 |
| 2 | deleted | 削除 |
| 3 | status_changed | ステータス変更 |
| 4 | priority_changed | 優先度変更 |

### スコープ

```python
# デフォルト: 作成日時降順（最新順）
def default_scope(query):
    return query.order_by("created_at", "desc")

# 最新N件
def recent(query, limit=10):
    return query.limit(limit)

# 特定フィールドの履歴
def for_field(query, field):
    return query.filter(field_name=field)

# 特定ユーザーの変更
def by_user(query, user):
    return query.filter(user_id=user.id)
```

### 可読化メソッド

履歴を日本語で可読化して返す。

```python
def human_readable_change(history):
    field = history.field_name
    old_val = history.old_value
    new_val = history.new_value

    if field == "status":
        old_status = translate_status(old_val)
        new_status = translate_status(new_val)
        return f"ステータスを「{old_status}」から「{new_status}」に変更"

    elif field == "priority":
        old_priority = translate_priority(old_val)
        new_priority = translate_priority(new_val)
        return f"優先度を「{old_priority}」から「{new_priority}」に変更"

    elif field == "due_date":
        return format_due_date_change(old_val, new_val)

    elif field == "completed":
        if new_val == "true":
            return "タスクを完了にしました"
        else:
            return "タスクを未完了に戻しました"

    elif field == "category_id":
        return format_category_change(old_val, new_val)

    elif field == "title":
        return f"タイトルを「{old_val}」から「{new_val}」に変更"

    elif field == "description":
        return "説明を更新しました"

    elif field == "todo" and history.action == "created":
        return f"タスク「{new_val}」を作成しました"

    else:
        return f"{field}を更新しました"

def translate_status(status):
    translations = {
        "pending": "未着手",
        "in_progress": "進行中",
        "completed": "完了",
        "0": "未着手",
        "1": "進行中",
        "2": "完了"
    }
    return translations.get(str(status), status)

def translate_priority(priority):
    translations = {
        "low": "低",
        "medium": "中",
        "high": "高",
        "0": "低",
        "1": "中",
        "2": "高"
    }
    return translations.get(str(priority), priority)

def format_due_date_change(old_val, new_val):
    if old_val and new_val:
        return f"期限を{format_date(old_val)}から{format_date(new_val)}に変更"
    elif new_val:
        return f"期限を{format_date(new_val)}に設定"
    else:
        return "期限を削除しました"

def format_date(date_str):
    # YYYY-MM-DD形式をYYYY年MM月DD日形式に変換
    date = datetime.strptime(date_str, "%Y-%m-%d")
    return date.strftime("%Y年%m月%d日")
```

---

## 8. Note モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| title | 150文字以下（オプション） | is too long (maximum is 150 characters) |
| body_md | 100,000文字以下（オプション） | is too long (maximum is 100000 characters) |

### コールバック

#### before_save: set_body_plain

Markdown本文からプレーンテキストを生成（検索用）。

```python
def set_body_plain(note):
    if note.body_md_changed:
        # Markdownタグを除去してプレーンテキストに変換
        note.body_plain = strip_markdown(note.body_md) if note.body_md else None
```

#### before_save: touch_last_edited_at

タイトルまたは本文が変更された場合、`last_edited_at`を更新。

```python
def touch_last_edited_at(note):
    if note.title_changed or note.body_md_changed:
        note.last_edited_at = datetime.now()
```

#### after_create_commit: record_initial_revision

Note作成後に初期リビジョンを記録。

```python
def record_initial_revision(note):
    NoteRevision.create(
        note_id=note.id,
        user_id=note.user_id,
        title=note.title,
        body_md=note.body_md
    )
```

#### after_update_commit: record_revision_if_changed

Note更新後にリビジョンを記録（本文が変更された場合）。

```python
def record_revision_if_changed(note):
    if note.body_md_changed:
        NoteRevision.create(
            note_id=note.id,
            user_id=note.current_user.id,
            title=note.title,
            body_md=note.body_md
        )
        prune_revisions(note)

# 最新50件のみ保持
def prune_revisions(note):
    revisions = NoteRevision.filter(note_id=note.id).order_by("created_at", "desc")
    if len(revisions) > 50:
        for revision in revisions[50:]:
            revision.delete()
```

### スコープ

```python
# ゴミ箱に入っていないNote
def not_trashed(query):
    return query.filter(trashed_at=None)

# ゴミ箱のNote
def trashed(query):
    return query.filter(trashed_at__isnot=None)

# アーカイブ済みNote
def archived(query):
    return query.filter(archived_at__isnot=None)

# アクティブなNote（ゴミ箱でもアーカイブでもない）
def active(query):
    return query.filter(trashed_at=None, archived_at=None)

# ピン留め優先、更新日時降順
def pinned_first(query):
    return query.order_by("pinned", "desc").order_by("updated_at", "desc")
```

---

## 9. NoteRevision モデル

### バリデーション

| フィールド | ルール | エラーメッセージ |
|-----------|--------|-----------------|
| title | 150文字以下（オプション） | is too long (maximum is 150 characters) |

### リビジョン管理

- Note1件につき最大50件のリビジョンを保持
- 50件を超えた場合、古いリビジョンから削除

---

## 10. タグ検証（Todo更新時）

Todoにタグを割り当てる際、他のユーザーのタグを使用できないよう検証。

```python
def validate_tag_ids(todo, tag_ids, current_user):
    if not tag_ids:
        return []

    # 現在のユーザーのタグのみ取得
    valid_tags = db.tags.filter(
        user_id=current_user.id,
        id__in=tag_ids
    )

    # 有効なIDのみ返す
    return [tag.id for tag in valid_tags]
```

---

## 11. 検索サービス

### TodoSearchService

複雑な検索・フィルタリング・ソート・ページネーションを一元管理。

#### フィルタリング機能

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| q | タイトル・説明での検索 | "meeting" |
| status | ステータスフィルタ（複数可） | "pending,in_progress" |
| priority | 優先度フィルタ | "high" |
| category_id | カテゴリフィルタ（-1でカテゴリなし） | "1" or "-1" |
| tag_ids | タグフィルタ（複数可） | "1,2,3" |
| tag_mode | タグ検索モード | "any"(OR) / "all"(AND) |
| due_date_from | 期限開始日 | "2025-01-01" |
| due_date_to | 期限終了日 | "2025-12-31" |

#### ソート機能

| sort_by | 説明 |
|---------|------|
| created_at | 作成日時 |
| updated_at | 更新日時 |
| due_date | 期限（NULLは最後） |
| title | タイトル |
| priority | 優先度 |
| status | ステータス |

#### ページネーション

| パラメータ | デフォルト | 範囲 |
|-----------|-----------|------|
| page | 1 | 1以上 |
| per_page | 20 | 1-100 |

---

## 12. ユーザースコープ

全てのリソースはユーザーにスコープされており、他のユーザーのデータにはアクセスできない。

```python
# 全てのクエリでuser_idをフィルタ
def get_todos(current_user):
    return db.todos.filter(user_id=current_user.id)

def get_todo(current_user, todo_id):
    todo = db.todos.filter(user_id=current_user.id, id=todo_id).first()
    if not todo:
        raise NotFoundError("Todo", todo_id)
    return todo
```

これにより、他のユーザーのデータを取得・変更しようとした場合は404 Not Foundが返される。
