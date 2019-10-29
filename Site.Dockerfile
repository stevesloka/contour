FROM jekyll/jekyll:3.8.5
WORKDIR /site
COPY site/Gemfile site
COPY site/Gemfile.lock site
RUN cd site && bundle install