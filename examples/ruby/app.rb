#! /usr/bin/env ruby

puts "$PROGRAM_NAME: #{$PROGRAM_NAME}"
puts "ARGV: #{ARGV.inspect}"

require 'pathname'
require 'fileutils'

path = Pathname.new(ARGV[0])
download_dir = Pathname.new(ARGV[1])
upload_dir = ARGV[2]
suffix = ARGV[3]
relpath = path.relative_path_from(download_dir)

dest_path = File.join(upload_dir, relpath.to_s)
dest_path.sub!(/\.([^\s]*)\z/, "-#{suffix}.\\1")

FileUtils::Verbose.mkdir_p(File.dirname(dest_path))
FileUtils::Verbose.copy_entry(path, dest_path)
