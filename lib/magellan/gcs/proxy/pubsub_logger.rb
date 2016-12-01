require "magellan/gcs/proxy"

require 'logger'
require 'json'

module Magellan
  module Gcs
    module Proxy
      class PubsubProgressNotifier
        attr_reader :topic
        def initialize(topic)
          @topic = topic
        end

        def notify(severity, job_message, data, attrs)
          topic.publish data, {level: severity}.merge(attrs)
        end

      end
    end
  end
end
