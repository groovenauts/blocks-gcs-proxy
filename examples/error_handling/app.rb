#! /usr/bin/env ruby

puts ARGV.inspect

code = ARGV.first.to_i
puts "Now exiting with code: #{code}"

exit code
