require 'rails_helper'

RSpec.describe Todo, type: :model do
  describe 'file validations' do
    let(:user) { create(:user) }
    let(:todo) { build(:todo, user: user) }

    describe 'file size validation' do
      it 'rejects files larger than 10MB' do
        # Create a fake large file
        large_file = {
          io: StringIO.new('x' * 11.megabytes),
          filename: 'large_file.txt',
          content_type: 'text/plain'
        }

        blob = ActiveStorage::Blob.create_and_upload!(**large_file)
        todo.files.attach(blob)

        expect(todo).not_to be_valid
        expect(todo.errors[:files]).to include(/ファイルサイズは10MB以下にしてください/)
      end

      it 'accepts files smaller than 10MB' do
        small_file = {
          io: StringIO.new('small content'),
          filename: 'small_file.txt',
          content_type: 'text/plain'
        }

        blob = ActiveStorage::Blob.create_and_upload!(**small_file)
        todo.files.attach(blob)

        expect(todo).to be_valid
      end
    end

    describe 'file type validation' do
      context 'allowed file types' do
        [
          ['image/jpeg', 'photo.jpg'],
          ['image/png', 'image.png'],
          ['image/gif', 'animation.gif'],
          ['image/webp', 'modern.webp'],
          ['application/pdf', 'document.pdf'],
          ['text/plain', 'readme.txt'],
          ['text/csv', 'data.csv'],
          ['application/msword', 'old_doc.doc'],
          ['application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'new_doc.docx'],
          ['application/vnd.ms-excel', 'old_sheet.xls'],
          ['application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 'new_sheet.xlsx']
        ].each do |content_type, filename|
          it "accepts #{content_type} files" do
            file = {
              io: StringIO.new('test content'),
              filename: filename,
              content_type: content_type
            }

            blob = ActiveStorage::Blob.create_and_upload!(**file)
            todo.files.attach(blob)

            expect(todo).to be_valid
          end
        end
      end

      context 'disallowed file types' do
        [
          ['application/x-executable', 'program.exe'],
          ['application/x-sh', 'script.sh'],
          ['application/javascript', 'code.js'],
          ['video/mp4', 'video.mp4'],
          ['audio/mpeg', 'song.mp3']
        ].each do |content_type, filename|
          it "rejects #{content_type} files" do
            file = {
              io: StringIO.new('test content'),
              filename: filename,
              content_type: content_type
            }

            blob = ActiveStorage::Blob.create_and_upload!(**file)
            todo.files.attach(blob)

            expect(todo).not_to be_valid
            expect(todo.errors[:files]).to include(/許可されていないファイルタイプです/)
          end
        end
      end
    end

    describe 'multiple file validation' do
      it 'validates each file independently' do
        # One valid file
        valid_file = {
          io: StringIO.new('valid content'),
          filename: 'valid.txt',
          content_type: 'text/plain'
        }
        valid_blob = ActiveStorage::Blob.create_and_upload!(**valid_file)

        # One invalid file (wrong type)
        invalid_file = {
          io: StringIO.new('invalid content'),
          filename: 'invalid.exe',
          content_type: 'application/x-executable'
        }
        invalid_blob = ActiveStorage::Blob.create_and_upload!(**invalid_file)

        todo.files.attach([valid_blob, invalid_blob])

        expect(todo).not_to be_valid
        expect(todo.errors[:files]).to include(/許可されていないファイルタイプです/)
      end
    end
  end

  describe 'file cleanup on destroy' do
    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }

    it 'deletes all attached files when todo is destroyed' do
      # Attach files
      file1 = {
        io: StringIO.new('content 1'),
        filename: 'file1.txt',
        content_type: 'text/plain'
      }
      file2 = {
        io: StringIO.new('content 2'),
        filename: 'file2.txt',
        content_type: 'text/plain'
      }

      blob1 = ActiveStorage::Blob.create_and_upload!(**file1)
      blob2 = ActiveStorage::Blob.create_and_upload!(**file2)

      todo.files.attach([blob1, blob2])

      # Get attachment count
      initial_attachment_count = ActiveStorage::Attachment.count
      file_count = todo.files.count

      # Destroy todo
      todo.destroy

      # Check that attachments are deleted
      expect(ActiveStorage::Attachment.count).to eq(initial_attachment_count - file_count)
    end
  end
end
