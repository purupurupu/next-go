# frozen_string_literal: true

# Shared context for tests that require a user
# This reduces the number of user records created during tests
RSpec.shared_context 'with user' do
  let(:user) { create(:user) }
end

# Shared context for tests that require multiple users
RSpec.shared_context 'with users' do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
end

# Shared context for tests with test_it_be (for suite-wide shared data)
RSpec.shared_context 'with shared user' do
  let_it_be(:shared_user) { create(:user) }
end
