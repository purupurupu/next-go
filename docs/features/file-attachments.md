# File Attachments Feature

## Overview

The file attachments feature allows users to attach files to their todos using Rails Active Storage. This provides a flexible way to associate documents, images, and other files with tasks.

## Features

### 1. File Upload
- Multiple file attachments per todo
- Drag-and-drop support
- File type validation
- Size limit enforcement
- Progress indication during upload

### 2. File Management
- View attached files with metadata
- Download files
- Delete individual attachments
- Thumbnail generation for images

### 3. Supported File Types
- **Documents**: PDF, DOC, DOCX, TXT, MD
- **Images**: JPG, JPEG, PNG, GIF, WebP
- **Spreadsheets**: XLS, XLSX, CSV
- **Archives**: ZIP, RAR, 7Z
- **Code**: JS, TS, PY, RB, JSON, XML

## Technical Implementation

### Backend

#### Model Configuration
```ruby
class Todo < ApplicationRecord
  has_many_attached :files
  
  validate :acceptable_files
  
  private
  
  def acceptable_files
    return unless files.attached?
    
    files.each do |file|
      unless file.blob.byte_size <= 10.megabytes
        errors.add(:files, "#{file.filename} is too large (max 10MB)")
      end
      
      unless acceptable_file_type?(file)
        errors.add(:files, "#{file.filename} has an invalid file type")
      end
    end
  end
  
  def acceptable_file_type?(file)
    acceptable_types = %w[
      image/jpeg image/jpg image/png image/gif image/webp
      application/pdf
      application/msword application/vnd.openxmlformats-officedocument.wordprocessingml.document
      application/vnd.ms-excel application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
      text/plain text/csv text/markdown
      application/zip application/x-rar-compressed application/x-7z-compressed
      application/json application/xml
      text/javascript application/javascript
    ]
    
    acceptable_types.include?(file.blob.content_type)
  end
end
```

#### API Endpoints

1. **Upload Files** (with todo creation/update)
   ```
   POST /api/v1/todos
   Content-Type: multipart/form-data
   
   todo[title]: "Task with attachments"
   todo[files][]: [file1]
   todo[files][]: [file2]
   ```

2. **Delete File**
   ```
   DELETE /api/v1/todos/:todo_id/files/:file_id
   ```
   
   Response:
   ```json
   {
     "message": "File deleted successfully",
     "todo": { /* updated todo without the file */ }
   }
   ```

#### Serialization
```ruby
class TodoSerializer < ActiveModel::Serializer
  attributes :id, :title, :completed, :files
  
  def files
    return [] unless object.files.attached?
    
    object.files.map do |file|
      {
        id: file.id,
        filename: file.filename.to_s,
        content_type: file.blob.content_type,
        byte_size: file.blob.byte_size,
        url: Rails.application.routes.url_helpers.rails_blob_url(file, only_path: true),
        created_at: file.blob.created_at
      }
    end
  end
end
```

### Frontend

#### Components

1. **FileUpload**
   - Drag-and-drop zone
   - File selection dialog
   - Upload progress
   - File validation

2. **AttachmentList**
   - Display attached files
   - File icons by type
   - Download links
   - Delete buttons

3. **FilePreview**
   - Image thumbnails
   - File metadata display
   - Quick actions

#### Usage Example

```typescript
// File upload component
<FileUpload
  onFilesSelected={handleFilesSelected}
  maxSize={10 * 1024 * 1024} // 10MB
  accept={ACCEPTED_FILE_TYPES}
  multiple
/>

// Handling file upload
const handleSubmit = async (data: TodoFormData) => {
  const formData = new FormData();
  formData.append('todo[title]', data.title);
  
  if (data.files) {
    data.files.forEach(file => {
      formData.append('todo[files][]', file);
    });
  }
  
  await todoApi.create(formData);
};

// Display attachments
<AttachmentList
  attachments={todo.files}
  onDelete={(fileId) => handleFileDelete(todo.id, fileId)}
/>
```

## File Storage

### Active Storage Configuration
```ruby
# config/storage.yml
local:
  service: Disk
  root: <%= Rails.root.join("storage") %>

# Production should use cloud storage
amazon:
  service: S3
  access_key_id: <%= ENV['AWS_ACCESS_KEY_ID'] %>
  secret_access_key: <%= ENV['AWS_SECRET_ACCESS_KEY'] %>
  region: <%= ENV['AWS_REGION'] %>
  bucket: <%= ENV['AWS_BUCKET'] %>
```

### Database Tables
Active Storage creates these tables:
- `active_storage_blobs`: File metadata
- `active_storage_attachments`: Polymorphic join table
- `active_storage_variant_records`: Image variants

## Security Considerations

1. **File Type Validation**: Whitelist allowed file types
2. **Size Limits**: Enforce maximum file size (10MB)
3. **Virus Scanning**: Implement for production
4. **Access Control**: Files inherit todo permissions
5. **Secure URLs**: Use signed URLs for downloads

## Performance Optimizations

1. **Lazy Loading**: Files loaded only when needed
2. **Direct Uploads**: Upload directly to storage service
3. **Background Processing**: Generate thumbnails asynchronously
4. **CDN Integration**: Serve files through CDN
5. **Compression**: Compress files before storage

## User Interface

### Upload States
1. **Idle**: Drag files here or click to browse
2. **Dragging**: Drop files to upload
3. **Uploading**: Show progress bar
4. **Success**: File uploaded successfully
5. **Error**: Display error message

### File Display
```
📎 project-spec.pdf (245 KB)
🖼️ mockup.png (1.2 MB) [Preview]
📊 data.xlsx (890 KB)
📁 source-code.zip (3.4 MB)
```

## Limitations

1. **File Size**: Maximum 10MB per file
2. **Total Storage**: Consider per-user limits
3. **File Types**: Limited to safe file types
4. **Concurrent Uploads**: Browser limitations apply

## Error Handling

### Common Errors
1. **File Too Large**: "File exceeds 10MB limit"
2. **Invalid Type**: "File type not supported"
3. **Upload Failed**: "Failed to upload file. Please try again"
4. **Network Error**: "Check your connection and try again"

## Future Enhancements

1. **Image Editor**: Basic image editing capabilities
2. **File Preview**: In-app preview for more file types
3. **Version Control**: Track file versions
4. **Collaboration**: Comments on files
5. **OCR**: Text extraction from images
6. **Compression**: Automatic file optimization
7. **Sharing**: Generate shareable links
8. **Integration**: Google Drive, Dropbox integration