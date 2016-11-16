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

        attr_reader :base_cmd
        def initialize(*args)
          @base_cmd = args.join(' ')
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
            gcs = paese(msg.attributes['gcs'])

            download(gcs['download_files']) if gcs

            cmd = base_cmd.dup
            cmd << ' ' << msg.data unless msg.data.nil?

            logger.info("Executing command: #{cmd.inspect}")

            if system(cmd)
              upload(gcs['upload_files']) if gcs

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
      end
    end
  end
end
