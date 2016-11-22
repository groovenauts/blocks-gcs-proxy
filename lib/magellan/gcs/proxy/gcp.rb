# coding: utf-8
require 'magellan/gcs/proxy'

require "google/cloud/pubsub"
require "google/cloud/storage"
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

        METADATA_HOST = 'metadata.google.internal'.freeze
        METADATA_PATH_BASE = '/computeMetadata/v1/'.freeze
        METADATA_HEADER = {"Metadata-Flavor" => "Google"}.freeze

        def retrieve_metadata(key)
          http = Net::HTTP.new(METADATA_HOST)
          res = http.get(METADATA_PATH_BASE + key, METADATA_HEADER)
          case res.code
          when /\A2\d{2}\z/ then res.body
          else raise "#{res.code} #{res.body}"
          end
        end

        def storage
          @storage ||= Google::Cloud::Storage.new(project: project_id)
        end

        def pubsub
          @pubsub ||= Google::Cloud::Pubsub.new(project: project_id)
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
