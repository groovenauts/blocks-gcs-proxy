require 'magellan/gcs/proxy'
require 'magellan/gcs/proxy/log'

require 'json'
require 'logger'
require 'tmpdir'
require 'logger_pipe'

module Magellan
  module Gcs
    module Proxy
      class Cli
        include Log

        attr_reader :cmd_template
        def initialize(*args)
          @cmd_template = args.join(' ')
        end

        def run
          logger.info("#{$PROGRAM_NAME}-#{VERSION} is running")
          logger.info('Start listening')
          GCP.subscription.listen do |msg|
            process_with_error_handling(msg)
          end
        rescue => e
          logger.error("[#{e.class.name}] #{e.message}")
          raise e
        end

        def process_with_error_handling(msg)
          process(msg)
        rescue => e
          logger.error("[#{e.class.name}] #{e.message}")
          verbose("Backtrace\n  " << e.backtrace.join("\n  "))
        end

        TOTAL = 14
        def process(msg)
          context = Context.new(msg)
          context.notify(1, TOTAL, "Processing message: #{msg.inspect}")
          context.setup do
            context.process_with_notification([2, 3, 4], TOTAL, 'Download', &:download)

            cmd = build_command(context)

            exec = ->(*) { LoggerPipe.run(logger, cmd, returns: :none, logging: :both) }
            context.process_with_notification([5, 6, 7], TOTAL, 'Command', exec) do
              context.process_with_notification([8, 9, 10], TOTAL, 'Upload', &:upload)

              context.process_with_notification([11, 12, 13], TOTAL, 'Acknowledge') do
                msg.acknowledge!
              end
            end
          end
          context.notify(14, TOTAL, 'Cleanup')
        end

        def build_command(context)
          msg_wrapper = MessageWrapper.new(context)
          ExpandVariable.expand_variables(cmd_template, msg_wrapper)
        end
      end
    end
  end
end
