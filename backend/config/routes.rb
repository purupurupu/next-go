Rails.application.routes.draw do
  devise_for :users, path: 'auth', path_names: {
    sign_in: 'sign_in',
    sign_out: 'sign_out',
    registration: 'sign_up'
  }, controllers: {
    sessions: 'users/sessions',
    registrations: 'users/registrations'
  }
  
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check

  # Defines the root path route ("/")
  # root "posts#index"

  # MCP endpoint
  post '/mcp', to: 'mcp#handle'

  # API routes
  namespace :api do
    # Explicit v1 namespace for URL-based versioning
    namespace :v1 do
      resources :todos do
        collection do
          patch 'update_order'
          get 'search'
        end
        member do
          patch 'tags', to: 'todos#update_tags'
          delete 'files/:file_id', to: 'todos#destroy_file', as: 'destroy_file'
        end
        resources :comments, only: [:index, :create, :update, :destroy]
        resources :histories, only: [:index], controller: 'todo_histories'
      end
      resources :categories
      resources :tags
      resources :notes do
        resources :revisions, only: [:index], controller: 'note_revisions' do
          post 'restore', on: :member
        end
      end
    end
    
  end

end
