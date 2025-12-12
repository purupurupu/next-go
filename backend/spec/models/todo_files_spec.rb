require 'rails_helper'

RSpec.describe Todo, type: :model do
  describe 'file uploads' do
    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }

    describe 'associations' do
      it 'can have many attached files' do
        expect(todo).to respond_to(:files)
        expect(todo.files).to be_an_instance_of(ActiveStorage::Attached::Many)
      end
    end

    describe 'uploading files' do
      let(:text_file) { fixture_file_upload('test_file.txt', 'text/plain') }
      let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

      it 'can upload a single file' do
        # Upload without analysis to avoid background job
        blob = ActiveStorage::Blob.create_and_upload!(
          io: text_file,
          filename: 'test_file.txt',
          content_type: 'text/plain'
        )
        todo.files.attach(blob)

        expect(todo.files).to be_attached
        expect(todo.files.count).to eq(1)
        expect(todo.files.first.filename.to_s).to eq('test_file.txt')
      end

      it 'can upload multiple files' do
        # Create blobs without analysis
        text_blob = ActiveStorage::Blob.create_and_upload!(
          io: text_file,
          filename: 'test_file.txt',
          content_type: 'text/plain'
        )
        image_blob = ActiveStorage::Blob.create_and_upload!(
          io: image_file,
          filename: 'test_image.png',
          content_type: 'image/png'
        )

        todo.files.attach([text_blob, image_blob])
        expect(todo.files.count).to eq(2)
        expect(todo.files.map(&:filename).map(&:to_s)).to contain_exactly('test_file.txt', 'test_image.png')
      end

      it 'preserves existing files when adding new ones' do
        # Upload first file
        text_blob = ActiveStorage::Blob.create_and_upload!(
          io: text_file,
          filename: 'test_file.txt',
          content_type: 'text/plain'
        )
        todo.files.attach(text_blob)
        expect(todo.files.count).to eq(1)

        # Upload second file
        image_blob = ActiveStorage::Blob.create_and_upload!(
          io: image_file,
          filename: 'test_image.png',
          content_type: 'image/png'
        )
        todo.files.attach(image_blob)
        expect(todo.files.count).to eq(2)
      end
    end

    describe 'file metadata' do
      let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

      before do
        blob = ActiveStorage::Blob.create_and_upload!(
          io: image_file,
          filename: 'test_image.png',
          content_type: 'image/png'
        )
        todo.files.attach(blob)
      end

      it 'provides access to file metadata' do
        file = todo.files.first
        expect(file.filename.to_s).to eq('test_image.png')
        expect(file.content_type).to eq('image/png')
        expect(file.byte_size).to be > 0
      end
    end

    describe 'file deletion' do
      let(:text_file) { fixture_file_upload('test_file.txt', 'text/plain') }
      let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

      before do
        text_blob = ActiveStorage::Blob.create_and_upload!(
          io: text_file,
          filename: 'test_file.txt',
          content_type: 'text/plain'
        )
        image_blob = ActiveStorage::Blob.create_and_upload!(
          io: image_file,
          filename: 'test_image.png',
          content_type: 'image/png'
        )
        todo.files.attach([text_blob, image_blob])
      end

      it 'can delete a specific file' do
        expect(todo.files.count).to eq(2)

        file_to_delete = todo.files.first
        file_to_delete.purge

        todo.reload
        expect(todo.files.count).to eq(1)
      end

      it 'deletes all files when todo is destroyed' do
        expect(todo.files.count).to eq(2)

        # Get attachment count before destruction
        initial_attachment_count = ActiveStorage::Attachment.count

        todo.destroy

        # Check that attachments are deleted
        expect(ActiveStorage::Attachment.count).to eq(initial_attachment_count - 2)
        # NOTE: Blobs may not be immediately deleted in test environment
      end
    end
  end
end
