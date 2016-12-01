require 'logger'
module Magellan
  module Gcs
    module Proxy
      module Log
        def logger
          @logger ||= Logger.new($stdout)
        CLOUD_LOGGING_RESOURCE_KEYS = [
          :project_id,
					:cluster_name,
					:namespace_id,
					:instance_id,
					:pod_id,
					:container_name,
					:zone,
        ].freeze

        def build_cloud_logging_logger(config)
          log_name = config['log_name']
          return nil unless log_name
          # container
          # GKE Container	A Google Container Engine (GKE) container instance.
          #   project_id: The identifier of the GCP project associated with this resource (e.g., my-project).
          #   cluster_name: An immutable name for the cluster the container is running in.
          #   namespace_id: Immutable ID of the cluster namespace the container is running in.
          #   instance_id: Immutable ID of the GCE instance the container is running in.
          #   pod_id: Immutable ID of the pod the container is running in.
          #   container_name: Immutable name of the container.
          #   zone: The GCE zone in which the instance is running.
          # See https://cloud.google.com/logging/docs/api/v2/resource-list
          options = CLOUD_LOGGING_RESOURCE_KEYS.each_with_object({}) do |key, d|
            if v = ENV["BLOCKS_BATCH_CLOUD_LOGGING_#{key.to_s.upcase}"]
              d[key] = v
            end
          end
          resource = GCP.logging.resource "container", options
          Google::Cloud::Logging::Logger.new GCP.logging, log_name, resource,
                                             {magellan_gcs_proxy: Magellan::Gcs::Proxy::VERSION}
        end

        def build_pubsub_logger(config)
          topic = GCP.pubsub.topic(config['topic'])
          Proxy::PubsubLogger.new(topic)
        end
      end
    end
  end
end
