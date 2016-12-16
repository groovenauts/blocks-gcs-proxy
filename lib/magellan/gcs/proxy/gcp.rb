# coding: utf-8
require 'magellan/gcs/proxy'

require 'google/cloud/logging'
require 'google/cloud/logging/version'
require 'google/cloud/storage'
require 'google/apis/pubsub_v1'
require 'net/http'

module Magellan
  module Gcs
    module Proxy
      module GCP
        extend Log
        include Log

        module_function

        def project_id
          @project_id ||= retrieve_project_id
        end

        def retrieve_project_id
          ENV['BLOCKS_BATCH_PROJECT_ID'] || retrieve_metadata('project/project-id')
        end

        SCOPES = [
          'https://www.googleapis.com/auth/devstorage.full_control',
          'https://www.googleapis.com/auth/pubsub'
        ].freeze

        def auth
          unless @auth
            @auth = ::Google::Auth.get_application_default(SCOPES)
            @auth.fetch_access_token!
          end
          @auth
        end

        METADATA_HOST = 'metadata.google.internal'.freeze
        METADATA_PATH_BASE = '/computeMetadata/v1/'.freeze
        METADATA_HEADER = { 'Metadata-Flavor' => 'Google' }.freeze

        def retrieve_metadata(key)
          http = Net::HTTP.new(METADATA_HOST)
          res = http.get(METADATA_PATH_BASE + key, METADATA_HEADER)
          case res.code
          when /\A2\d{2}\z/ then res.body
          else raise "[#{res.code}] #{res.body}"
          end
        end

        def storage
          @storage ||= Google::Cloud::Storage.new(project: project_id)
        end

        def pubsub
          @pubsub ||= Google::Apis::PubsubV1::PubsubService.new.tap {|api| api.authorization = auth }
        end

        def subscription
          unless @subscription
            @subscription = PubsubSubscription.new(ENV['BLOCKS_BATCH_PUBSUB_SUBSCRIPTION'] || 'test-subscription')
            logger.info("subscription: #{@subscription.inspect}")
          end
          @subscription
        end

        def logging
          @logging ||= Google::Cloud::Logging.new(project: project_id)
        end

        def reset
          instance_variables.each { |ivar| instance_variable_set(ivar, nil) }
        end
      end
    end
  end
end
