FROM ruby:3.4.5

# 必要なパッケージのインストール
RUN apt-get update -qq && apt-get install -y nodejs postgresql-client

# 作業ディレクトリの設定
WORKDIR /app

# GemfileとGemfile.lockをコピー
COPY Gemfile Gemfile.lock ./

# Bundlerのインストールと実行
RUN gem install bundler
RUN bundle install

# アプリケーションのコードをコピー
COPY . .

# 環境変数の設定
ENV RAILS_ENV=development
ENV PATH="/app/bin:${PATH}"

# RAILS_MASTER_KEYを環境変数から設定
ARG RAILS_MASTER_KEY
ENV RAILS_MASTER_KEY=${RAILS_MASTER_KEY}

# エントリーポイントスクリプトの作成
RUN echo '#!/bin/bash\nset -e\n\nif [ -f tmp/pids/server.pid ]; then\n  rm tmp/pids/server.pid\nfi\n\nexec "$@"' > /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]

EXPOSE 3000

# Railsサーバーの起動
CMD ["rails", "server", "-b", "0.0.0.0"]
