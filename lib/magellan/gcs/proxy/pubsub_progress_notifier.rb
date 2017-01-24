require 'magellan/gcs/proxy'
require 'magellan/gcs/proxy/log'

require 'logger'
require 'json'

module Magellan
  module Gcs
    module Proxy
      class PubsubProgressNotifier
        include Log

        attr_reader :topic_name
        def initialize(topic_name)
          @topic_name = topic_name
        end

        def topic
          @topic ||= GCP.pubsub.topic(topic_name)
        end

        def notify(severity, job_message, data, attrs)
          attrs = { level: severity, job_message_id: job_message.message_id }.merge(attrs)
          # attrs must be an [Hash<String,String>]
          attrs = attrs.each_with_object({}) { |(k, v), d| d[k.to_s] = v.to_s }
          logger.debug("Publishing progress: #{attrs.inspect}")
          msg = Google::Apis::PubsubV1::Message.new(data: data, attributes: attrs)
          req = Google::Apis::PubsubV1::PublishRequest.new(messages: [msg])
          GCP.pubsub.publish_topic(topic_name, req)
        end
      end
    end
  end
end
