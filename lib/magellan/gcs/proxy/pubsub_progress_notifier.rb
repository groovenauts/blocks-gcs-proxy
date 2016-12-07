require 'magellan/gcs/proxy'

require 'logger'
require 'json'

module Magellan
  module Gcs
    module Proxy
      class PubsubProgressNotifier
        attr_reader :topic_name
        def initialize(topic_name)
          @topic_name = topic_name
        end

        def topic
          @topic ||= GCP.pubsub.topic(topic_name)
        end

        def notify(severity, job_message, data, attrs)
          topic.publish data, { level: severity, job_message_id: job_message.message_id }.merge(attrs)
        end
      end
    end
  end
end
