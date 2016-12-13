require 'dotenv'
Dotenv.load

require 'magellan/gcs/proxy/version'
require 'magellan/gcs/proxy/expand_variable'
require 'magellan/gcs/proxy/config'
require 'magellan/gcs/proxy/log'
require 'magellan/gcs/proxy/gcp'

require 'magellan/gcs/proxy/composite_logger'
require 'magellan/gcs/proxy/progress_notifier_adapter'
require 'magellan/gcs/proxy/pubsub_progress_notifier'
require 'magellan/gcs/proxy/pubsub_sustainer'
require 'magellan/gcs/proxy/progress_notification'

require 'magellan/gcs/proxy/message_wrapper'
require 'magellan/gcs/proxy/context'
require 'magellan/gcs/proxy/cli'

module Magellan
  module Gcs
    module Proxy
      class << self
        def config
          @config ||= Config.new
        end
      end
    end
  end
end
