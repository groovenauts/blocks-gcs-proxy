# coding: utf-8
require 'magellan/gcs/proxy'

require "google/cloud/pubsub"
require "google/cloud/storage"

module Magellan
  module Gcs
    module Proxy
      module GCP

        module_function

        SCOPES = [
          "https://www.googleapis.com/auth/devstorage.full_control",
          "https://www.googleapis.com/auth/pubsub",
        ].freeze

        # See
        #   https://developers.google.com/identity/protocols/application-default-credentials
        #   https://github.com/google/google-auth-library-ruby
        def authorization
          @authorization ||= Google::Auth.get_application_default(SCOPES)
        end

        def storage
          @storage ||= Google::Cloud::Storage.new(
            # default credential を利用するため、プロジェクトの指定はしない
            # project: ENV['GOOGLE_PROJECT'] || 'dummy-project-id',
            # keyfile: ENV['GOOGLE_KEY_JSON_FILE'],
          )
        end

        def pubsub
          @pubsub ||= Google::Cloud::Pubsub.new(
            # default credential を利用するため、プロジェクトの指定はしない
            # project: ENV['GOOGLE_PROJECT'] || 'dummy-project-id',
            # keyfile: ENV['GOOGLE_KEY_JSON_FILE'],
          )
        end

        def subscription
          unless @subscription
            @subscription = pubsub.subscription(ENV['BLOCKS_BATCH_PUBSUB_SUBSCRIPTION'] || 'test-subscription')
            logger.info("subscription: #{@subscription.inspect}")
          end
          @subscription
        end
      end
    end
  end
end
