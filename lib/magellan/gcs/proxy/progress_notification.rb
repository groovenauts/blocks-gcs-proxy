require 'magellan/gcs/proxy'

module Magellan
  module Gcs
    module Proxy
      module ProgressNotification
        include Log

        def process_with_notification(numbers, total, base_message, main = nil)
          start_no, complete_no, error_no = *numbers
          notify(start_no, total, "#{base_message} starting")
          begin
            main ? main.call(self) : yield(self)
          rescue => e
            notify(error_no, total, "#{base_message} error: [#{e.class}] #{e.message}", severity: :error)
            raise e unless main
          else
            notify(complete_no, total, "#{base_message} completed")
            yield(self) if main
          end
        end

        def notify(progress, total, data, severity: :info)
          notifier.notify(severity, message, data, progress: progress, total: total)
        end

        def notifier
          @notifier ||= build_notifier
        end

        # Build the Notifier object like these...
        #
        # CompositeNotifier
        #   @notifiers:
        #     PubsubProgressNotifier
        #     ProgressNotifierAdapter
        #       @logger:
        #         CompositeLogger
        #           @loggers:
        #             Logger
        #             Google::Cloud::Logging::Logger
        def build_notifier
          notifiers = []
          if c = Proxy.config[:progress_notification]
            notifiers << PubsubProgressNotifier.new(c['topic'])
          end
          notifiers << ProgressNotifierAdapter.new(logger)
          case notifiers.length
          when 1 then notifiers.first
          else CompositeNotifier.new(notifiers)
          end
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
