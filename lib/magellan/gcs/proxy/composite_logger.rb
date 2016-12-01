require "magellan/gcs/proxy"

module Magellan
  module Gcs
    module Proxy
      class CompositeLogger
        attr_reader :loggers
        def initialize(loggers)
          @loggers = loggers
        end

        def add(severity, message = nil, &block)
          loggers.each do |logger|
            begin
              logger.add(severity, message, &block)
            rescue => e
              $stderr.puts("[#{e.class}] #{e.message}")
            end
          end
        end

        Logger::Severity.constants.each do |level|
          level_dc = level.downcase
          module_eval <<-INSTANCE_METHODS
            def #{level_dc}(message = nil, &block)
              add(:#{level_dc}, message, &block)
            end

            def #{level_dc}?
              loggers.any?(&:#{level_dc}?)
            end
          INSTANCE_METHODS
        end

      end
    end
  end
end
