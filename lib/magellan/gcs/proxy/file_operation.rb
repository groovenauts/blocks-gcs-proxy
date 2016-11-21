require "magellan/gcs/proxy"

require 'uri'
require "google/cloud/storage"

module Magellan
  module Gcs
    module Proxy
      module FileOperation

        def download(base_dir, urls)
          (urls || []).each do |url|
            logger.info("Downloading: #{url}")
            uri = parse_uri(url)
            bucket = storage.bucket(uri.host)
            file = bucket.file uri.path.sub(/\A\//, '')
            file.download File.join(base_dir, uri.path)
          end
        end

        def upload(base_dir, urls)
          (urls || []).each do |url|
            logger.info("Uploading: #{url}")
            uri = parse_uri(url)
            bucket = storage.bucket(uri.host)
            bucket.create_file File.join(base_dir, uri.path), uri.path.sub(/\A\//, '')
          end
        end

        def parse_uri(str)
          uri = URI.parse(str)
          raise "Unsupported scheme #{uri.scheme.inspect} of #{str}" unless uri.scheme == 'gs'
          uri
        end

      end
    end
  end
end
