require "magellan/gcs/proxy"

require 'json'
require 'logger'
require 'tmpdir'

module Magellan
  module Gcs
    module Proxy
      class Command
        include FileOperation
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
            download(flatten_values(paese(msg.attributes['download_files'])))

            cmd = build_command(msg, dir)

            logger.info("Executing command: #{cmd.inspect}")

            if system(cmd)
              upload(flatten_values(paese(msg.attributes['upload_files'])))

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

        def flatten_values(obj)
          case obj
          when Hash then flatten_values(obj.values)
          when Array then obj.map{|i| flatten_values(i) }
          else obj
          end
        end

        def build_command(msg)
          msg_wrapper = MessageWrapper.new(msg)
          ExpandVariable.expand_variables(cmd_template, msg_wrapper)
        end
      end
    end
  end
end
