require "magellan/gcs/proxy"

module Magellan
  module Gcs
    module Proxy
      class ProgressNotifierAdapter
        attr_reader :logger
        def initialize(logger)
          @logger = logger
        end

        def ltsv(hash)
          hash.map{|k,v| "#{k}:#{v}"}.join("\t")
        end

        def notify(severity, job_message, data, attrs)
          d = {job_message_id: job_message.message_id}.merge(attrs)
          d[:data] = data # Show data at the end of string
          logger.send(severity, ltsv(d))
        end

      end
    end
  end
end
