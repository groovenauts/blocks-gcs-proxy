require 'magellan/gcs/proxy'
require 'magellan/gcs/proxy/log'

require 'json'
require 'logger'
require 'tmpdir'
require 'logger_pipe'

module Magellan
  module Gcs
    module Proxy
      class BuildError < StandardError
      end

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

        PROCESSING     =  1
        DOWNLOADING    =  2
        DOWNLOAD_OK    =  3
        DOWNLOAD_ERROR =  4
        EXECUTING      =  5
        EXECUTE_OK     =  6
        EXECUTE_ERROR  =  7
        UPLOADING      =  8
        UPLOAD_OK      =  9
        UPLOAD_ERROR   = 10
        ACKSENDING     = 11
        ACKSEND_OK     = 12
        ACKSEND_ERROR  = 13
        CLEANUP        = 14

        TOTAL = CLEANUP

        def process(msg)
          context = Context.new(msg)
          context.notify(PROCESSING, TOTAL, "Processing message: #{msg.inspect}")
          context.setup do
            context.process_with_notification([DOWNLOADING, DOWNLOAD_OK, DOWNLOAD_ERROR], TOTAL, 'Download', &:download)

            cmd = build_command(context)

            exec = ->(*) { LoggerPipe.run(logger, cmd, returns: :none, logging: :both, dry_run: Proxy.config[:dryrun]) }
            context.process_with_notification([EXECUTING, EXECUTE_OK, EXECUTE_ERROR], TOTAL, 'Command', exec) do
              context.process_with_notification([UPLOADING, UPLOAD_OK, UPLOAD_ERROR], TOTAL, 'Upload', &:upload)

              context.process_with_notification([ACKSENDING, ACKSEND_OK, ACKSEND_ERROR], TOTAL, 'Acknowledge') do
                msg.acknowledge!
              end
            end
          end
          context.notify(CLEANUP, TOTAL, 'Cleanup')
        end

        def build_command(context)
          msg_wrapper = MessageWrapper.new(context)
          r = ExpandVariable.expand_variables(cmd_template, msg_wrapper)
          if commands = Proxy.config[:commands]
            if template = commands[r]
              msg_wrapper = MessageWrapper.new(context)
              return ExpandVariable.expand_variables(template, msg_wrapper)
            else
              raise BuildError, "Invalid command key #{r.inspect} was given"
            end
          else
            return r
          end
        end
      end
    end
  end
end
