require 'logger'
module Magellan
  module Gcs
    module Proxy
      module Log
        def logger
          @logger ||= Logger.new($stdout)
        end
      end
    end
  end
end
