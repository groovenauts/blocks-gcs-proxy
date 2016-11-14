require "magellan/gcs/proxy"

require 'json'

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
          sub.listen do |msg|
            p msg

            gcs = msg.attributes['gcs']
            gcs = JSON.parse(gcs) if gcs

            download(gcs['download_files']) if gcs

            cmd = base_cmd.dup
            cmd << ' ' << msg.data unless msg.data.nil?
            p cmd

            if system(cmd)
              download(gcs['upload_files']) if gcs

              sub.acknowledge msg
              puts "acknowledge!"
            else
              puts "Error: #{cmd.inspect}"
            end

            cleanup(gcs) if gcs
          end
        end

        def cleanup(gcs)
          deleted_files =
            gcs['download_files'].map{|obj| obj['dest']} +
            gcs['upload_files'  ].map{|obj| obj['src']}
          puts "Deleting..."
          p deleted_files
          deleted_files.each{|f| File.delete(f)}
        end

      end
    end
  end
end
