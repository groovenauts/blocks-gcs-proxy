require "magellan/gcs/proxy"

module Magellan
  module Gcs
    module Proxy
      module PubsubOperation
        def topic
          unless @topic
            topic_name = ENV['BATCH_TOPIC_NAME'] || 'test-topic'
            @topic = GCP.pubsub.topic(topic_name) || pubsub.create_topic(topic_name)
            logger.info("topic: #{@topic.inspect}")
          end
          @topic
        end

        def sub
          unless @sub
            sub_name = ENV['BATCH_SUBSCRIPTION_NAME'] || 'test-subscription'
            @sub = topic.subscription(sub_name) || topic.subscribe(sub_name)
            logger.info("subscription: #{@sub.inspect}")
          end
          @sub
        end

      end
    end
  end
end
