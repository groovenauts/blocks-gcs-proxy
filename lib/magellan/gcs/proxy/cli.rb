require "magellan/gcs/proxy"
require "magellan/gcs/proxy/pubsub_operation"

require 'json'
require 'logger'
require 'tmpdir'

module Magellan
  module Gcs
    module Proxy
      class Cli
        include PubsubOperation

        attr_reader :cmd_template
        def initialize(*args)
          @cmd_template = args.join(' ')
        end

        def run
          logger.info("Start listening")
          sub.listen do |msg|
            begin
              process(msg)
            rescue => e
              logger.error("[#{e.class.name}] #{e.message}")
            end
          end
        rescue => e
          logger.error("[#{e.class.name}] #{e.message}")
          raise e
        end

        def process(msg)
          logger.info("Processing message: #{msg.inspect}")
          Dir.mktmpdir 'workspace' do |dir|
            dfiles = parse(msg.attributes['download_files'])
            ufiles =  parse(msg.attributes['upload_files'])

            context = Context.new(dir, dfiles, ufiles)
            context.setup

            context.download

            cmd = build_command(msg, context)

            logger.info("Executing command: #{cmd.inspect}")

            if system(cmd)
              context.upload

              sub.acknowledge msg
              logger.info("Complete processing and acknowledged")
            else
              logger.error("Error: #{cmd.inspect}")
            end
          end
        end

        def logger
          @logger ||= Logger.new($stdout)
        end

        def parse(str)
          return nil if str.nil? || str.empty?
          JSON.parse(str)
        end


        def build_command(msg, context)
          msg_wrapper = MessageWrapper.new(msg, context)
          ExpandVariable.expand_variables(cmd_template, msg_wrapper)
        end
      end
    end
  end
end
