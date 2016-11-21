# coding: utf-8
require "magellan/gcs/proxy"
require "magellan/gcs/proxy/log"

require 'fileutils'
require 'uri'

module Magellan
  module Gcs
    module Proxy
      class Context
        include Log

        attr_reader :workspace, :remote_download_files
        def initialize(workspace, remote_download_files)
          @workspace = workspace
          @remote_download_files = remote_download_files
        end

        KEYS = [
          :workspace,
          :downloads_dir, :uploads_dir,
          :download_files,
          :local_download_files,
          :remote_download_files,
        ].freeze

        def [](key)
          case key.to_sym
          when *KEYS then send(key)
          else nil
          end
        end

        def include?(key)
          KEYS.include?(key)
        end

        def downloads_dir
          File.join(workspace, 'downloads')
        end

        def download_mapping
          @download_mapping ||= build_mapping(downloads_dir, remote_download_files)
        end

        def local_download_files
          @local_download_files ||= build_local_files_obj(remote_download_files, download_mapping)
        end
        alias_method :download_files, :local_download_files

        def uploads_dir
          File.join(workspace, 'uploads')
        end

        def setup
          setup_dirs
        end

        def download
          download_mapping.each do |url, path|
            FileUtils.mkdir_p File.dirname(path)
            logger.debug("Downloading: #{url} to #{path}")
            uri = parse_uri(url)
            bucket = GCP.storage.bucket(uri.host)
            file = bucket.file uri.path.sub(/\A\//, '')
            file.download(path)
            logger.info("Download OK: #{url} to #{path}")
          end
        end

        def upload
          upload_mapping.each do |url, path|
            logger.info("Uploading: #{path} to #{url}")
            uri = parse_uri(url)
            bucket = GCP.storage.bucket(uri.host)
            bucket.create_file path, uri.path.sub(/\A\//, '')
            logger.info("Upload OK: #{path} to #{url}")
          end
        end

        def setup_dirs
          [:downloads_dir, :uploads_dir].each{|k| Dir.mkdir(send(k))}
        end

        def build_mapping(base_dir, obj)
          flatten_values(obj).flatten.each_with_object({}) do |url, d|
            uri = parse_uri(url)
            d[url] = File.join(base_dir, uri.path)
          end
        end

        def flatten_values(obj)
          case obj
          when nil then []
          when Hash then flatten_values(obj.values)
          when Array then obj.map{|i| flatten_values(i) }
          else obj
          end
        end

        def parse_uri(str)
          uri = URI.parse(str)
          raise "Unsupported scheme #{uri.scheme.inspect} of #{str}" unless uri.scheme == 'gs'
          uri
        end

        def build_local_files_obj(obj, mapping)
          case obj
          when Hash then obj.each_with_object({}){|(k,v), d| d[k] = build_local_files_obj(v, mapping)}
          when Array then obj.map{|i| build_local_files_obj(i, mapping)}
          when String then mapping[obj]
          else obj
          end
        end

      end
    end
  end
end
