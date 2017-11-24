path = File.expand_path('../../../version.go', __FILE__)
GCS_PROXY_VERSION = File.read(path).scan(/\"(.+)\"/).flatten.first
