# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.0].define(version: 2025_10_17_090001) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "pg_catalog.plpgsql"

  create_table "active_storage_attachments", force: :cascade do |t|
    t.string "name", null: false
    t.string "record_type", null: false
    t.bigint "record_id", null: false
    t.bigint "blob_id", null: false
    t.datetime "created_at", null: false
    t.index ["blob_id"], name: "index_active_storage_attachments_on_blob_id"
    t.index ["record_type", "record_id", "name", "blob_id"], name: "index_active_storage_attachments_uniqueness", unique: true
  end

  create_table "active_storage_blobs", force: :cascade do |t|
    t.string "key", null: false
    t.string "filename", null: false
    t.string "content_type"
    t.text "metadata"
    t.string "service_name", null: false
    t.bigint "byte_size", null: false
    t.string "checksum"
    t.datetime "created_at", null: false
    t.index ["key"], name: "index_active_storage_blobs_on_key", unique: true
  end

  create_table "active_storage_variant_records", force: :cascade do |t|
    t.bigint "blob_id", null: false
    t.string "variation_digest", null: false
    t.index ["blob_id", "variation_digest"], name: "index_active_storage_variant_records_uniqueness", unique: true
  end

  create_table "categories", force: :cascade do |t|
    t.string "name", null: false
    t.string "color", default: "#6B7280", null: false
    t.bigint "user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "todos_count", default: 0, null: false
    t.index ["user_id", "name"], name: "index_categories_on_user_id_and_name", unique: true
    t.index ["user_id"], name: "index_categories_on_user_id"
  end

  create_table "comments", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "commentable_type", null: false
    t.bigint "commentable_id", null: false
    t.text "content", null: false
    t.datetime "deleted_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["commentable_type", "commentable_id", "deleted_at"], name: "index_comments_on_commentable_and_deleted_at"
    t.index ["commentable_type", "commentable_id"], name: "index_comments_on_commentable"
    t.index ["deleted_at"], name: "index_comments_on_deleted_at"
    t.index ["user_id"], name: "index_comments_on_user_id"
  end

  create_table "jwt_denylists", force: :cascade do |t|
    t.string "jti"
    t.datetime "exp"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["jti"], name: "index_jwt_denylists_on_jti"
  end

  create_table "note_revisions", force: :cascade do |t|
    t.bigint "note_id", null: false
    t.bigint "user_id", null: false
    t.string "title", limit: 150
    t.text "body_md"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["note_id", "created_at"], name: "index_note_revisions_on_note_id_and_created_at"
    t.index ["note_id"], name: "index_note_revisions_on_note_id"
    t.index ["user_id"], name: "index_note_revisions_on_user_id"
  end

  create_table "notes", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.string "title", limit: 150
    t.text "body_md"
    t.text "body_plain"
    t.boolean "pinned", default: false, null: false
    t.datetime "archived_at"
    t.datetime "trashed_at"
    t.datetime "last_edited_at", default: -> { "CURRENT_TIMESTAMP" }, null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["archived_at"], name: "index_notes_on_archived_at"
    t.index ["body_plain"], name: "index_notes_on_body_plain"
    t.index ["last_edited_at"], name: "index_notes_on_last_edited_at"
    t.index ["pinned"], name: "index_notes_on_pinned"
    t.index ["trashed_at"], name: "index_notes_on_trashed_at"
    t.index ["user_id", "archived_at"], name: "index_notes_on_user_id_and_archived_at"
    t.index ["user_id", "last_edited_at"], name: "index_notes_on_user_id_and_last_edited_at"
    t.index ["user_id", "pinned"], name: "index_notes_on_user_id_and_pinned"
    t.index ["user_id", "trashed_at"], name: "index_notes_on_user_id_and_trashed_at"
    t.index ["user_id"], name: "index_notes_on_user_id"
  end

  create_table "tags", force: :cascade do |t|
    t.string "name", null: false
    t.bigint "user_id", null: false
    t.string "color", default: "#6B7280"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id", "name"], name: "index_tags_on_user_id_and_name", unique: true
    t.index ["user_id"], name: "index_tags_on_user_id"
  end

  create_table "todo_histories", force: :cascade do |t|
    t.bigint "todo_id", null: false
    t.bigint "user_id", null: false
    t.string "field_name", null: false
    t.text "old_value"
    t.text "new_value"
    t.integer "action", default: 0, null: false
    t.datetime "created_at", null: false
    t.index ["field_name"], name: "index_todo_histories_on_field_name"
    t.index ["todo_id", "created_at"], name: "index_todo_histories_on_todo_id_and_created_at"
    t.index ["todo_id"], name: "index_todo_histories_on_todo_id"
    t.index ["user_id"], name: "index_todo_histories_on_user_id"
  end

  create_table "todo_tags", force: :cascade do |t|
    t.bigint "todo_id", null: false
    t.bigint "tag_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["tag_id"], name: "index_todo_tags_on_tag_id"
    t.index ["todo_id", "tag_id"], name: "index_todo_tags_on_todo_id_and_tag_id", unique: true
    t.index ["todo_id"], name: "index_todo_tags_on_todo_id"
  end

  create_table "todos", force: :cascade do |t|
    t.string "title", null: false
    t.integer "position"
    t.boolean "completed", default: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.date "due_date"
    t.bigint "user_id", null: false
    t.integer "priority", default: 1, null: false
    t.integer "status", default: 0, null: false
    t.text "description"
    t.bigint "category_id"
    t.index ["category_id"], name: "index_todos_on_category_id"
    t.index ["created_at"], name: "index_todos_on_created_at"
    t.index ["description"], name: "index_todos_on_description"
    t.index ["due_date"], name: "index_todos_on_due_date"
    t.index ["position"], name: "index_todos_on_position"
    t.index ["priority"], name: "index_todos_on_priority"
    t.index ["status"], name: "index_todos_on_status"
    t.index ["title"], name: "index_todos_on_title"
    t.index ["updated_at"], name: "index_todos_on_updated_at"
    t.index ["user_id", "category_id"], name: "index_todos_on_user_id_and_category_id"
    t.index ["user_id", "due_date"], name: "index_todos_on_user_id_and_due_date"
    t.index ["user_id", "position"], name: "index_todos_on_user_id_and_position"
    t.index ["user_id", "priority"], name: "index_todos_on_user_id_and_priority"
    t.index ["user_id", "status"], name: "index_todos_on_user_id_and_status"
    t.index ["user_id"], name: "index_todos_on_user_id"
  end

  create_table "users", force: :cascade do |t|
    t.string "email", default: "", null: false
    t.string "encrypted_password", default: "", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "remember_created_at"
    t.string "name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["email"], name: "index_users_on_email", unique: true
    t.index ["reset_password_token"], name: "index_users_on_reset_password_token", unique: true
  end

  add_foreign_key "active_storage_attachments", "active_storage_blobs", column: "blob_id"
  add_foreign_key "active_storage_variant_records", "active_storage_blobs", column: "blob_id"
  add_foreign_key "categories", "users"
  add_foreign_key "comments", "users"
  add_foreign_key "note_revisions", "notes"
  add_foreign_key "note_revisions", "users"
  add_foreign_key "notes", "users"
  add_foreign_key "tags", "users"
  add_foreign_key "todo_histories", "todos"
  add_foreign_key "todo_histories", "users"
  add_foreign_key "todo_tags", "tags"
  add_foreign_key "todo_tags", "todos"
  add_foreign_key "todos", "categories"
  add_foreign_key "todos", "users"
end
