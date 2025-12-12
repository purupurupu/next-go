# frozen_string_literal: true

require 'rails_helper'

RSpec.describe Comment, type: :model do
  # 学習ポイント：モデルテストのベストプラクティス

  describe 'associations' do
    it { is_expected.to belong_to(:user) }
    it { is_expected.to belong_to(:commentable) }
  end

  describe 'validations' do
    it { is_expected.to validate_presence_of(:content) }
    it { is_expected.to validate_length_of(:content).is_at_most(1000) }
  end

  describe 'scopes' do
    let!(:active_comment) { create(:comment) }
    let!(:deleted_comment) { create(:comment, :deleted) }

    describe '.active' do
      it 'returns only non-deleted comments' do
        # デフォルトスコープを無視するためにunscoped
        expect(described_class.unscoped.active).to include(active_comment)
        expect(described_class.unscoped.active).not_to include(deleted_comment)
      end
    end

    describe '.deleted' do
      it 'returns only deleted comments' do
        expect(described_class.unscoped.deleted).to include(deleted_comment)
        expect(described_class.unscoped.deleted).not_to include(active_comment)
      end
    end

    describe '.chronological' do
      let!(:old_comment) { create(:comment, created_at: 2.days.ago) }
      let!(:new_comment) { create(:comment, created_at: 1.day.ago) }

      it 'orders comments by created_at ascending' do
        result = described_class.chronological
        expect(result.first.created_at).to be < result.last.created_at
        expect(result).to include(old_comment, new_comment)
      end
    end

    describe '.recent' do
      let!(:old_comment) { create(:comment, created_at: 2.days.ago) }
      let!(:new_comment) { create(:comment, created_at: 1.day.ago) }

      it 'orders comments by created_at descending' do
        result = described_class.recent
        expect(result.first.created_at).to be > result.last.created_at
        expect(result).to include(old_comment, new_comment)
      end
    end
  end

  describe 'default scope' do
    let!(:active_comment) { create(:comment) }
    let!(:deleted_comment) { create(:comment, :deleted) }

    it 'excludes deleted comments by default' do
      expect(described_class.all).to include(active_comment)
      expect(described_class.all).not_to include(deleted_comment)
    end
  end

  describe '#soft_delete!' do
    let(:comment) { create(:comment) }

    it 'sets deleted_at timestamp' do
      expect { comment.soft_delete! }.to change(comment, :deleted_at).from(nil)
    end

    it 'does not destroy the record' do
      comment.soft_delete!
      expect(described_class.unscoped.find(comment.id)).to be_present
    end
  end

  describe '#restore!' do
    let(:comment) { create(:comment, :deleted) }

    it 'clears deleted_at timestamp' do
      expect { comment.restore! }.to change(comment, :deleted_at).to(nil)
    end
  end

  describe '#deleted?' do
    let(:active_comment) { create(:comment) }
    let(:deleted_comment) { create(:comment, :deleted) }

    it 'returns true for deleted comments' do
      expect(deleted_comment.deleted?).to be true
    end

    it 'returns false for active comments' do
      expect(active_comment.deleted?).to be false
    end
  end

  describe '#editable?' do
    context 'when comment is not deleted' do
      let(:comment) { create(:comment) }

      it 'returns true within 15 minutes of creation' do
        expect(comment.editable?).to be true
      end

      it 'returns false after 15 minutes of creation' do
        comment.update!(created_at: 20.minutes.ago)
        expect(comment.editable?).to be false
      end
    end

    context 'when comment is deleted' do
      let(:comment) { create(:comment, :deleted) }

      it 'returns false' do
        expect(comment.editable?).to be false
      end
    end
  end

  describe '#owned_by?' do
    let(:user) { create(:user) }
    let(:other_user) { create(:user) }
    let(:comment) { create(:comment, user: user) }

    it 'returns true when user owns the comment' do
      expect(comment.owned_by?(user)).to be true
    end

    it 'returns false when user does not own the comment' do
      expect(comment.owned_by?(other_user)).to be false
    end

    it 'returns false when user is nil' do
      expect(comment.owned_by?(nil)).to be false
    end
  end

  describe 'polymorphic association' do
    let(:todo) { create(:todo) }
    let(:comment) { create(:comment, commentable: todo) }

    it 'can be associated with a todo' do
      expect(comment.commentable).to eq(todo)
      expect(comment.commentable_type).to eq('Todo')
      expect(comment.commentable_id).to eq(todo.id)
    end

    it 'can be accessed through the todo' do
      expect(todo.comments).to include(comment)
    end
  end
end
