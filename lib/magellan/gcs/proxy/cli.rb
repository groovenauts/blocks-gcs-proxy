require "magellan/gcs/proxy"
require "magellan/gcs/proxy/log"

require 'json'
require 'logger'
require 'tmpdir'

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
          logger.info("Start listening")
          GCP.subscription.listen do |msg|
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
            logger.info("dfiles: #{dfiles}")
            ufiles =  parse(msg.attributes['upload_files'])
            logger.info("ufiles: #{ufiles}")

            context = Context.new(dir, dfiles, ufiles)
            context.setup
            logger.info("context.setup done.")

            context.download
            logger.info("context.download done.")
            logger.info("msg: #{msg}")
            logger.info("context: #{context}")

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
