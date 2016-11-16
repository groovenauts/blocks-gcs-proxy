require "magellan/gcs/proxy/version"

module Magellan
  module Gcs
    module Proxy
      autoload :Cli, 'magellan/gcs/proxy/cli'
      autoload :Context, 'magellan/gcs/proxy/context'
      autoload :ExpandVariable, 'magellan/gcs/proxy/expand_variable'
      autoload :FileOperation, 'magellan/gcs/proxy/file_operation'
      autoload :MessageWrapper, 'magellan/gcs/proxy/message_wrapper'
      autoload :PubsubOperation, 'magellan/gcs/proxy/pubsub_operation'
    end
  end
end
