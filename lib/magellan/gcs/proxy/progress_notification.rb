require "magellan/gcs/proxy"

module Magellan
  module Gcs
    module Proxy
      module ProgressNotification
        include Log

        def notifier
          @notifier ||= build_notifier
        end

        def build_notifier
          notifiers = []
          if c = Proxy.config[:progress_notification]
            notifiers << ProgressNotifierAdapter.new(c['topic'])
          end
          notifiers << ProgressNotifierAdapter.new(logger)
          CompositeNotifier.new(notifiers)
        end

        class CompositeNotifier
          attr_reader :notifiers
          def initialize(notifiers)
            @notifiers = notifiers
          end

          def notify(*args, &block)
            notifiers.each do |notifier|
              begin
                notifier.notify(*args, &block)
              rescue => e
                $stderr.puts("[#{e.class}] #{e.message}")
              end
            end
          end
        end

      end
    end
  end
end
