require "magellan/gcs/proxy"

require 'logger'
require 'json'

module Magellan
  module Gcs
    module Proxy
      class PubsubLogger
        attr_reader :topic
        def initialize(topic, level = Logger::Severity::INFO)
          @topic = topic
          @level = level
        end

        SEVERITY_TO_NAME = Logger::Severity.constants.each_with_object({}) do |key, d|
          d[Logger::Severity.const_get(key)] = key.downcase
        end

        def add(severity, message = nil)
          return true if severity < @level
          topic.publish message, level: SEVERITY_TO_NAME[severity]
        end

        Logger::Severity.constants.each do |level|
          level_dc = level.downcase
          module_eval <<-INSTANCE_METHODS
            def #{level_dc}(message = nil, &block)
              add(Logger::Severity::#{level}, message, &block)
            end

            def #{level_dc}?
              @level <= Logger::Severity::#{level}
            end
          INSTANCE_METHODS
        end

      end
    end
  end
end
