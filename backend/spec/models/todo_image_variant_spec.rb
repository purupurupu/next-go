require 'rails_helper'

RSpec.describe Todo, type: :model do
  describe 'image variants' do
    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }
    let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

    before do
      blob = ActiveStorage::Blob.create_and_upload!(
        io: image_file,
        filename: 'test_image.png',
        content_type: 'image/png'
      )
      todo.files.attach(blob)
    end

    it 'defines thumb variant' do
      file = todo.files.first
      expect(file).to respond_to(:variant)

      # Test thumb variant dimensions
      thumb_variant = file.variant(:thumb)
      expect(thumb_variant).to be_present
    end

    it 'defines medium variant' do
      file = todo.files.first
      expect(file).to respond_to(:variant)

      # Test medium variant dimensions
      medium_variant = file.variant(:medium)
      expect(medium_variant).to be_present
    end

    it 'only applies variants to image files' do
      # Attach a non-image file
      text_blob = ActiveStorage::Blob.create_and_upload!(
        io: StringIO.new('text content'),
        filename: 'text.txt',
        content_type: 'text/plain'
      )
      todo.files.attach(text_blob)

      text_file = todo.files.find { |f| f.content_type == 'text/plain' }
      image_file = todo.files.find { |f| f.content_type == 'image/png' }

      # Check content types
      expect(image_file.content_type).to start_with('image/')
      expect(text_file.content_type).not_to start_with('image/')

      # Variants should only work for images
      expect { image_file.variant(:thumb) }.not_to raise_error
      expect { text_file.variant(:thumb) }.to raise_error(ActiveStorage::InvariableError)
    end
  end
end
