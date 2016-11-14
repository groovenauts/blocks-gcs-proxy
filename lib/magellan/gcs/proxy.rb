require "magellan/gcs/proxy/version"

module Magellan
  module Gcs
    module Proxy
      autoload :Command, 'magellan/gcs/proxy/command'
      autoload :FileOperation, 'magellan/gcs/proxy/file_operation'
      autoload :PubsubOperation, 'magellan/gcs/proxy/pubsub_operation'
    end
  end
end
